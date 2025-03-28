package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	gcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/sirupsen/logrus"
	vaultType "github.com/vultisig/commondata/go/vultisig/vault/v1"
	"github.com/vultisig/mobile-tss-lib/tss"

	"github.com/vultisig/vultisigner/common"
	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/contexthelper"
	"github.com/vultisig/vultisigner/internal/syncer"
	"github.com/vultisig/vultisigner/internal/tasks"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/pkg/uniswap"
	"github.com/vultisig/vultisigner/plugin"
	"github.com/vultisig/vultisigner/plugin/dca"
	"github.com/vultisig/vultisigner/plugin/payroll"
	"github.com/vultisig/vultisigner/relay"
	"github.com/vultisig/vultisigner/storage"
	"github.com/vultisig/vultisigner/storage/postgres"
)

type WorkerService struct {
	cfg          config.Config
	verifierPort int64
	redis        *storage.RedisStorage
	logger       *logrus.Logger
	queueClient  *asynq.Client
	sdClient     *statsd.Client
	blockStorage *storage.BlockStorage
	inspector    *asynq.Inspector
	plugin       plugin.Plugin
	db           storage.DatabaseStorage
	rpcClient    *ethclient.Client
	syncer       syncer.PolicySyncer
	authService  *AuthService
}

// NewWorker creates a new worker service
func NewWorker(cfg config.Config, verifierPort int64, queueClient *asynq.Client, sdClient *statsd.Client, syncer syncer.PolicySyncer, authService *AuthService, blockStorage *storage.BlockStorage, inspector *asynq.Inspector) (*WorkerService, error) {
	logger := logrus.WithField("service", "worker").Logger

	redis, err := storage.NewRedisStorage(cfg)
	if err != nil {
		return nil, fmt.Errorf("storage.NewRedisStorage failed: %w", err)
	}

	db, err := postgres.NewPostgresBackend(false, cfg.Server.Database.DSN)
	if err != nil {
		return nil, fmt.Errorf("fail to connect to database: %w", err)
	}

	rpcClient, err := ethclient.Dial(cfg.Server.Plugin.Eth.Rpc)
	if err != nil {
		logrus.Fatalf("Failed to connect to RPC: %v", err)
		return nil, err
	}

	var plugin plugin.Plugin
	if cfg.Server.Mode == "plugin" {
		switch cfg.Server.Plugin.Type {
		case "payroll":
			plugin = payroll.NewPayrollPlugin(db, logrus.WithField("service", "plugin").Logger, rpcClient)
		case "dca":
			uniswapV2RouterAddress := gcommon.HexToAddress(cfg.Server.Plugin.Eth.Uniswap.V2Router)
			uniswapCfg := uniswap.NewConfig(
				rpcClient,
				&uniswapV2RouterAddress,
				2000000, // TODO: config
				50000,   // TODO: config
				time.Duration(cfg.Server.Plugin.Eth.Uniswap.Deadline)*time.Minute,
			)
			plugin, err = dca.NewDCAPlugin(uniswapCfg, db, logger)
			if err != nil {
				return nil, fmt.Errorf("fail to initialize DCA plugin: %w", err)
			}
		default:
			logger.Fatalf("Invalid plugin type: %s", cfg.Server.Plugin.Type)
		}
	}

	return &WorkerService{
		cfg:          cfg,
		db:           db,
		redis:        redis,
		blockStorage: blockStorage,
		rpcClient:    rpcClient,
		queueClient:  queueClient,
		sdClient:     sdClient,
		inspector:    inspector,
		plugin:       plugin,
		logger:       logger,
		syncer:       syncer,
		authService:  authService,
		verifierPort: verifierPort,
	}, nil
}

type KeyGenerationTaskResult struct {
	EDDSAPublicKey string
	ECDSAPublicKey string
}

func (s *WorkerService) incCounter(name string, tags []string) {
	if err := s.sdClient.Count(name, 1, tags, 1); err != nil {
		s.logger.Errorf("fail to count metric, err: %v", err)
	}
}

func (s *WorkerService) measureTime(name string, start time.Time, tags []string) {
	if err := s.sdClient.Timing(name, time.Since(start), tags, 1); err != nil {
		s.logger.Errorf("fail to measure time metric, err: %v", err)
	}
}

