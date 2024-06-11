package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/storage"
)

type Server struct {
	port   int64
	s      *storage.RedisStorage
	client *asynq.Client
}

// NewServer returns a new server.
func NewServer(port int64, s *storage.RedisStorage, client *asynq.Client) *Server {
	return &Server{
		port:   port,
		s:      s,
		client: client,
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
