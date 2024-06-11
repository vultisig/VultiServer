package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"github.com/vultisig/vultisigner/internal/models"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/storage"
)

type Server struct {
	port          int64
	s             *storage.RedisStorage
	client        *asynq.Client
	vaultFilePath string
}

// NewServer returns a new server.
func NewServer(port int64,
	s *storage.RedisStorage,
	client *asynq.Client,
	vaultFilePath string) *Server {
	return &Server{
		port:          port,
		s:             s,
		client:        client,
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
	e.GET("/ping", s.Ping)
	grp := e.Group("/vault")
	grp.POST("/create", s.CreateVault)
	grp.POST("/upload", s.UploadVault)
	grp.GET("/download/{publicKeyECDSA}", s.DownloadVault)
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
	if err := s.s.SetVaultCacheItem(c.Request().Context(), &cacheItem); err != nil {
		return fmt.Errorf("fail to set vault cache item, err: %w", err)
	}

	resp := types.VaultCreateResponse{
		Name:             req.Name,
		SessionID:        sessionID,
		HexEncryptionKey: encryptionKey,
		HexChainCode:     hexChainCode,
	}
	task, err := resp.Task()
	if err != nil {
		return fmt.Errorf("fail to create task, err: %w", err)
	}
	_, err = s.client.Enqueue(task, asynq.MaxRetry(1), asynq.Timeout(1*time.Minute), asynq.Unique(time.Hour))
	if err != nil {
		return fmt.Errorf("fail to enqueue task, err: %w", err)
	}

	return c.JSON(http.StatusOK, resp)
}

// UploadVault is a handler that receives a vault file from integration.
func (s *Server) UploadVault(c echo.Context) error {
	var vaultBackup models.VaultBackup
	if err := c.Bind(&vaultBackup); err != nil {
		return fmt.Errorf("fail to parse request, err: %w", err)
	}
	filePathName := filepath.Join(s.vaultFilePath, vaultBackup.Vault.PubKeyECDSA+".dat")
	file, err := os.Create(filePathName)
	if err != nil {
		return fmt.Errorf("fail to create file, err: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			c.Logger().Errorf("fail to close file, err: %v", err)
		}
	}()
	buf, err := json.Marshal(vaultBackup)
	if err != nil {
		return fmt.Errorf("fail to serialize vault backup, err: %w", err)
	}
	if _, err := file.Write(buf); err != nil {
		return fmt.Errorf("fail to write file, err: %w", err)
	}
	return c.NoContent(http.StatusOK)
}

func (s *Server) DownloadVault(c echo.Context) error {
	publicKeyECDSA := c.Param("publicKeyECDSA")
	if publicKeyECDSA == "" {
		return fmt.Errorf("public key is required")
	}
	filePathName := filepath.Join(s.vaultFilePath, publicKeyECDSA+".dat")
	_, err := os.Stat(filePathName)
	if err != nil {
		return fmt.Errorf("fail to get file info, err: %w", err)
	}
	return c.File(filePathName)
}
