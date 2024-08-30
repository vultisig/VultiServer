package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/mailgun/mailgun-go/v4"
	"github.com/sirupsen/logrus"

	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/contexthelper"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/storage"
)

type WorkerService struct {
	cfg         config.Config
	redis       *storage.RedisStorage
	logger      *logrus.Logger
	queueClient *asynq.Client
}

// NewWorker creates a new worker service
func NewWorker(cfg config.Config, queueClient *asynq.Client) (*WorkerService, error) {
	redis, err := storage.NewRedisStorage(cfg)
	if err != nil {
		return nil, fmt.Errorf("storage.NewRedisStorage failed: %w", err)
	}

	return &WorkerService{
		redis:       redis,
		cfg:         cfg,
		logger:      logrus.WithField("service", "worker").Logger,
		queueClient: queueClient,
	}, nil
}

type KeyGenerationTaskResult struct {
	EDDSAPublicKey string
	ECDSAPublicKey string
}

func (s *WorkerService) HandleKeyGeneration(ctx context.Context, t *asynq.Task) error {
	if err := contexthelper.CheckCancellation(ctx); err != nil {
		return err
	}

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

	keyECDSA, keyEDDSA, err := s.JoinKeyGeneration(req)
	if err != nil {
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
		return fmt.Errorf("json.Marshal failed: %v: %w", err, asynq.SkipRetry)
	}

	if _, err := t.ResultWriter().Write(resultBytes); err != nil {
		return fmt.Errorf("t.ResultWriter.Write failed: %v: %w", err, asynq.SkipRetry)
	}

	return nil
}

func (s *WorkerService) HandleKeySign(ctx context.Context, t *asynq.Task) error {
	if err := contexthelper.CheckCancellation(ctx); err != nil {
		return err
	}
	var p types.KeysignRequest
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	s.logger.WithFields(logrus.Fields{
		"PublicKey":  p.PublicKey,
		"session":    p.SessionID,
		"Messages":   p.Messages,
		"DerivePath": p.DerivePath,
		"IsECDSA":    p.IsECDSA,
	}).Info("joining keysign")

	signatures, err := s.JoinKeySign(p)
	if err != nil {
		return fmt.Errorf("join keysign failed: %v: %w", err, asynq.SkipRetry)
	}

	s.logger.WithFields(logrus.Fields{
		"Signatures": signatures,
	}).Info("localPartyID sign completed")

	resultBytes, err := json.Marshal(signatures)
	if err != nil {
		return fmt.Errorf("json.Marshal failed: %v: %w", err, asynq.SkipRetry)
	}

	if _, err := t.ResultWriter().Write(resultBytes); err != nil {
		return fmt.Errorf("t.ResultWriter.Write failed: %v: %w", err, asynq.SkipRetry)
	}

	return nil
}
func (s *WorkerService) HandleEmailVaultBackup(ctx context.Context, t *asynq.Task) error {
	if err := contexthelper.CheckCancellation(ctx); err != nil {
		return err
	}
	var req types.EmailRequest
	if err := json.Unmarshal(t.Payload(), &req); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	s.logger.WithFields(logrus.Fields{
		"email":    req.Email,
		"filename": req.FileName,
	}).Info("sending email")

	mg := mailgun.NewMailgun("vultisig.com", s.cfg.EmailServer.ApiKey)
	msg := mg.NewMessage("fastvault@vultisig.com", "Vault Backup ", "Your vault backup is ready", req.Email)
	msg.AddBufferAttachment(req.FileName, []byte(req.FileContent))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	resp, id, err := mg.Send(ctx, msg)
	if err != nil {
		return fmt.Errorf("mg.Send failed: %v", err)
	}
	s.logger.WithFields(logrus.Fields{
		"resp": resp,
		"id":   id,
	}).Info("email sent")

	if _, err := t.ResultWriter().Write([]byte("email sent:" + resp + ",id:" + id)); err != nil {
		return fmt.Errorf("t.ResultWriter.Write failed: %v: %w", err, asynq.SkipRetry)
	}
	return nil
}
