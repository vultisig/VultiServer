package payroll

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	gcommon "github.com/ethereum/go-ethereum/common"
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/mobile-tss-lib/tss"
	"github.com/vultisig/vultisigner/internal/types"
)

func (p *PayrollPlugin) ProposeTransactions(policy types.PluginPolicy) ([]types.PluginKeysignRequest, error) {
	var txs []types.PluginKeysignRequest
	err := p.ValidatePluginPolicy(policy)
	if err != nil {
		return txs, fmt.Errorf("failed to validate plugin policy: %v", err)
	}

	var payrollPolicy types.PayrollPolicy
	if err := json.Unmarshal(policy.Policy, &payrollPolicy); err != nil {
		return txs, fmt.Errorf("fail to unmarshal payroll policy, err: %w", err)
	}

	for i, recipient := range payrollPolicy.Recipients {
		txHash, rawTx, err := p.generatePayrollTransaction(
			recipient.Amount,
			recipient.Address,
			payrollPolicy.ChainID[i],
			payrollPolicy.TokenID[i],
			policy.PublicKey,
		)
		fmt.Printf("Chain ID TEST 1: %s\n", payrollPolicy.ChainID[i])
		if err != nil {
			return []types.PluginKeysignRequest{}, fmt.Errorf("failed to generate transaction hash: %v", err)
		}

		chainIDInt, _ := strconv.ParseInt(payrollPolicy.ChainID[i], 10, 64)

		// Create signing request
		signRequest := types.PluginKeysignRequest{
			KeysignRequest: types.KeysignRequest{
				PublicKey:        policy.PublicKey,
				Messages:         []string{txHash},
				SessionID:        uuid.New().String(),
				HexEncryptionKey: "0123456789abcdef0123456789abcdef",
				DerivePath:       "m/44/60/0/0/0",
				IsECDSA:          IsECDSA(chainIDInt),
				VaultPassword:    "your-secure-password",
			},
			Transaction: hex.EncodeToString(rawTx),
			PluginID:    policy.PluginID,
			PolicyID:    policy.ID,
		}
		txs = append(txs, signRequest)
	}

	signRequest := txs[0]
	txBytes, err := hex.DecodeString(signRequest.Transaction)
	if err != nil {
		p.logger.Errorf("Failed to decode transaction hex: %v", err)
		return []types.PluginKeysignRequest{}, fmt.Errorf("failed to decode transaction hex: %w", err)
	}
	//unmarshal tx from sign req.transaction
	tx := &gtypes.Transaction{}
	err = tx.UnmarshalBinary(txBytes)
	if err != nil {
		p.logger.Errorf("Failed to unmarshal transaction: %v", err)
		return []types.PluginKeysignRequest{}, fmt.Errorf("failed to unmarshal transaction: %d:", err)
	}
	fmt.Printf("Chain ID TEST 2: %s\n", tx.ChainId().String())
	fmt.Printf("len TEST 2: %d\n", len(txs))

	return txs, nil
}

func (p *PayrollPlugin) generatePayrollTransaction(amountString string, recipientString string, chainID string, tokenID string, publicKey string) (string, []byte, error) {
	amount := new(big.Int)
	amount.SetString(amountString, 10)
	recipient := gcommon.HexToAddress(recipientString)

	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse ABI: %v", err)
	}

	inputData, err := parsedABI.Pack("transfer", recipient, amount)
	if err != nil {
		return "", nil, fmt.Errorf("failed to pack transfer data: %v", err)
	}

	// create call message to estimate gas
	callMsg := ethereum.CallMsg{
		From:  recipient, //todo : this works, but maybe better to put the correct sender address once we have it
		To:    &recipient,
		Data:  inputData,
		Value: big.NewInt(0),
	}
	// estimate gas limit
	gasLimit, err := p.rpcClient.EstimateGas(context.Background(), callMsg)
	if err != nil {
		return "", nil, fmt.Errorf("failed to estimate gas: %v", err)
	}
	// add 20% to gas limit for safety
	gasLimit = gasLimit * 300 / 100
	// get suggested gas price
	gasPrice, err := p.rpcClient.SuggestGasPrice(context.Background())
	if err != nil {
		return "", nil, fmt.Errorf("failed to get gas price: %v", err)
	}
	gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(3))
	// Parse chain ID
	chainIDInt := new(big.Int)
	chainIDInt.SetString(chainID, 10)
	fmt.Printf("Chain ID TEST 3: %s\n", chainIDInt.String())

	derivedAddress, err := DeriveAddress(publicKey, "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", "m/44/60/0/0/0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to derive address: %v", err)
	}

	nextNonce, err := p.GetNextNonce(derivedAddress.Hex())
	if err != nil {
		return "", nil, fmt.Errorf("failed to get nonce: %v", err)
	}

	// Create unsigned transaction data
	txData := []interface{}{
		uint64(nextNonce),             // nonce
		gasPrice,                      // gas price
		uint64(gasLimit),              // gas limit
		gcommon.HexToAddress(tokenID), // to address
		big.NewInt(0),                 // value
		inputData,                     // data
		chainIDInt,                    // chain id
		uint(0),                       // empty v
		uint(0),                       // empty r
	}

	// Log each component separately
	p.logger.WithFields(logrus.Fields{
		"nonce":     txData[0],
		"gas_price": txData[1].(*big.Int).String(),
		"gas_limit": txData[2],
		"to":        txData[3].(gcommon.Address).Hex(),
		"value":     txData[4].(*big.Int).String(),
		"data_hex":  hex.EncodeToString(txData[5].([]byte)),
		"empty_v":   txData[6],
		"empty_r":   txData[7],
		"recipient": recipient.Hex(),
		"amount":    amount.String(),
	}).Info("Transaction components")

	rawTx, err := rlp.EncodeToBytes(txData)
	if err != nil {
		return "", nil, fmt.Errorf("failed to RLP encode transaction: %v", err)
	}

	txHash := crypto.Keccak256(rawTx)

	p.logger.WithFields(logrus.Fields{
		"raw_tx_hex":   hex.EncodeToString(rawTx),
		"hash_to_sign": hex.EncodeToString(txHash),
	}).Info("Final transaction data")

	/*txBytes, err := hex.DecodeString(string(rawTx))
	if err != nil {
		p.logger.Errorf("Failed to decode transaction hex: %v", err)
		return []types.PluginKeysignRequest{}, fmt.Errorf("failed to decode transaction hex: %w", err)
	}*/
	//unmarshal tx from sign req.transaction
	tx := &gtypes.Transaction{}
	err = rlp.DecodeBytes(rawTx, tx)
	if err != nil {
		p.logger.Errorf("Failed to RLP decode transaction: %v", err)
		return "", nil, fmt.Errorf("failed to RLP decode transaction: %v: %w", err, asynq.SkipRetry)
	}
	fmt.Printf("Chain ID TEST 4: %s\n", tx.ChainId().String())

	return hex.EncodeToString(txHash), rawTx, nil
}

