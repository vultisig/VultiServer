package payroll

import (
	"embed"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/vultisigner/plugin"
	"github.com/vultisig/vultisigner/storage"
)

//go:embed frontend
var frontend embed.FS

type PayrollPlugin struct {
	db           storage.DatabaseStorage
	nonceManager *plugin.NonceManager
	rpcClient    *ethclient.Client
	logger       logrus.FieldLogger
}

func NewPayrollPlugin(db storage.DatabaseStorage, logger logrus.FieldLogger, rpcClient *ethclient.Client) *PayrollPlugin {
	return &PayrollPlugin{
		db:           db,
		rpcClient:    rpcClient,
		nonceManager: plugin.NewNonceManager(rpcClient),
		logger:       logger,
	}
}

func (p *PayrollPlugin) FrontendSchema() embed.FS {
	return frontend
}

func (p *PayrollPlugin) GetNextNonce(address string) (uint64, error) {
	return p.nonceManager.GetNextNonce(address)
}
