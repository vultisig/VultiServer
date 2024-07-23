package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"

	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/internal/keygen"
	"github.com/vultisig/vultisigner/internal/keysign"
	"github.com/vultisig/vultisigner/internal/logging"
	"github.com/vultisig/vultisigner/internal/tasks"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/storage"
)

type WorkerService struct {
	redis *storage.RedisStorage
}

func (s *WorkerService) Initialize() error {
	redis, err := storage.NewRedisStorage(config.AppConfig)
	if err != nil {
		return fmt.Errorf("storage.NewRedisStorage failed: %w", err)
	}
	s.redis = redis
	return nil
}

type KeyGenerationTaskResult struct {
	EDDSAPublicKey string
	ECDSAPublicKey string
}

type KeySignTaskResult struct {
	SignatureEncoded []string
}

func (s *WorkerService) HandleKeyGeneration(ctx context.Context, t *asynq.Task) error {
	var p tasks.KeyGenerationPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	logging.Logger.WithFields(logrus.Fields{
		"name":             p.Name,
		"session":          p.SessionID,
		"local_key":        p.LocalKey,
		"chain_code":       p.ChainCode,
		"HexEncryptionKey": p.HexEncryptionKey,
	}).Info("Joining keygen")

	keyECDSA, keyEDDSA, err := keygen.JoinKeyGeneration(&types.KeyGeneration{
		Name:               p.Name,
		Key:                p.LocalKey,
		Session:            p.SessionID,
		ChainCode:          p.ChainCode,
		HexEncryptionKey:   p.HexEncryptionKey,
		EncryptionPassword: p.EncryptionPassword,
	})
	if err != nil {
		return fmt.Errorf("keygen.JoinKeyGeneration failed: %v: %w", err, asynq.SkipRetry)
	}

	logging.Logger.WithFields(logrus.Fields{
		"keyECDSA": keyECDSA,
		"keyEDDSA": keyEDDSA,
	}).Info("Key generation completed")

	err = s.redis.RemoveVaultCacheItem(ctx, fmt.Sprintf("vault-%s-%s", p.Name, p.SessionID))
	if err != nil {
		logging.Logger.Errorf("redis.RemoveVaultCacheItem failed: %v", err)
	}

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
	var p tasks.KeysignPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	logging.Logger.WithFields(logrus.Fields{
		"PublicKeyECDSA":   p.PublicKeyECDSA,
		"session":          p.SessionID,
		"local_key":        p.LocalKey,
		"Messages":         p.Messages,
		"HexEncryptionKey": p.HexEncryptionKey,
		"DerivePath":       p.DerivePath,
		"IsECDSA":          p.IsECDSA,
	}).Info("Joining keygen")

	resp, err := keysign.JoinKeySign(&types.KeysignRequest{
		PublicKeyECDSA:   p.PublicKeyECDSA,
		Key:              p.LocalKey,
		Session:          p.SessionID,
		Messages:         p.Messages,
		HexEncryptionKey: p.HexEncryptionKey,
		DerivePath:       p.DerivePath,
		IsECDSA:          p.IsECDSA,
	})
	if err != nil {
		return fmt.Errorf("keysign.JoinKeySign failed: %v: %w", err, asynq.SkipRetry)
	}

	logging.Logger.WithFields(logrus.Fields{
		"signatureEncoded": resp,
	}).Info("Key sign completed")

	result := KeySignTaskResult{
		SignatureEncoded: resp,
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
