package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/vultisig/mobile-tss-lib/tss"

	"github.com/vultisig/vultisigner/common"
	"github.com/vultisig/vultisigner/internal/tasks"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/storage"
)

type Server struct {
	port          int64
	redis         *storage.RedisStorage
	client        *asynq.Client
	inspector     *asynq.Inspector
	vaultFilePath string
}

// NewServer returns a new server.
func NewServer(port int64,
	redis *storage.RedisStorage,
	client *asynq.Client,
	inspector *asynq.Inspector,
	vaultFilePath string) *Server {
	return &Server{
		port:          port,
		redis:         redis,
		client:        client,
		inspector:     inspector,
		vaultFilePath: vaultFilePath,
	}
}

func (s *Server) StartServer() error {
	e := echo.New()
	e.Logger.SetLevel(log.DEBUG)
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit("2M")) // set maximum allowed size for a request body to 2M

	e.Use(middleware.CORS())
	e.GET("/ping", s.Ping)

	//serve demo/generated/img folder as img
	e.Static("/img", "./demo/generated/img")
	//serve demo/generated/static folder as static
	e.Static("/static", "./demo/generated/static")
	e.GET("/demo", func(c echo.Context) error {
		//server index.html file in demo folder
		return c.File("./demo/generated/index.html")
	})

	e.GET("/getDerivedPublicKey", s.GetDerivedPublicKey)
	grp := e.Group("/vault")

	grp.POST("/create", s.CreateVault)
	grp.POST("/upload", s.UploadVault)
	grp.GET("/download/:publicKeyECDSA", s.DownloadVault)
	grp.GET("/get/:publicKeyECDSA", s.GetVault)           // Get Vault Data
	grp.POST("/sign", s.SignMessages)                     // Sign messages
	grp.GET("/sign/response/:taskId", s.GetKeysignResult) // Get keysign result
	return e.Start(fmt.Sprintf(":%d", s.port))
}

func (s *Server) Ping(c echo.Context) error {
	return c.String(http.StatusOK, "Vultisigner is running")
}

func (s *Server) getHexEncodedRandomBytes() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("fail to generate random bytes, err: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func (s *Server) GetDerivedPublicKey(c echo.Context) error {
	publicKey := c.QueryParam("publicKey")
	if publicKey == "" {
		return fmt.Errorf("publicKey is required")
	}
	hexChainCode := c.QueryParam("hexChainCode")
	if hexChainCode == "" {
		return fmt.Errorf("hexChainCode is required")
	}
	derivePath := c.QueryParam("derivePath")
	if derivePath == "" {
		return fmt.Errorf("derivePath is required")
	}
	isEdDSA := false
	isEdDSAstr := c.QueryParam("isEdDSA")
	if isEdDSAstr == "true" {
		isEdDSA = true
	}

	derivedPublicKey, err := tss.GetDerivedPubKey(publicKey, hexChainCode, derivePath, isEdDSA)
	if err != nil {
		return fmt.Errorf("fail to get derived public key from tss, err: %w", err)
	}

	return c.JSON(http.StatusOK, derivedPublicKey)
}

func (s *Server) CreateVault(c echo.Context) error {
	var req types.VaultCreateRequest
	if err := c.Bind(&req); err != nil {
		return fmt.Errorf("fail to parse request, err: %w", err)
	}
	if err := req.IsValid(); err != nil {
		return fmt.Errorf("invalid request, err: %w", err)
	}
	buf, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("fail to marshal to json, err: %w", err)
	}
	_, err = s.client.Enqueue(asynq.NewTask(tasks.TypeKeyGeneration, buf),
		asynq.MaxRetry(-1),
		asynq.Timeout(7*time.Minute),
		asynq.Retention(10*time.Minute),
		asynq.Queue(tasks.QUEUE_NAME))
	if err != nil {
		return fmt.Errorf("fail to enqueue task, err: %w", err)
	}
	return c.NoContent(http.StatusOK)
}

// UploadVault is a handler that receives a vault file from integration.
func (s *Server) UploadVault(c echo.Context) error {
	bodyReader := http.MaxBytesReader(c.Response(), c.Request().Body, 2<<20) // 2M
	content, err := io.ReadAll(bodyReader)
	if err != nil {
		return fmt.Errorf("fail to read body, err: %w", err)
	}

	passwd := c.Request().Header.Get("x-password")
	if passwd == "" {
		return fmt.Errorf("vault backup password is required")
	}

	vault, err := common.DecryptVaultFromBackup(passwd, content)
	if err != nil {
		return fmt.Errorf("fail to decrypt vault from the backup, err: %w", err)
	}

	filePathName := filepath.Join(s.vaultFilePath, vault.PublicKeyEcdsa+".bak")
	file, err := os.Create(filePathName)
	if err != nil {
		return fmt.Errorf("fail to create file, err: %w", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			c.Logger().Errorf("fail to close file, err: %v", err)
		}
	}()

	if _, err := file.Write(content); err != nil {
		return fmt.Errorf("fail to write file, err: %w", err)
	}

	return c.NoContent(http.StatusOK)
}

