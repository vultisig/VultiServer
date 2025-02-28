package api

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/vultisig/vultisigner/common"
	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/internal/tasks"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/plugin"
	"github.com/vultisig/vultisigner/plugin/dca"
	"github.com/vultisig/vultisigner/plugin/payroll"
	"github.com/vultisig/vultisigner/uniswap"

	gcommon "github.com/ethereum/go-ethereum/common"
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
)

func (s *Server) SignPluginMessages(c echo.Context) error {
	s.logger.Debug("PLUGIN SERVER: SIGN MESSAGES")

	var req types.PluginKeysignRequest
	if err := c.Bind(&req); err != nil {
		return fmt.Errorf("fail to parse request, err: %w", err)
	}

	// Plugin-specific validations
	if len(req.Messages) != 1 {
		return fmt.Errorf("plugin signing requires exactly one message hash, current: %d", len(req.Messages))
	}

	// Get policy from database
	policy, err := s.db.GetPluginPolicy(c.Request().Context(), req.PolicyID)
	if err != nil {
		return fmt.Errorf("failed to get policy from database: %w", err)
	}

	// Validate policy matches plugin
	if policy.PluginID != req.PluginID {
		return fmt.Errorf("policy plugin ID mismatch")
	}

	// We re-init plugin as verification server doesn't have plugin defined
	var plg plugin.Plugin
	plg, err = s.initializePlugin(policy.PluginType)
	if err != nil {
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	if err := plg.ValidateTransactionProposal(policy, []types.PluginKeysignRequest{req}); err != nil {
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

	req.StartSession = false
	req.Parties = []string{common.PluginPartyID, common.VerifierPartyID}

	buf, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("fail to marshal to json, err: %w", err)
	}

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

	s.logger.Debug("PLUGIN SERVER: KEYSIGN TASK")

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
	if err := s.db.UpdateTransactionStatus(txID, types.StatusSigned, metadata); err != nil {
		s.logger.Errorf("Failed to update transaction with task ID: %v", err)
	}

	s.logger.Infof("Created transaction history for tx from plugin: %s...", req.Transaction[:min(20, len(req.Transaction))])

	return c.JSON(http.StatusOK, ti.ID)
}

func (s *Server) GetPluginPolicyById(c echo.Context) error {
	policyID := c.Param("policyId")
	if policyID == "" {
		err := fmt.Errorf("policy id is required")
		message := map[string]interface{}{
			"message": "failed to get policy",
			"error":   err.Error(),
		}
		s.logger.Error(err)

		return c.JSON(http.StatusBadRequest, message)

	}

	policy, err := s.policyService.GetPluginPolicy(c.Request().Context(), policyID)
	if err != nil {
		err = fmt.Errorf("failed to get policy: %w", err)
		message := map[string]interface{}{
			"message": fmt.Sprintf("failed to get policy: %s", policyID),
		}
		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, message)
	}

	return c.JSON(http.StatusOK, policy)
}

func (s *Server) GetAllPluginPolicies(c echo.Context) error {
	publicKey := c.Request().Header.Get("public_key")
	if publicKey == "" {
		err := fmt.Errorf("missing required header: public_key")
		message := map[string]interface{}{
			"message": "failed to get policies",
			"error":   err.Error(),
		}
		s.logger.Error(err)
		return c.JSON(http.StatusBadRequest, message)
	}

	pluginType := c.Request().Header.Get("plugin_type")
	if pluginType == "" {
		err := fmt.Errorf("missing required header: plugin_type")
		message := map[string]interface{}{
			"message": "failed to get policies",
			"error":   err.Error(),
		}
		s.logger.Error(err)
		return c.JSON(http.StatusBadRequest, message)
	}

	policies, err := s.policyService.GetPluginPolicies(c.Request().Context(), publicKey, pluginType)
	if err != nil {
		message := map[string]interface{}{
			"message": fmt.Sprintf("failed to get policies for public_key: %s", publicKey),
		}
		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, message)
	}

	return c.JSON(http.StatusOK, policies)
}

// TODO: verify the signature to authorize the operation
func (s *Server) CreatePluginPolicy(c echo.Context) error {
	var policy types.PluginPolicy
	if err := c.Bind(&policy); err != nil {
		return fmt.Errorf("fail to parse request, err: %w", err)
	}

	// We re-init plugin as verification server doesn't have plugin defined
	var plg plugin.Plugin
	plg, err := s.initializePlugin(policy.PluginType)
	if err != nil {
		err = fmt.Errorf("failed to initialize plugin: %w", err)
		s.logger.Error(err)

		message := map[string]interface{}{
			"message": fmt.Sprintf("failed to initialize plugin: %s", policy.PluginType),
		}

		return c.JSON(http.StatusBadRequest, message)
	}

	if err := plg.SetupPluginPolicy(&policy); err != nil {
		err = fmt.Errorf("failed to setup policy: %w", err)
		message := map[string]interface{}{
			"message": "failed to setup policy",
		}
		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, message)
	}

	if err := plg.ValidatePluginPolicy(policy); err != nil {
		if errors.Unwrap(err) != nil {
			err = fmt.Errorf("failed to validate policy: %w", err)
			s.logger.Error(err)
			message := map[string]interface{}{
				"message": "failed to validate policy",
			}
			return c.JSON(http.StatusBadRequest, message)
		}

		err = fmt.Errorf("failed to validate policy: %w", err)
		s.logger.Error(err)
		message := map[string]interface{}{
			"error":   err.Error(), // only if error is not wrapped
			"message": "failed to validate policy",
		}

		return c.JSON(http.StatusBadRequest, message)
	}

	newPolicy, err := s.policyService.CreatePolicyWithSync(c.Request().Context(), policy)
	if err != nil {
		err = fmt.Errorf("failed to create plugin policy: %w", err)
		message := map[string]interface{}{
			"message": "failed to create policy",
		}
		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, message)
	}

	return c.JSON(http.StatusOK, newPolicy)
}

