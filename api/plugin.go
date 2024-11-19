package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/vultisig/vultisigner/internal/types"
)

func (s *Server) SignPluginMessages(c echo.Context) error {
	return nil
}

func (s *Server) CreatePluginPolicy(c echo.Context) error {
	var policy types.PluginPolicy
	if err := c.Bind(&policy); err != nil {
		return fmt.Errorf("fail to parse request, err: %w", err)
	}

	policyPath := fmt.Sprintf("policies/%s.json", policy.ID)
	content, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("fail to marshal policy, err: %w", err)
	}

	if err := s.blockStorage.UploadFile(content, policyPath); err != nil {
		return fmt.Errorf("fail to upload file, err: %w", err)
	}

	return c.NoContent(http.StatusOK)
}

func (s *Server) ConfigurePlugin(c echo.Context) error {
	return nil
}
