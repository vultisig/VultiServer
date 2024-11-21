package payroll

import (
	"embed"

	"github.com/labstack/echo/v4"
	"github.com/vultisig/vultisigner/internal/types"
)

//go:embed frontend
var frontend embed.FS

type PayrollPlugin struct{}

func NewPayrollPlugin() *PayrollPlugin {
	return &PayrollPlugin{}
}

func (p *PayrollPlugin) SignPluginMessages(e echo.Context) error {
	return nil
}

func (p *PayrollPlugin) ValidatePluginPolicy(policyDoc types.PluginPolicy) error {
	return nil
}

func (p *PayrollPlugin) ConfigurePlugin(e echo.Context) error {
	return nil
}

func (p *PayrollPlugin) Frontend() embed.FS {
	return frontend
}