func (s *Server) DownloadVault(c echo.Context) error {
	publicKeyECDSA := c.Param("publicKeyECDSA")
	if publicKeyECDSA == "" {
		return fmt.Errorf("public key is required")
	}

	filePathName := filepath.Join(s.vaultFilePath, publicKeyECDSA+".bak")
	_, err := os.Stat(filePathName)
	if err != nil {
		return fmt.Errorf("fail to get file info, err: %w", err)
	}

	passwd := c.Request().Header.Get("x-password")
	if passwd == "" {
		return fmt.Errorf("vault backup password is required")
	}

	content, err := os.ReadFile(filePathName)
	if err != nil {
		return fmt.Errorf("fail to read file, err: %w", err)
	}

	_, err = common.DecryptVaultFromBackup(passwd, content)
	if err != nil {
		return fmt.Errorf("fail to decrypt vault from the backup, err: %w", err)
	}

	// when we get to this point, the vault file is valid and can be decoded by the client , so pass it to them
	return c.File(filePathName)
}

func (s *Server) GetVault(c echo.Context) error {
	publicKeyECDSA := c.Param("publicKeyECDSA")
	if publicKeyECDSA == "" {
		return fmt.Errorf("public key is required")
	}

	filePathName := filepath.Join(s.vaultFilePath, publicKeyECDSA+".bak")
	_, err := os.Stat(filePathName)
	if err != nil {
		return fmt.Errorf("fail to get file info, err: %w", err)
	}

	passwd := c.Request().Header.Get("x-password")
	if passwd == "" {
		return fmt.Errorf("vault backup password is required")
	}

	content, err := os.ReadFile(filePathName)
	if err != nil {
		return fmt.Errorf("fail to read file, err: %w", err)
	}

	vault, err := common.DecryptVaultFromBackup(passwd, content)
	if err != nil {
		return fmt.Errorf("fail to decrypt vault from the backup, err: %w", err)
	}

	return c.JSON(http.StatusOK, types.VaultGetResponse{
		Name:           vault.Name,
		PublicKeyEcdsa: vault.PublicKeyEcdsa,
		PublicKeyEddsa: vault.PublicKeyEddsa,
		HexChainCode:   vault.HexChainCode,
		LocalPartyId:   vault.LocalPartyId,
	})
}

// SignMessages is a handler to process Keysing request
func (s *Server) SignMessages(c echo.Context) error {
	var req types.KeysignRequest
	if err := c.Bind(&req); err != nil {
		return fmt.Errorf("fail to parse request, err: %w", err)
	}
	if err := req.IsValid(); err != nil {
		return fmt.Errorf("invalid request, err: %w", err)
	}

	filePathName := filepath.Join(s.vaultFilePath, req.PublicKey+".bak")
	_, err := os.Stat(filePathName)
	if err != nil {
		return fmt.Errorf("fail to get file info, err: %w", err)
	}

	content, err := os.ReadFile(filePathName)
	if err != nil {
		return fmt.Errorf("fail to read file, err: %w", err)
	}

	_, err = common.DecryptVaultFromBackup(req.VaultPassword, content)
	if err != nil {
		return fmt.Errorf("fail to decrypt vault from the backup, err: %w", err)
	}
	buf, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("fail to marshal to json, err: %w", err)
	}
	ti, err := s.client.EnqueueContext(c.Request().Context(),
		asynq.NewTask(tasks.TypeKeySign, buf),
		asynq.MaxRetry(-1),
		asynq.Timeout(2*time.Minute),
		asynq.Retention(5*time.Minute),
		asynq.Queue(tasks.QUEUE_NAME))

	if err != nil {
		return fmt.Errorf("fail to enqueue task, err: %w", err)
	}

	return c.JSON(http.StatusOK, ti.ID)

}

// GetKeysignResult is a handler to get the keysign response
func (s *Server) GetKeysignResult(c echo.Context) error {
	taskID := c.Param("taskId")
	if taskID == "" {
		return fmt.Errorf("task id is required")
	}
	task, err := s.inspector.GetTaskInfo(tasks.QUEUE_NAME, taskID)
	if err != nil {
		return fmt.Errorf("fail to find task, err: %w", err)
	}

	if task == nil {
		return fmt.Errorf("task not found")
	}

	if task.State == asynq.TaskStatePending {
		return c.JSON(http.StatusOK, "Task is still in progress")
	}

	if task.State == asynq.TaskStateCompleted {
		return c.JSON(http.StatusOK, task.Result)
	}

	return fmt.Errorf("task state is invalid")
}