// TODO: verify the signature to authorize the operation
func (s *Server) UpdatePluginPolicyById(c echo.Context) error {
	var policy types.PluginPolicy
	if err := c.Bind(&policy); err != nil {
		return fmt.Errorf("fail to parse request, err: %w", err)
	}

	// We re-init plugin as verification server doesn't have plugin defined
	var plg plugin.Plugin
	plg, err := s.initializePlugin(policy.PluginType)
	if err != nil {
		if errors.Unwrap(err) != nil {
			err = fmt.Errorf("failed to initialize plugin: %w", err)

			message := map[string]interface{}{
				"message": fmt.Sprintf("failed to initialize plugin: %s", policy.PluginType),
			}

			s.logger.Error(err)
			return c.JSON(http.StatusBadRequest, message)
		}

		err = fmt.Errorf("failed to initialize plugin: %w", err)

		message := map[string]interface{}{
			"message": fmt.Sprintf("failed to initialize plugin: %s", policy.PluginType),
			"error":   err.Error(),
		}

		s.logger.Error(err)
		return c.JSON(http.StatusBadRequest, message)
	}

	if err := plg.ValidatePluginPolicy(policy); err != nil {
		if errors.Unwrap(err) != nil {
			err = fmt.Errorf("failed to validate policy: %w", err)
			s.logger.Error(err)
			message := map[string]interface{}{
				"message": fmt.Sprintf("failed to validate policy: %s", policy.ID),
			}
			return c.JSON(http.StatusBadRequest, message)
		}

		err = fmt.Errorf("failed to validate policy: %w", err)
		s.logger.Error(err)
		message := map[string]interface{}{
			"error":   err.Error(), // only if error is not wrapped
			"message": fmt.Sprintf("failed to validate policy: %s", policy.ID),
		}
		return c.JSON(http.StatusBadRequest, message)
	}

	s.logger.Debug("Policy Signature", policy.Signature)

	updatedPolicy, err := s.policyService.UpdatePolicyWithSync(c.Request().Context(), policy)
	if err != nil {
		err = fmt.Errorf("failed to update plugin policy: %w", err)
		message := map[string]interface{}{
			"message": fmt.Sprintf("failed to update policy: %s", policy.ID),
		}
		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, message)
	}

	return c.JSON(http.StatusOK, updatedPolicy)
}

// TODO: verify the signature to authorize the operation
func (s *Server) DeletePluginPolicyById(c echo.Context) error {
	policyID := c.Param("policyId")
	if policyID == "" {
		err := fmt.Errorf("policy id is required")
		message := map[string]interface{}{
			"message": "failed to delete policy",
			"error":   err.Error(),
		}
		s.logger.Error(err)

		return c.JSON(http.StatusBadRequest, message)
	}

	if err := s.policyService.DeletePolicyWithSync(c.Request().Context(), policyID); err != nil {
		err = fmt.Errorf("failed to delte policy: %w", err)
		message := map[string]interface{}{
			"message": fmt.Sprintf("failed to delete policy: %s", policyID),
		}
		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, message)
	}

	return c.NoContent(http.StatusNoContent)
}

// TODO: do we actually need this?
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

	chainID := tx.ChainId()

	signer := gtypes.NewEIP155Signer(chainID)

	hash := signer.Hash(tx).String()[2:]
	return hash, nil
}

func (s *Server) initializePlugin(pluginType string) (plugin.Plugin, error) {
	switch pluginType {
	case "payroll":
		return payroll.NewPayrollPlugin(s.db), nil

	case "dca":
		cfg, err := config.ReadConfig("config-plugin")
		if err != nil {
			return nil, fmt.Errorf("failed to read plugin config: %w", err)
		}

		rpcClient, err := ethclient.Dial(cfg.Server.Plugin.Eth.Rpc)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize rpc client: %w", err)
		}

		uniswapV2RouterAddress := gcommon.HexToAddress(cfg.Server.Plugin.Eth.Uniswap.V2Router)
		uniswapCfg := uniswap.NewConfig(
			rpcClient,
			&uniswapV2RouterAddress,
			2000000, // TODO: config
			50000,   // TODO: config
			time.Duration(cfg.Server.Plugin.Eth.Uniswap.Deadline)*time.Minute,
		)

		return dca.NewDCAPlugin(uniswapCfg, s.db, s.logger)

	default:
		return nil, fmt.Errorf("unknown plugin type: %s", pluginType)
	}
}
