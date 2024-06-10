package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"github.com/vultisig/vultisigner/internal/types"
)

type Server struct {
	port int64
}

// NewServer returns a new server.
func NewServer(port int64) *Server {
	return &Server{
		port: port,
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

func (s *Server) CreateVault(c echo.Context) error {
	var req types.VaultCreateRequest
	if err := c.Bind(&req); err != nil {
		return fmt.Errorf("fail to parse request, err: %w", err)
	}
	// TODO: create vault
	return nil
}