func (p *PayrollPlugin) SigningComplete(ctx context.Context, signature tss.KeysignResponse, signRequest types.PluginKeysignRequest, policy types.PluginPolicy) error {
	R, S, V, originalTx, chainID, _, err := p.convertData(signature, signRequest, policy)
	if err != nil {
		return fmt.Errorf("failed to convert R and S: %v", err)
	}

	innerTx := &gtypes.LegacyTx{
		Nonce:    originalTx.Nonce(),
		GasPrice: originalTx.GasPrice(),
		Gas:      originalTx.Gas(),
		To:       originalTx.To(),
		Value:    originalTx.Value(),
		Data:     originalTx.Data(),
		V:        V,
		R:        R,
		S:        S,
	}

	signedTx := gtypes.NewTx(innerTx)
	signer := gtypes.NewLondonSigner(chainID)
	sender, err := signer.Sender(signedTx)
	if err != nil {
		p.logger.WithError(err).Warn("Could not determine sender")
	} else {
		p.logger.WithField("sender", sender.Hex()).Info("Transaction sender")
	}

	// Check if RPC client is initialized
	if p.rpcClient == nil {
		return fmt.Errorf("RPC client not initialized")
	}

	err = p.rpcClient.SendTransaction(ctx, signedTx)
	if err != nil {
		p.logger.WithError(err).Error("Failed to broadcast transaction")
		return p.handleBroadcastError(err, sender)
	}

	p.logger.WithField("hash", signedTx.Hash().Hex()).Info("Transaction successfully broadcast")

	return p.monitorTransaction(signedTx)
}

func (p *PayrollPlugin) convertData(signature tss.KeysignResponse, signRequest types.PluginKeysignRequest, policy types.PluginPolicy) (R *big.Int, S *big.Int, V *big.Int, originalTx *gtypes.Transaction, chainID *big.Int, recoveryID int64, err error) {
	// convert R and S from hex strings to big.Int
	R = new(big.Int)
	R.SetString(signature.R, 16)
	if R == nil {
		return nil, nil, nil, nil, nil, 0, fmt.Errorf("failed to parse R value")
	}

	S = new(big.Int)
	S.SetString(signature.S, 16)
	if S == nil {
		return nil, nil, nil, nil, nil, 0, fmt.Errorf("failed to parse S value")
	}

	// Decode the hex string to bytes first
	txBytes, err := hex.DecodeString(signRequest.Transaction)
	if err != nil {
		p.logger.Errorf("Failed to decode transaction hex: %v", err)
		return nil, nil, nil, nil, nil, 0, fmt.Errorf("failed to decode transaction hex: %w", err)
	}

	originalTx = new(gtypes.Transaction)
	if err := rlp.DecodeBytes(txBytes, originalTx); err != nil {
		p.logger.Errorf("Failed to unmarshal transaction: %v", err)
		return nil, nil, nil, nil, nil, 0, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	policybytes := policy.Policy
	payrollPolicy := types.PayrollPolicy{}
	err = json.Unmarshal(policybytes, &payrollPolicy)
	if err != nil {
		p.logger.Errorf("Failed to unmarshal policy: %v", err)
		return nil, nil, nil, nil, nil, 0, fmt.Errorf("failed to unmarshal policy: %w", err)
	}
	chainID = new(big.Int)
	chainID.SetString(payrollPolicy.ChainID[0], 10)

	/*chainID = originalTx.ChainId()
	fmt.Printf("Chain ID TEST: %s\n", chainID.String())*/

	// calculate V according to EIP-155
	recoveryID, err = strconv.ParseInt(signature.RecoveryID, 10, 64)
	if err != nil {
		return nil, nil, nil, nil, nil, 0, fmt.Errorf("failed to parse recovery ID: %w", err)
	}

	V = new(big.Int).Set(chainID)
	V.Mul(V, big.NewInt(2))
	V.Add(V, big.NewInt(35+recoveryID))

	return R, S, V, originalTx, chainID, recoveryID, nil
}
