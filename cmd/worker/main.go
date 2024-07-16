package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/sirupsen/logrus"
	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/internal/keygen"
	"github.com/vultisig/vultisigner/internal/logging"
	"github.com/vultisig/vultisigner/internal/tasks"
	"github.com/vultisig/vultisigner/internal/types"

	"github.com/hibiken/asynq"
)

func main() {
	redisAddr := config.AppConfig.Redis.Host + ":" + config.AppConfig.Redis.Port

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	logging.Logger.WithFields(logrus.Fields{
		"redis": redisAddr,
	}).Info("Starting server")

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeKeyGeneration, HandleKeyGeneration)
	// mux.Handle(tasks.TypeKeyGeneration, tasks.I())
	// ...register other handlers...

	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}

type KeyGenerationTaskResult struct {
	EDDSAPublicKey string
	ECDSAPublicKey string
}

func HandleKeyGeneration(ctx context.Context, t *asynq.Task) error {
	var p tasks.KeyGenerationPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	logging.Logger.WithFields(logrus.Fields{
		"session":    p.SessionID,
		"local_key":  p.LocalKey,
		"chain_code": p.ChainCode,
		"HexEncryptionKey": p.HexEncryptionKey,
	}).Info("Joining keygen")

	keyECDSA, keyEDDSA, err := keygen.JoinKeyGeneration(&types.KeyGeneration{
		Key:              p.LocalKey,
		Session:          p.SessionID,
		ChainCode:        p.ChainCode,
		HexEncryptionKey: p.HexEncryptionKey,
	})
	if err != nil {
		return fmt.Errorf("keygen.JoinKeyGeneration failed: %v: %w", err, asynq.SkipRetry)
	}

	logging.Logger.WithFields(logrus.Fields{
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

	if _, err := t.ResultWriter().Write([]byte(resultBytes)); err != nil {
		return fmt.Errorf("t.ResultWriter.Write failed: %v: %w", err, asynq.SkipRetry)
	}

	return nil
}
