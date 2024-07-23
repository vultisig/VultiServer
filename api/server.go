package api

import (
	"bytes"
	"compress/flate"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	keygenTypes "github.com/vultisig/commondata/go/vultisig/keygen/v1"
	"google.golang.org/protobuf/proto"

	"github.com/vultisig/vultisigner/common"
	"github.com/vultisig/vultisigner/config"
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
	grp := e.Group("/vault")
	grp.Use(middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
		Validator: s.AuthenticationValidator,
	}))
	grp.POST("/create", s.CreateVault)
	grp.POST("/upload", s.UploadVault)
	grp.GET("/download/{publicKeyECDSA}", s.DownloadVault)
	grp.POST("/sign", s.SignMessages)                       // Sign messages
	grp.GET("/sign/response/{task_id}", s.GetKeysignResult) // Get keysign result
	host := config.AppConfig.Server.Host
	return e.Start(fmt.Sprintf("%s:%d", host, s.port))
}

// AuthenticationValidator is a middleware that validates the basic auth credentials.
func (s *Server) AuthenticationValidator(username string, password string, c echo.Context) (bool, error) {
	if username == "" || password == "" {
		return false, nil
	}
	// save the user/password in redis, it is not idea , but vultisigner mostly run by integration partners
	// they can add db support later
	passwd, err := s.redis.GetUser(c.Request().Context(), username)
	if err != nil {
		return false, fmt.Errorf("fail to get user, err: %w", err)
	}
	if passwd == password {
		return true, nil
	}
	return false, nil
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

func (s *Server) CreateVault(c echo.Context) error {
	var req types.VaultCreateRequest
	if err := c.Bind(&req); err != nil {
		return fmt.Errorf("fail to parse request, err: %w", err)
	}
	sessionID := uuid.New().String()
	encryptionKey, err := s.getHexEncodedRandomBytes()
	if err != nil {
		return fmt.Errorf("fail to generate hex encryption key, err: %w", err)
	}

	hexChainCode, err := s.getHexEncodedRandomBytes()
	if err != nil {
		return fmt.Errorf("fail to generate hex chain code, err: %w", err)
	}

	cacheItem := types.VaultCacheItem{
		Name:               req.Name,
		SessionID:          sessionID,
		HexEncryptionKey:   encryptionKey,
		HexChainCode:       hexChainCode,
		EncryptionPassword: req.EncryptionPassword,
	}

	// Save the item into cache, so work can retrieve it
	if err := s.redis.SetVaultCacheItem(c.Request().Context(), &cacheItem); err != nil {
		return fmt.Errorf("fail to set vault cache item, err: %w", err)
	}

	keygenMsg := &keygenTypes.KeygenMessage{
		SessionId:        sessionID,
		HexChainCode:     hexChainCode,
		ServiceName:      "VultiSignerApp",
		EncryptionKeyHex: encryptionKey,
		UseVultisigRelay: true,
		VaultName:        req.Name,
	}

	serializedData, err := proto.Marshal(keygenMsg)
	if err != nil {
		return fmt.Errorf("fail to Marshal keygenMsg, err: %w", err)
	}

	var buf bytes.Buffer
	writer, err := flate.NewWriter(&buf, 5)
	if err != nil {
		return fmt.Errorf("flate.NewWriter failed, err: %w", err)
	}
	_, err = writer.Write(serializedData)
	if err != nil {
		return fmt.Errorf("writer.Write failed, err: %w", err)
	}
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("writer.Close failed, err: %w", err)
	}

	resp := types.VaultCreateResponse{
		Name:             req.Name,
		SessionID:        sessionID,
		HexEncryptionKey: encryptionKey,
		HexChainCode:     hexChainCode,
		KeygenMsg:        base64.StdEncoding.EncodeToString(buf.Bytes()),
	}
	task, err := cacheItem.Task()
	if err != nil {
		return fmt.Errorf("fail to create task, err: %w", err)
	}

	_, err = s.client.Enqueue(task,
		asynq.MaxRetry(-1),
		asynq.Timeout(7*time.Minute),
		asynq.Unique(time.Hour),
		asynq.Retention(10*time.Minute),
		asynq.Queue(tasks.QUEUE_NAME))
	if err != nil {
		return fmt.Errorf("fail to enqueue task, err: %w", err)
	}

	return c.JSON(http.StatusOK, resp)
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

// SignMessages is a handler to process Keysing request
func (s *Server) SignMessages(c echo.Context) error {
	var keysignReq types.KeysignRequest
	if err := c.Bind(&keysignReq); err != nil {
		return fmt.Errorf("fail to parse request, err: %w", err)
	}
	if err := keysignReq.IsValid(); err != nil {
		return fmt.Errorf("invalid request, err: %w", err)
	}

	filePathName := filepath.Join(s.vaultFilePath, keysignReq.PublicKeyECDSA+".bak")
	_, err := os.Stat(filePathName)
	if err != nil {
		return fmt.Errorf("fail to get file info, err: %w", err)
	}

	// password that used to decrypt the vault file
	// if the password can't be used to decrypt the vault file, the keysign request should be rejected
	passwd := c.Request().Header.Get("x-password")
	if passwd == "" {
		return fmt.Errorf("vault backup password is required")
	}
	// TODO: decrypt the vault file , if it failed to decrypt file , then reject the request

	task, err := keysignReq.NewKeysignTask(passwd)
	if err != nil {
		return fmt.Errorf("fail to create task, err: %w", err)
	}

	ti, err := s.client.EnqueueContext(c.Request().Context(), task, asynq.MaxRetry(-1),
		asynq.Timeout(2*time.Minute),
		asynq.Retention(5*time.Minute))

	if err != nil {
		return fmt.Errorf("fail to enqueue task, err: %w", err)
	}
	// return the task id to the client , so we can use the id to retrieve the task result
	return c.JSON(http.StatusOK, ti.ID)

}

// GetKeysignResult is a handler to get the keysign response
func (s *Server) GetKeysignResult(c echo.Context) error {
	taskID := c.QueryParam("task_id")
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
