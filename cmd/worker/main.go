package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/internal/keygen"
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

	fmt.Println("Worker is running...")

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
	log.Printf("Joining keygen for local key: local_key=%s, session_id=%s, chain_code=%s", p.LocalKey, p.SessionID, p.ChainCode)

	// Join keygen
	keyECDSA, keyEDDSA, err := keygen.JoinKeyGeneration(&types.KeyGeneration{
		Key: p.LocalKey,
		// Parties:   p.Parties,
		Session:   p.SessionID,
		ChainCode: p.ChainCode,
	})
	if err != nil {
		return fmt.Errorf("keygen.JoinKeyGeneration failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("Key generation completed keys: keyECDSA=%s, keyEDDSA=%s", keyECDSA, keyEDDSA)

	result := KeyGenerationTaskResult{
		EDDSAPublicKey: keyEDDSA,
		ECDSAPublicKey: keyECDSA,
	}

	resultBytes, err := json.Marshal(result)

	if _, err := t.ResultWriter().Write([]byte(resultBytes)); err != nil {
		return fmt.Errorf("t.ResultWriter.Write failed: %v: %w", err, asynq.SkipRetry)
	}

	return nil
}
