package plugin

import (
	"embed"

	"github.com/labstack/echo/v4"
	"github.com/vultisig/vultisigner/internal/types"
)

type Plugin interface {
	SignPluginMessages(c echo.Context) error
	SetupPluginPolicy(policyDoc *types.PluginPolicy) error
	ValidatePluginPolicy(policyDoc types.PluginPolicy) error
	ConfigurePlugin(c echo.Context) error
	Frontend() embed.FS

	ProposeTransactions(policy types.PluginPolicy) ([]types.PluginKeysignRequest, error)
	ValidateTransactionProposal(policy types.PluginPolicy, txs []types.PluginKeysignRequest) error
}
