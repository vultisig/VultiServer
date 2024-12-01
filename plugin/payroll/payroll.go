package payroll

import (
	"embed"

	"github.com/labstack/echo/v4"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/storage"
)

//go:embed frontend
var frontend embed.FS

type PayrollPlugin struct {
	db storage.DatabaseStorage
}

func NewPayrollPlugin(db storage.DatabaseStorage) *PayrollPlugin {
	return &PayrollPlugin{
		db: db,
	}
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