func (s *WorkerService) HandleKeyGeneration(ctx context.Context, t *asynq.Task) error {
	if err := contexthelper.CheckCancellation(ctx); err != nil {
		return err
	}
	defer s.measureTime("worker.vault.create.latency", time.Now(), []string{})
	var req types.VaultCreateRequest
	if err := json.Unmarshal(t.Payload(), &req); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	s.logger.WithFields(logrus.Fields{
		"name":           req.Name,
		"session":        req.SessionID,
		"local_party_id": req.LocalPartyId,
		"email":          req.Email,
	}).Info("Joining keygen")
	s.incCounter("worker.vault.create", []string{})
	if err := req.IsValid(); err != nil {
		return fmt.Errorf("invalid vault create request: %s: %w", err, asynq.SkipRetry)
	}
	keyECDSA, keyEDDSA, err := s.JoinKeyGeneration(req)
	if err != nil {
		_ = s.sdClient.Count("worker.vault.create.error", 1, nil, 1)
		s.logger.Errorf("keygen.JoinKeyGeneration failed: %v", err)
		return fmt.Errorf("keygen.JoinKeyGeneration failed: %v: %w", err, asynq.SkipRetry)
	}

	s.logger.WithFields(logrus.Fields{
		"keyECDSA": keyECDSA,
		"keyEDDSA": keyEDDSA,
	}).Info("localPartyID generation completed")

	result := KeyGenerationTaskResult{
		EDDSAPublicKey: keyEDDSA,
		ECDSAPublicKey: keyECDSA,
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		s.logger.Errorf("json.Marshal failed: %v", err)
		return fmt.Errorf("json.Marshal failed: %v: %w", err, asynq.SkipRetry)
	}

	if _, err := t.ResultWriter().Write(resultBytes); err != nil {
		s.logger.Errorf("t.ResultWriter.Write failed: %v", err)
		return fmt.Errorf("t.ResultWriter.Write failed: %v: %w", err, asynq.SkipRetry)
	}

	return nil
}

func (s *WorkerService) HandleKeySign(ctx context.Context, t *asynq.Task) error {
	if err := contexthelper.CheckCancellation(ctx); err != nil {
		s.logger.Error("Context cancelled")
		return err
	}
	var p types.KeysignRequest
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		s.logger.Errorf("json.Unmarshal failed: %v", err)
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	defer s.measureTime("worker.vault.sign.latency", time.Now(), []string{})
	s.incCounter("worker.vault.sign", []string{})
	s.logger.WithFields(logrus.Fields{
		"PublicKey":  p.PublicKey,
		"session":    p.SessionID,
		"Messages":   p.Messages,
		"DerivePath": p.DerivePath,
		"IsECDSA":    p.IsECDSA,
	}).Info("joining keysign")

	signatures, err := s.JoinKeySign(p)
	if err != nil {
		s.logger.Errorf("join keysign failed: %v", err)
		return fmt.Errorf("join keysign failed: %v: %w", err, asynq.SkipRetry)
	}

	s.logger.WithFields(logrus.Fields{
		"Signatures": signatures,
	}).Info("localPartyID sign completed")

	resultBytes, err := json.Marshal(signatures)
	if err != nil {
		s.logger.Errorf("json.Marshal failed: %v", err)
		return fmt.Errorf("json.Marshal failed: %v: %w", err, asynq.SkipRetry)
	}

	if _, err := t.ResultWriter().Write(resultBytes); err != nil {
		s.logger.Errorf("t.ResultWriter.Write failed: %v", err)
		return fmt.Errorf("t.ResultWriter.Write failed: %v: %w", err, asynq.SkipRetry)
	}

	return nil
}

