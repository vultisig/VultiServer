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
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	vaultType "github.com/vultisig/commondata/go/vultisig/vault/v1"

	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/contexthelper"
	"github.com/vultisig/vultisigner/internal/tasks"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/plugin"
	"github.com/vultisig/vultisigner/plugin/payroll"
	"github.com/vultisig/vultisigner/relay"
	"github.com/vultisig/vultisigner/storage"
	"github.com/vultisig/vultisigner/storage/postgres"
)

type WorkerService struct {
	cfg          config.Config
	redis        *storage.RedisStorage
	logger       *logrus.Logger
	queueClient  *asynq.Client
	sdClient     *statsd.Client
	blockStorage *storage.BlockStorage
	plugin       plugin.Plugin
	db           storage.DatabaseStorage
}

// NewWorker creates a new worker service
func NewWorker(cfg config.Config, queueClient *asynq.Client, sdClient *statsd.Client, blockStorage *storage.BlockStorage) (*WorkerService, error) {
	redis, err := storage.NewRedisStorage(cfg)
	if err != nil {
		return nil, fmt.Errorf("storage.NewRedisStorage failed: %w", err)
	}

	db, err := postgres.NewPostgresBackend(false, cfg.Database.DSN)
	if err != nil {
		logrus.Fatalf("Failed to connect to database: %v", err)
	}

	var plugin plugin.Plugin
	if cfg.Server.Mode == "pluginserver" {
		switch cfg.Plugin.Type {
		case "payroll":
			plugin = payroll.NewPayrollPlugin(db)
		default:
			logrus.Fatalf("Invalid plugin type: %s", cfg.Plugin.Type)
		}
	}

	return &WorkerService{
		redis:        redis,
		cfg:          cfg,
		logger:       logrus.WithField("service", "worker").Logger,
		queueClient:  queueClient,
		sdClient:     sdClient,
		blockStorage: blockStorage,
		plugin:       plugin,
		db:           db,
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
	s.logger.Info("Starting HandleKeySign")
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
	s.incCounter("worker.plugin.transaction", []string{})
	s.logger.WithFields(logrus.Fields{
		"policy_id": triggerEvent.PolicyID,
	}).Info("plugin transaction request")

	policy, err := s.db.GetPluginPolicy(triggerEvent.PolicyID)
	if err != nil {
		s.logger.Errorf("db.GetPluginPolicy failed: %v", err)
		return fmt.Errorf("db.GetPluginPolicy failed: %v: %w", err, asynq.SkipRetry)
	}

	s.logger.WithFields(logrus.Fields{
		"policy_id":   policy.ID,
		"public_key":  policy.PublicKey,
		"plugin_type": policy.PluginType,
	}).Info("Retrieved policy for signing")

	signRequest, err := plugin.ProposeTransactions(policy)
	if err != nil {
		s.logger.Errorf("Failed to create signing request: %v", err)
		return fmt.Errorf("Failed to create signing request: %v: %w", err, asynq.SkipRetry)
	}

	for _, signRequest := range signRequest {
		signBytes, err := json.Marshal(signRequest)
		if err != nil {
			s.logger.Errorf("Failed to marshal sign request: %v", err)
			continue
		}

		signResp, err := http.Post(
			fmt.Sprintf("http://localhost:%d/signFromPlugin", 8080),
			"application/json",
			bytes.NewBuffer(signBytes),
		)
		if err != nil {
			s.logger.Errorf("Failed to make sign request: %v", err)
			return err
		}
		defer signResp.Body.Close()

		respBody, err := io.ReadAll(signResp.Body)
		if err != nil {
			s.logger.Errorf("Failed to read response: %v", err)
			return err
		}

		if signResp.StatusCode == http.StatusOK {
			// Enqueue the same signing request locally
			signRequest.KeysignRequest.StartSession = true
			signRequest.KeysignRequest.Parties = []string{"1", "2"}
			buf, err := json.Marshal(signRequest.KeysignRequest)
			if err != nil {
				s.logger.Errorf("Failed to marshal local sign request: %v", err)
				return err
			}

			// Enqueue TypeKeySign directly
			ti, err := s.queueClient.Enqueue(
				asynq.NewTask(tasks.TypeKeySign, buf),
				asynq.MaxRetry(-1),
				asynq.Timeout(2*time.Minute),
				asynq.Retention(5*time.Minute),
				asynq.Queue(tasks.QUEUE_NAME),
			)
			if err != nil {
				s.logger.Errorf("Failed to enqueue signing task: %v", err)
				continue
			}

			s.logger.Infof("Enqueued signing task: %s", ti.ID)

			if err := s.db.UpdateTriggerExecution(policy.ID); err != nil {
				s.logger.Errorf("Failed to update last execution: %v", err)
			}
			//update here tx history
		}

		s.logger.Infof("Plugin signing test complete. Status: %d, Response: %s",
			signResp.StatusCode, string(respBody))
	}

	return nil
}
