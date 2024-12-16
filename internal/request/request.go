package request

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	gcommon "github.com/ethereum/go-ethereum/common"
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
	"github.com/vultisig/vultisigner/internal/types"
)

const erc20ABI = `[{
    "name": "transfer",
    "type": "function",
    "inputs": [
        {"name": "recipient", "type": "address"},
        {"name": "amount", "type": "uint256"}
    ],
    "outputs": [{"name": "", "type": "bool"}]
}]`

func CreateSigningRequest(policy types.PluginPolicy) ([]types.PluginKeysignRequest, error) {
	//check policy.pluginType.
	//depending on the pluginType, create the correct signing request
	if policy.PluginType == "payroll" {

		var payrollPolicy types.PayrollPolicy
		if err := json.Unmarshal(policy.Policy, &payrollPolicy); err != nil {
			return []types.PluginKeysignRequest{}, fmt.Errorf("fail to unmarshal payroll policy, err: %w", err)
		}

		signRequests := []types.PluginKeysignRequest{}

		for _, recipient := range payrollPolicy.Recipients {
			txHash, rawTx, err := GenerateTxHash(recipient.Amount, recipient.Address, payrollPolicy.ChainID, payrollPolicy.TokenID)
			if err != nil {
				return []types.PluginKeysignRequest{}, fmt.Errorf("failed to generate transaction hash: %v", err)
			}

			// Create signing request
			signRequest := types.PluginKeysignRequest{
				KeysignRequest: types.KeysignRequest{
					PublicKey:        "0200f9d07b02d182cd130afa088823f3c9dea027322dd834f5cffcb4b5e4a972e4",
					Messages:         []string{txHash}, //check how to correctly construct tx hash which depends on blockchain infos like nounce
					SessionID:        uuid.New().String(),
					HexEncryptionKey: "0123456789abcdef0123456789abcdef",
					DerivePath:       "m/44/60/0/0/0",
					IsECDSA:          true,
					VaultPassword:    "your-secure-password",
				},
				Transactions: []string{hex.EncodeToString(rawTx)},
				PluginID:     policy.PluginID,
				PolicyID:     policy.ID,
			}
			signRequests = append(signRequests, signRequest)
		}

		return signRequests, nil

		/*amount := new(big.Int)
		amount.SetString("1000000", 10) // 1 USDC
		recipient := gcommon.HexToAddress("0x742d35Cc6634C0532925a3b844Bc454e4438f44e")
		txHash, rawTx, err := GenerateTxHash(amount, recipient)
		if err != nil {
			return types.PluginKeysignRequest{}, fmt.Errorf("failed to generate transaction hash: %v", err)
		}

		// Create signing request
		signRequest := types.PluginKeysignRequest{
			KeysignRequest: types.KeysignRequest{
				PublicKey:        "0200f9d07b02d182cd130afa088823f3c9dea027322dd834f5cffcb4b5e4a972e4",
				Messages:         []string{txHash}, //check how to correctly construct tx hash which depends on blockchain infos like nounce
				SessionID:        uuid.New().String(),
				HexEncryptionKey: "0123456789abcdef0123456789abcdef",
				DerivePath:       "m/44/60/0/0/0",
				IsECDSA:          true,
				VaultPassword:    "your-secure-password",
			},
			Transactions: []string{hex.EncodeToString(rawTx)},
			PluginID:     policy.PluginID,
			PolicyID:     policy.ID,
		}

		fmt.Println("signRequest", signRequest)

		return signRequest, nil*/
	}

	return []types.PluginKeysignRequest{}, nil
}

func GenerateTxHash(amountString string, recipientString string, chainID string, tokenID string) (string, []byte, error) {

	amount := new(big.Int)
	amount.SetString(amountString, 10)
	recipient := gcommon.HexToAddress(recipientString)

	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse ABI: %v", err)
	}

	// Create transfer data
	inputData, err := parsedABI.Pack("transfer", recipient, amount)
	if err != nil {
		return "", nil, fmt.Errorf("failed to pack transfer data: %v", err)
	}

	// Create transaction
	tx := gtypes.NewTransaction(
		0,                             // nonce  //TODO : to be updated.
		gcommon.HexToAddress(tokenID), // USDC contract
		big.NewInt(0),                 // value, if it is not eth. If it is eth, we have to set the value. How to tell to send eth at plugin creation?
		100000,                        // gas limit
		big.NewInt(2000000000),        // gas price (2 gwei)
		inputData,
	)

	// Get the raw transaction bytes
	rawTx, err := tx.MarshalBinary()
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal transaction: %v", err)
	}

	// Calculate transaction hash
	txHash := tx.Hash().Hex()[2:]

	return txHash, rawTx, nil
}
