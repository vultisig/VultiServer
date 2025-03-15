package plugin

import (
	"context"
	"embed"

	"github.com/labstack/echo/v4"
	"github.com/vultisig/mobile-tss-lib/tss"
	"github.com/vultisig/vultisigner/internal/types"
)

type Plugin interface {
	ValidatePluginPolicy(policyDoc types.PluginPolicy) error

	ProposeTransactions(policy types.PluginPolicy) ([]types.PluginKeysignRequest, error)
	ValidateTransactionProposal(policy types.PluginPolicy, txs []types.PluginKeysignRequest) error

	SignPluginMessages(c echo.Context) error
	SigningComplete(ctx context.Context, signature tss.KeysignResponse, signRequest types.PluginKeysignRequest, policy types.PluginPolicy) error

	Frontend() embed.FS

	// TODO: remove
	GetNextNonce(address string) (uint64, error)
}
