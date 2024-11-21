package plugin

import (
	"github.com/labstack/echo/v4"
	"github.com/vultisig/vultisigner/internal/types"
)

type Plugin interface {
	SignPluginMessages(c echo.Context) error
	ValidatePluginPolicy(policyDoc types.PluginPolicy) error
	ConfigurePlugin(c echo.Context) error
}
