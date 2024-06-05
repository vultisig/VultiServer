// package main

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"log"

// 	"vultisigner/pkg/tasks"

// 	"github.com/hibiken/asynq"
// )

// func HandleKeyGeneration(ctx context.Context, t *asynq.Task) error {
// 	var p tasks.KeyGenerationPayload
// 	if err := json.Unmarshal(t.Payload(), &p); err != nil {
// 		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
// 	}
// 	log.Printf("Joining keygen for user: user_id=%d", p.UserID)

// 	// join key generation here

// 	return nil
// }
