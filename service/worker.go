package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"

	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/contexthelper"
	"github.com/vultisig/vultisigner/internal/tasks"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/storage"
)

type WorkerService struct {
	cfg    config.Config
	redis  *storage.RedisStorage
	logger *logrus.Logger
}

// NewWorker creates a new worker service
func NewWorker(cfg config.Config) (*WorkerService, error) {
	redis, err := storage.NewRedisStorage(cfg)
	if err != nil {
		return nil, fmt.Errorf("storage.NewRedisStorage failed: %w", err)
	}

	return &WorkerService{
		redis:  redis,
		cfg:    cfg,
		logger: logrus.WithField("service", "worker").Logger,
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
	}).Info("Key generation completed")

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

	var p tasks.KeysignPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	logging.Logger.WithFields(logrus.Fields{
		"PublicKey":        p.PublicKey,
		"session":          p.SessionID,
		"Messages":         p.Messages,
		"HexEncryptionKey": p.HexEncryptionKey,
		"DerivePath":       p.DerivePath,
		"IsECDSA":          p.IsECDSA,
	}).Info("Joining keysign")

	signatures, err := JoinKeySign(&types.KeysignRequest{
		PublicKey:        p.PublicKey,
		Session:          p.SessionID,
		Messages:         p.Messages,
		HexEncryptionKey: p.HexEncryptionKey,
		DerivePath:       p.DerivePath,
		IsECDSA:          p.IsECDSA,
		VaultPassword:    p.VaultPassword,
	})
	if err != nil {
		return fmt.Errorf("keysign.JoinKeySign failed: %v: %w", err, asynq.SkipRetry)
	}

	logging.Logger.WithFields(logrus.Fields{
		"Signatures": signatures,
	}).Info("Key sign completed")

	resultBytes, err := json.Marshal(signatures)
	if err != nil {
		return fmt.Errorf("json.Marshal failed: %v: %w", err, asynq.SkipRetry)
	}

	if _, err := t.ResultWriter().Write(resultBytes); err != nil {
		return fmt.Errorf("t.ResultWriter.Write failed: %v: %w", err, asynq.SkipRetry)
	}

	return nil
}
