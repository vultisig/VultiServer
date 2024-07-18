package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/internal/keygen"
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

	if _, err := t.ResultWriter().Write([]byte(resultBytes)); err != nil {
		return fmt.Errorf("t.ResultWriter.Write failed: %v: %w", err, asynq.SkipRetry)
	}

	return nil
}
