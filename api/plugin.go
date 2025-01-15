package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/vultisig/vultisigner/plugin/dca"
	"net/http"
	"time"

	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	"github.com/vultisig/vultisigner/common"
	"github.com/vultisig/vultisigner/internal/tasks"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/plugin"
	"github.com/vultisig/vultisigner/plugin/payroll"
)

func (s *Server) SignPluginMessages(c echo.Context) error {
	s.logger.Info("Starting SignPluginMessages")
	var req types.PluginKeysignRequest
	if err := c.Bind(&req); err != nil {
		return fmt.Errorf("fail to parse request, err: %w", err)
	}

	// Plugin-specific validations
	if len(req.Messages) != 1 {
		return fmt.Errorf("plugin signing requires exactly one message hash, current: %d", len(req.Messages))
	}

	// Get policy from database
	policy, err := s.db.GetPluginPolicy(req.PolicyID)
	if err != nil {
		return fmt.Errorf("failed to get policy from database: %w", err)
	}

	// Validate policy matches plugin
	if policy.PluginID != req.PluginID {
		return fmt.Errorf("policy plugin ID mismatch")
	}

	// We re-init plugin as verification server doesn't have plugin defined
	var plugin plugin.Plugin
	switch policy.PluginType {
	case "payroll":
		plugin = payroll.NewPayrollPlugin(s.db)
	}

	if plugin == nil {
		err := fmt.Errorf("unknown plugin type: %s", policy.PluginType)
		s.logger.Error(err)
		return c.JSON(http.StatusBadRequest, err)
	}

	if err := plugin.ValidateTransactionProposal(policy, []types.PluginKeysignRequest{req}); err != nil {
		return fmt.Errorf("failed to validate transaction proposal: %w", err)
	}

	// Validate message hash matches transaction
	txHash, err := calculateTransactionHash(req.Transaction)
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
		wrappedErr := fmt.Errorf("fail to read file, err: %w", err)
		s.logger.Infof("fail to read file in SignPluginMessages, err: %v", err)
		s.logger.Error(wrappedErr)
		return wrappedErr
	}

	_, err = common.DecryptVaultFromBackup(req.VaultPassword, content)
	if err != nil {
		return fmt.Errorf("fail to decrypt vault from the backup, err: %w", err)
	}

	buf, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("fail to marshal to json, err: %w", err)
	}

	//Todo : check that tx is done only once per period

	// Create transaction with PENDING status first
	policyUUID, err := uuid.Parse(req.PolicyID)
	if err != nil {
		s.logger.Errorf("Failed to parse policy ID as UUID: %v", err)
		return fmt.Errorf("invalid policy ID format: %w", err)
	}

	metadata := map[string]interface{}{
		"timestamp":  time.Now().Format(time.RFC3339),
		"plugin_id":  req.PluginID,
		"public_key": req.PublicKey,
		"session_id": req.SessionID,
	}

	newTx := types.TransactionHistory{
		PolicyID: policyUUID,
		TxBody:   req.Transaction,
		Status:   types.StatusPending,
		Metadata: metadata,
	}

	txID, err := s.db.CreateTransactionHistory(newTx)
	if err != nil {
		s.logger.Errorf("Failed to create transaction history: %v", err)
		return fmt.Errorf("failed to create transaction record: %w", err)
	}

	ti, err := s.client.EnqueueContext(c.Request().Context(),
		asynq.NewTask(tasks.TypeKeySign, buf),
		asynq.MaxRetry(-1),
		asynq.Timeout(2*time.Minute),
		asynq.Retention(5*time.Minute),
		asynq.Queue(tasks.QUEUE_NAME))

	if err != nil {
		metadata["error"] = err.Error()
		if updateErr := s.db.UpdateTransactionStatus(txID, types.StatusSigningFailed, metadata); updateErr != nil {
			s.logger.Errorf("Failed to update transaction status: %v", updateErr)
		}
		return fmt.Errorf("fail to enqueue task, err: %w", err)
	}

	metadata["task_id"] = ti.ID
	if err := s.db.UpdateTransactionStatus(txID, types.StatusPending, metadata); err != nil {
		s.logger.Errorf("Failed to update transaction with task ID: %v", err)
	}

	s.logger.Infof("Created transaction history for tx from plugin: %s...",
		req.Transaction[:min(20, len(req.Transaction))],
	)

	return c.JSON(http.StatusOK, ti.ID)
}

func (s *Server) CreatePluginPolicy(c echo.Context) error {
	var policy types.PluginPolicy
	if err := c.Bind(&policy); err != nil {
		return fmt.Errorf("fail to parse request, err: %w", err)
	}

	// We re-init plugin as verification server doesn't have plugin defined
	var plugin plugin.Plugin
	switch policy.PluginType {
	case "payroll":
		plugin = payroll.NewPayrollPlugin(s.db)
	case "dca":
		plugin = dca.NewDCAPlugin(s.db)
	}

	if plugin == nil {
		err := fmt.Errorf("unknown plugin type: %s", policy.PluginType)
		s.logger.Error(err)
		return c.JSON(http.StatusBadRequest, err)
	}

	if err := plugin.ValidatePluginPolicy(policy); err != nil {
		err = fmt.Errorf("failed to validate policy: %w", err)
		s.logger.Error(err)
		return c.JSON(http.StatusBadRequest, err)
	}

	if err := s.db.InsertPluginPolicy(policy); err != nil {
		return fmt.Errorf("failed to insert policy: %w", err)
	}

	if s.scheduler != nil {
		if err := s.scheduler.CreateTimeTrigger(policy); err != nil {
			s.logger.Errorf("Failed to create time trigger: %v", err)
		}
	}

	return c.NoContent(http.StatusOK)
}

func (s *Server) ConfigurePlugin(c echo.Context) error {
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
