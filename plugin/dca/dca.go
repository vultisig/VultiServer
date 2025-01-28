package dca

import (
	"embed"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	gcommon "github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/storage"
)

const (
	PLUGIN_TYPE    = "dca"
	PLUGIN_VERSION = "0.0.1"
	POLICY_VERSION = "0.0.1"
)

type DCAPlugin struct {
	db storage.DatabaseStorage
}

func NewDCAPlugin(db storage.DatabaseStorage) *DCAPlugin {
	return &DCAPlugin{db}
}

func (p *DCAPlugin) SignPluginMessages(e echo.Context) error { return nil }

func (p *DCAPlugin) SetupPluginPolicy(policyDoc *types.PluginPolicy) error {
	if policyDoc == nil {
		return fmt.Errorf("no policy to set up")
	}

	if policyDoc.ID == "" {
		policyDoc.ID = uuid.NewString()
	}

	if policyDoc.PolicyVersion == "" {
		policyDoc.PolicyVersion = POLICY_VERSION
	}

	if policyDoc.PluginVersion == "" {
		policyDoc.PluginVersion = PLUGIN_VERSION
	}

	if policyDoc.PluginID == "" {
		policyDoc.PluginID = uuid.NewString()
	}

	return nil
}

func (p *DCAPlugin) ValidatePluginPolicy(policyDoc types.PluginPolicy) error {
	if policyDoc.PluginType != PLUGIN_TYPE {
		return fmt.Errorf("policy does not match plugin type, expected: %s, got: %s", PLUGIN_TYPE, policyDoc.PluginType)
	}

	var dcaPolicy types.DCAPolicy
	if err := json.Unmarshal(policyDoc.Policy, &dcaPolicy); err != nil {
		return fmt.Errorf("failed to unmarshal DCA policy: %w", err)
	}

	mixedCaseTokenIn, err := gcommon.NewMixedcaseAddressFromString(dcaPolicy.SourceTokenID)
	if err != nil {
		return fmt.Errorf("invalid source token address: %s", dcaPolicy.SourceTokenID)
	}
	if strings.ToLower(dcaPolicy.SourceTokenID) != dcaPolicy.SourceTokenID {
		if !mixedCaseTokenIn.ValidChecksum() {
			return fmt.Errorf("invalid source token address checksum: %s", dcaPolicy.SourceTokenID)
		}
	}

	mixedCaseTokenOut, err := gcommon.NewMixedcaseAddressFromString(dcaPolicy.DestinationTokenID)
	if err != nil {
		return fmt.Errorf("invalid destination token address: %s", dcaPolicy.DestinationTokenID)
	}
	if strings.ToLower(dcaPolicy.DestinationTokenID) != dcaPolicy.DestinationTokenID {
		if !mixedCaseTokenOut.ValidChecksum() {
			return fmt.Errorf("invalid destination token address checksum: %s", dcaPolicy.DestinationTokenID)
		}
	}

	if dcaPolicy.SourceTokenID == dcaPolicy.DestinationTokenID {
		return fmt.Errorf("source token and destination token addresses are the same")
	}

	if dcaPolicy.TotalAmount == "" {
		return fmt.Errorf("total amount is required")
	}
	totalAmount, ok := new(big.Int).SetString(dcaPolicy.TotalAmount, 10)
	if !ok {
		return fmt.Errorf("invalid total amount %s", dcaPolicy.TotalAmount)
	}
	if totalAmount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("total amount must be greater than 0")
	}

	if dcaPolicy.TotalOrders == "" {
		return fmt.Errorf("total orders is required")
	}
	totalOrders, ok := new(big.Int).SetString(dcaPolicy.TotalOrders, 10)
	if !ok {
		return fmt.Errorf("invalid total orders %s", dcaPolicy.TotalOrders)
	}
	if totalOrders.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("total orders must be greater than 0")
	}

	// if dcaPolicy.SlippagePercentage == "" {
	// 	return fmt.Errorf("slippage percentage is required")
	// }
	// slippage, err := strconv.ParseFloat(dcaPolicy.SlippagePercentage, 64)
	// if err != nil {
	// 	return fmt.Errorf("invalid slippage percentage %s", dcaPolicy.SlippagePercentage)
	// }
	// if slippage <= 0 || slippage > 100 {
	// 	return fmt.Errorf("slippage percentage must be between 0 and 100 %s", dcaPolicy.SlippagePercentage)
	// }

	if dcaPolicy.ChainID == "" {
		return fmt.Errorf("chain id is required")
	}

	return nil
}

func (p *DCAPlugin) ConfigurePlugin(e echo.Context) error { return nil }

func (p *DCAPlugin) ProposeTransactions(policy types.PluginPolicy) ([]types.PluginKeysignRequest, error) {
	return nil, nil
}

func (p *DCAPlugin) ValidateTransactionProposal(policy types.PluginPolicy, txs []types.PluginKeysignRequest) error {
	return nil
}

func (p *DCAPlugin) Frontend() embed.FS { return embed.FS{} }
