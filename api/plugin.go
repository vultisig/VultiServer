package api

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	gcommon "github.com/ethereum/go-ethereum/common"
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	"github.com/vultisig/vultisigner/common"
	"github.com/vultisig/vultisigner/internal/tasks"
	"github.com/vultisig/vultisigner/internal/types"
)

// ERC20 transfer method ABI
const erc20ABI = `[{
    "name": "transfer",
    "type": "function",
    "inputs": [
        {"name": "recipient", "type": "address"},
        {"name": "amount", "type": "uint256"}
    ],
    "outputs": [{"name": "", "type": "bool"}]
}]`

func (s *Server) SignPluginMessages(c echo.Context) error {
	var req types.PluginKeysignRequest
	if err := c.Bind(&req); err != nil {
		return fmt.Errorf("fail to parse request, err: %w", err)
	}

	// Plugin-specific validations
	if len(req.Messages) != 1 {
		return fmt.Errorf("plugin signing requires exactly one message hash")
	}
	if len(req.Transactions) != 1 {
		return fmt.Errorf("plugin signing requires exactly one transaction")
	}

	// Get policy
	policyPath := fmt.Sprintf("policies/%s.json", req.PolicyID)
	policyContent, err := s.blockStorage.GetFile(policyPath)
	if err != nil {
		return fmt.Errorf("fail to read policy file, err: %w", err)
	}

	var policy types.PluginPolicy
	if err := json.Unmarshal(policyContent, &policy); err != nil {
		return fmt.Errorf("fail to unmarshal policy, err: %w", err)
	}

	// Validate policy matches plugin
	if policy.PluginID != req.PluginID {
		return fmt.Errorf("policy plugin ID mismatch")
	}

	// Parse payroll policy
	var payrollPolicy types.PayrollPolicy
	if err := json.Unmarshal(policy.Policy, &payrollPolicy); err != nil {
		return fmt.Errorf("fail to unmarshal payroll policy, err: %w", err)
	}

	// Validate transaction matches policy
	if err := validateTransaction(req.Transactions[0], payrollPolicy); err != nil {
		return fmt.Errorf("transaction validation failed: %w", err)
	}

	// Validate message hash matches transaction
	txHash, err := calculateTransactionHash(req.Transactions[0])
	if err != nil {
		return fmt.Errorf("fail to calculate transaction hash: %w", err)
	}
	if txHash != req.Messages[0] {
		return fmt.Errorf("message hash does not match transaction hash. expected %s, got %s", txHash, req.Messages[0])
	}

	// Reuse existing signing logic
	result, err := s.redis.Get(c.Request().Context(), req.SessionID)
	if err == nil && result != "" {
		return c.NoContent(http.StatusOK)
	}

	if err := s.redis.Set(c.Request().Context(), req.SessionID, req.SessionID, 30*time.Minute); err != nil {
		s.logger.Errorf("fail to set session, err: %v", err)
	}

	filePathName := req.PublicKey + ".bak"
	content, err := s.blockStorage.GetFile(filePathName)
	if err != nil {
		return fmt.Errorf("fail to read file, err: %w", err)
	}

	_, err = common.DecryptVaultFromBackup(req.VaultPassword, content)
	if err != nil {
		return fmt.Errorf("fail to decrypt vault from the backup, err: %w", err)
	}

	buf, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("fail to marshal to json, err: %w", err)
	}

	ti, err := s.client.EnqueueContext(c.Request().Context(),
		asynq.NewTask(tasks.TypeKeySign, buf),
		asynq.MaxRetry(-1),
		asynq.Timeout(2*time.Minute),
		asynq.Retention(5*time.Minute),
		asynq.Queue(tasks.QUEUE_NAME))

	if err != nil {
		return fmt.Errorf("fail to enqueue task, err: %w", err)
	}

	return c.JSON(http.StatusOK, ti.ID)
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

// Helper functions that need to be implemented
func validateTransaction(txData string, policy types.PayrollPolicy) error {
	// Decode the raw transaction
	rawTx, err := hex.DecodeString(strings.TrimPrefix(txData, "0x"))
	if err != nil {
		return fmt.Errorf("failed to decode transaction data: %w", err)
	}

	tx := new(gtypes.Transaction)
	if err := tx.UnmarshalBinary(rawTx); err != nil {
		return fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	// Validate basic transaction properties
	if tx.To() == nil {
		return fmt.Errorf("transaction must have a recipient")
	}

	// Parse ERC20 ABI
	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return fmt.Errorf("failed to parse ERC20 ABI: %w", err)
	}

	// Get the first 4 bytes of the data (method ID)
	if len(tx.Data()) < 4 {
		return fmt.Errorf("transaction data too short")
	}
	methodID := tx.Data()[:4]

	// Check if this is a transfer method
	transferMethodID := parsedABI.Methods["transfer"].ID
	if !bytes.Equal(methodID, transferMethodID) {
		return fmt.Errorf("transaction is not an ERC20 transfer")
	}

	// Decode transfer parameters
	data := tx.Data()
	if len(data) != 68 { // 4 bytes method ID + 32 bytes address + 32 bytes amount
		return fmt.Errorf("invalid transfer data length")
	}

	// Extract recipient and amount from calldata
	recipientBytes := data[4:36]
	amountBytes := data[36:68]

	recipient := gcommon.BytesToAddress(recipientBytes[12:]) // last 20 bytes of the 32-byte field
	amount := new(big.Int).SetBytes(amountBytes)

	// Validate against policy
	if len(policy.Recipients) != 1 {
		return fmt.Errorf("policy must specify exactly one recipient")
	}

	policyRecipient := policy.Recipients[0]

	// Check recipient matches
	if !strings.EqualFold(recipient.Hex(), policyRecipient.Address) {
		return fmt.Errorf("recipient mismatch: expected %s, got %s",
			policyRecipient.Address, recipient.Hex())
	}

	// Check amount matches (convert policy amount from string to big.Int)
	policyAmount := new(big.Int)
	policyAmount.SetString(policyRecipient.Amount, 10)
	if amount.Cmp(policyAmount) != 0 {
		return fmt.Errorf("amount mismatch: expected %s, got %s",
			policyAmount.String(), amount.String())
	}

	// Check token contract matches
	if !strings.EqualFold(tx.To().Hex(), policy.TokenID) {
		return fmt.Errorf("token contract mismatch: expected %s, got %s",
			policy.TokenID, tx.To().Hex())
	}

	return nil
}

func calculateTransactionHash(txData string) (string, error) {
	tx := &gtypes.Transaction{}
	rawTx, err := hex.DecodeString(txData)
	if err != nil {
		return "", err
	}

	err = tx.UnmarshalBinary(rawTx)
	if err != nil {
		return "", err
	}

	hash := tx.Hash().Hex()[2:]
	return hash, nil
}