func (s *WorkerService) HandleEmailVaultBackup(ctx context.Context, t *asynq.Task) error {
	if err := contexthelper.CheckCancellation(ctx); err != nil {
		return err
	}
	s.incCounter("worker.vault.backup.email", []string{})
	var req types.EmailRequest
	if err := json.Unmarshal(t.Payload(), &req); err != nil {
		s.logger.Errorf("json.Unmarshal failed: %v", err)
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	s.logger.WithFields(logrus.Fields{
		"email":    req.Email,
		"filename": req.FileName,
	}).Info("sending email")
	emailServer := "https://mandrillapp.com/api/1.0/messages/send-template"
	payload := MandrillPayload{
		Key:          s.cfg.EmailServer.ApiKey,
		TemplateName: "fastvault",
		TemplateContent: []MandrilMergeVarContent{
			{
				Name:    "VAULT_NAME",
				Content: req.VaultName,
			},
			{
				Name:    "VERIFICATION_CODE",
				Content: req.Code,
			},
		},
		Message: MandrillMessage{
			To: []MandrillTo{
				{
					Email: req.Email,
					Type:  "to",
				},
			},
			MergeVars: []MandrillVar{
				{
					Rcpt: req.Email,
					Vars: []MandrilMergeVarContent{
						{
							Name:    "VAULT_NAME",
							Content: req.VaultName,
						},
						{
							Name:    "VERIFICATION_CODE",
							Content: req.Code,
						},
					},
				},
			},
			SendingDomain: "vultisig.com",
			Attachments: []MandrillAttachment{
				{
					Type:    "application/octet-stream",
					Name:    req.FileName,
					Content: base64.StdEncoding.EncodeToString([]byte(req.FileContent)),
				},
			},
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.logger.Errorf("json.Marshal failed: %v", err)
		return fmt.Errorf("json.Marshal failed: %v: %w", err, asynq.SkipRetry)
	}
	resp, err := http.Post(emailServer, "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		s.logger.Errorf("http.Post failed: %v", err)
		return fmt.Errorf("http.Post failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			s.logger.Errorf("failed to close body: %v", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		s.logger.Errorf("http.Post failed: %s", resp.Status)
		return fmt.Errorf("http.Post failed: %s: %w", resp.Status, asynq.SkipRetry)
	}
	result, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Errorf("io.ReadAll failed: %v", err)
		return fmt.Errorf("io.ReadAll failed: %w", err)
	}
	s.logger.Info(string(result))
	if _, err := t.ResultWriter().Write([]byte("email sent")); err != nil {
		return fmt.Errorf("t.ResultWriter.Write failed: %v", err)
	}
	return nil
}

func (s *WorkerService) HandleReshare(ctx context.Context, t *asynq.Task) error {
	if err := contexthelper.CheckCancellation(ctx); err != nil {
		return err
	}
	var req types.ReshareRequest
	if err := json.Unmarshal(t.Payload(), &req); err != nil {
		s.logger.Errorf("json.Unmarshal failed: %v", err)
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	defer s.measureTime("worker.vault.reshare.latency", time.Now(), []string{})
	s.incCounter("worker.vault.reshare", []string{})
	s.logger.WithFields(logrus.Fields{
		"name":           req.Name,
		"session":        req.SessionID,
		"local_party_id": req.LocalPartyId,
		"email":          req.Email,
	}).Info("reshare request")
	if err := req.IsValid(); err != nil {
		return fmt.Errorf("invalid reshare request: %s: %w", err, asynq.SkipRetry)
	}
	localState, err := relay.NewLocalStateAccessorImp(req.LocalPartyId, s.cfg.Server.VaultsFilePath, req.PublicKey, req.EncryptionPassword, s.blockStorage)
	if err != nil {
		s.logger.Errorf("relay.NewLocalStateAccessorImp failed: %v", err)
		return fmt.Errorf("relay.NewLocalStateAccessorImp failed: %v: %w", err, asynq.SkipRetry)
	}
	var vault *vaultType.Vault
	if localState.Vault != nil {
		// reshare vault
		vault = localState.Vault

	} else {
		vault = &vaultType.Vault{
			Name:           req.Name,
			PublicKeyEcdsa: "",
			PublicKeyEddsa: "",
			HexChainCode:   req.HexChainCode,
			LocalPartyId:   req.LocalPartyId,
			Signers:        req.OldParties,
			ResharePrefix:  req.OldResharePrefix,
		}
		// create new vault
	}
	if err := s.Reshare(vault,
		req.SessionID,
		req.HexEncryptionKey,
		s.cfg.Relay.Server,
		req.EncryptionPassword,
		req.Email); err != nil {
		s.logger.Errorf("reshare failed: %v", err)
		return fmt.Errorf("reshare failed: %v: %w", err, asynq.SkipRetry)
	}

	return nil
}

func (s *WorkerService) HandlePluginTransaction(ctx context.Context, t *asynq.Task) error {
	if err := contexthelper.CheckCancellation(ctx); err != nil {
		return err
	}

	var triggerEvent types.PluginTriggerEvent
	if err := json.Unmarshal(t.Payload(), &triggerEvent); err != nil {
		s.logger.Errorf("json.Unmarshal failed: %v", err)
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	defer s.measureTime("worker.plugin.transaction.latency", time.Now(), []string{})

	// Always update back to PENDING status so the scheduler can enqueue task.
	defer func() {
		if err := s.db.UpdateTriggerStatus(ctx, triggerEvent.PolicyID, types.StatusTimeTriggerPending); err != nil {
			s.logger.Errorf("db.UpdateTriggerStatus failed: %v", err)
		}
		if err := s.db.UpdateTimeTriggerLastExecution(ctx, triggerEvent.PolicyID); err != nil {
			s.logger.Errorf("db.UpdateTimeTriggerLastExecution failed: %v", err)
		}
	}()

	s.incCounter("worker.plugin.transaction", []string{})
	s.logger.WithFields(logrus.Fields{
		"policy_id": triggerEvent.PolicyID,
	}).Info("plugin transaction request")

	policy, err := s.db.GetPluginPolicy(ctx, triggerEvent.PolicyID)
	if err != nil {
		s.logger.Errorf("db.GetPluginPolicy failed: %v", err)
		return fmt.Errorf("db.GetPluginPolicy failed: %v: %w", err, asynq.SkipRetry)
	}

	s.logger.WithFields(logrus.Fields{
		"policy_id":   policy.ID,
		"public_key":  policy.PublicKey,
		"plugin_type": policy.PluginType,
	}).Info("Retrieved policy for signing")

	// Propose transactions to sign
	signRequests, err := s.plugin.ProposeTransactions(policy)
	if err != nil {
		s.logger.Errorf("Failed to create signing request: %v", err)
		return fmt.Errorf("failed to create signing request: %v: %w", err, asynq.SkipRetry)
	}

	jwtToken, err := s.authService.GenerateToken()
	if err != nil {
		s.logger.Errorf("Failed to generate jwt token: %v", err)
	}

	for _, signRequest := range signRequests {
		policyUUID, err := uuid.Parse(signRequest.PolicyID)
		if err != nil {
			s.logger.Errorf("Failed to parse policy ID as UUID: %v", err)
			return err
		}

		// create transaction with PENDING status
		metadata := map[string]interface{}{
			"timestamp":        time.Now(),
			"plugin_id":        signRequest.PluginID,
			"public_key":       signRequest.KeysignRequest.PublicKey,
			"transaction_type": signRequest.TransactionType,
		}

		newTx := types.TransactionHistory{
			PolicyID: policyUUID,
			TxBody:   signRequest.Transaction,
			TxHash:   signRequest.Messages[0],
			Status:   types.StatusPending,
			Metadata: metadata,
		}

		if err := s.upsertAndSyncTransaction(ctx, syncer.CreateAction, &newTx, jwtToken); err != nil {
			return fmt.Errorf("upsertAndSyncTransaction failed: %w", err)
		}

		// start TSS signing process
		err = s.initiateTxSignWithVerifier(ctx, signRequest, metadata, newTx, jwtToken)
		if err != nil {
			return err
		}

		// prepare local sign request
		signRequest.KeysignRequest.StartSession = true
		signRequest.KeysignRequest.Parties = []string{common.PluginPartyID, common.VerifierPartyID}
		buf, err := json.Marshal(signRequest.KeysignRequest)
		if err != nil {
			s.logger.Errorf("Failed to marshal local sign request: %v", err)
			return err
		}

		// Enqueue TypeKeySign directly
		ti, err := s.queueClient.Enqueue(
			asynq.NewTask(tasks.TypeKeySign, buf),
			asynq.MaxRetry(0),
			asynq.Timeout(2*time.Minute),
			asynq.Retention(5*time.Minute),
			asynq.Queue(tasks.QUEUE_NAME),
		)
		if err != nil {
			s.logger.Errorf("Failed to enqueue signing task: %v", err)
			continue
		}

		s.logger.Infof("Enqueued signing task: %s", ti.ID)

		// wait for result with timeout
		result, err := s.waitForTaskResult(ti.ID, 120*time.Second) // adjust timeout as needed (each policy provider should be able to set it, but there should be an incentive to not retry too much)
		if err != nil {                                            //do we consider that the signature is always valid if err = nil?
			metadata["error"] = err.Error()
			metadata["task_id"] = ti.ID
			newTx.Status = types.StatusSigningFailed
			newTx.Metadata = metadata
			if err := s.upsertAndSyncTransaction(ctx, syncer.UpdateAction, &newTx, jwtToken); err != nil {
				s.logger.Errorf("upsertAndSyncTransaction failed: %v", err)
			}
			return err
		}

		// Update to SIGNED status with result
		metadata["task_id"] = ti.ID
		metadata["result"] = result
		newTx.Status = types.StatusSigned
		newTx.Metadata = metadata
		if err := s.upsertAndSyncTransaction(ctx, syncer.UpdateAction, &newTx, jwtToken); err != nil {
			return fmt.Errorf("upsertAndSyncTransaction failed: %v", err)
		}

		var signatures map[string]tss.KeysignResponse
		if err := json.Unmarshal(result, &signatures); err != nil {
			s.logger.Errorf("Failed to unmarshal signatures: %v", err)
			return fmt.Errorf("failed to unmarshal signatures: %w", err)
		}
		var signature tss.KeysignResponse
		for _, sig := range signatures {
			signature = sig
			break
		}

		err = s.plugin.SigningComplete(ctx, signature, signRequest, policy)
		if err != nil {
			s.logger.Errorf("Failed to complete signing: %v", err)

			newTx.Status = types.StatusRejected
			newTx.Metadata = metadata
			if err := s.upsertAndSyncTransaction(ctx, syncer.UpdateAction, &newTx, jwtToken); err != nil {
				s.logger.Errorf("upsertAndSyncTransaction failed: %v", err)
			}
			return fmt.Errorf("fail to complete signing: %w", err)
		}

		newTx.Status = types.StatusMined
		newTx.Metadata = metadata
		if err := s.upsertAndSyncTransaction(ctx, syncer.UpdateAction, &newTx, jwtToken); err != nil {
			s.logger.Errorf("upsertAndSyncTransaction failed: %v", err)
		}
	}

	return nil
}

func (s *WorkerService) initiateTxSignWithVerifier(ctx context.Context, signRequest types.PluginKeysignRequest, metadata map[string]interface{}, newTx types.TransactionHistory, jwtToken string) error {
	signBytes, err := json.Marshal(signRequest)
	if err != nil {
		s.logger.Errorf("Failed to marshal sign request: %v", err)
		return err
	}

	signResp, err := http.Post(
		fmt.Sprintf("http://localhost:%d/signFromPlugin", s.verifierPort),
		"application/json",
		bytes.NewBuffer(signBytes),
	)
	if err != nil {
		metadata["error"] = err.Error()
		newTx.Status = types.StatusSigningFailed
		newTx.Metadata = metadata
		if err = s.upsertAndSyncTransaction(ctx, syncer.UpdateAction, &newTx, jwtToken); err != nil {
			s.logger.Errorf("upsertAndSyncTransaction failed: %v", err)
		}
		return err
	}
	defer signResp.Body.Close()

	respBody, err := io.ReadAll(signResp.Body)
	if err != nil {
		s.logger.Errorf("Failed to read response: %v", err)
		return err
	}

	if signResp.StatusCode != http.StatusOK {
		metadata["error"] = string(respBody)
		newTx.Status = types.StatusSigningFailed
		newTx.Metadata = metadata
		if err := s.upsertAndSyncTransaction(ctx, syncer.UpdateAction, &newTx, jwtToken); err != nil {
			s.logger.Errorf("upsertAndSyncTransaction failed: %v", err)
		}
		return err
	}
	return nil
}

func (s *WorkerService) upsertAndSyncTransaction(ctx context.Context, action syncer.Action, tx *types.TransactionHistory, jwtToken string) error {
	s.logger.Info("upsertAndSyncTransaction started with action: ", action)
	dbTx, err := s.db.Pool().Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer dbTx.Rollback(ctx)

	if action == syncer.CreateAction {
		txID, err := s.db.CreateTransactionHistoryTx(ctx, dbTx, *tx)
		if err != nil {
			s.logger.Errorf("Failed to create (or update) transaction history tx: %v", err)
			return fmt.Errorf("failed to create transaction history: %w", err)
		}
		tx.ID = txID
	} else {
		if err = s.db.UpdateTransactionStatusTx(ctx, dbTx, tx.ID, tx.Status, tx.Metadata); err != nil {
			return fmt.Errorf("failed to update transaction status: %w", err)
		}
	}

	if err = s.syncer.SyncTransaction(action, jwtToken, *tx); err != nil {
		return fmt.Errorf("failed to sync transaction: %w", err)
	}

	if err = dbTx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (s *WorkerService) waitForTaskResult(taskID string, timeout time.Duration) ([]byte, error) {
	start := time.Now()
	pollInterval := time.Second

	for {
		if time.Since(start) > timeout {
			return nil, fmt.Errorf("timeout waiting for task result after %v", timeout)
		}

		task, err := s.inspector.GetTaskInfo(tasks.QUEUE_NAME, taskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get task info: %w", err)
		}

		switch task.State {
		case asynq.TaskStateCompleted:
			s.logger.Info("Task completed successfully")
			return task.Result, nil
		case asynq.TaskStateArchived:
			return nil, fmt.Errorf("task archived: %s", task.LastErr)
		case asynq.TaskStateRetry:
			s.logger.Debug("Task scheduled for retry...")
		case asynq.TaskStatePending, asynq.TaskStateActive, asynq.TaskStateScheduled:
			s.logger.Debug("Task still in progress, waiting...")
		case asynq.TaskStateAggregating:
			s.logger.Debug("Task aggregating, waiting...")
		default:
			return nil, fmt.Errorf("unexpected task state: %s", task.State)
		}

		time.Sleep(pollInterval)
	}
}
