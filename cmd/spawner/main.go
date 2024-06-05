package main

import (
	"log"
	"time"

	"vultisigner/pkg/tasks"

	"github.com/hibiken/asynq"
)

const redisAddr = "127.0.0.1:6371"

func main() {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	defer client.Close()

	task, err := tasks.NewKeyGeneration("local_key", "session_id", "chain_code")
	if err != nil {
		log.Fatalf("could not create task: %v", err)
	}

	// this will make sure there is only one task active, the lock lasts 1 hour, if the job somehow lasts over 1 hour, the lock will be released
	// and then it will be possible to enqueue another task, this is useful for if it's stuck
	info, err := client.Enqueue(task, asynq.MaxRetry(10), asynq.Timeout(3*time.Minute), asynq.Unique(time.Hour))
	if err != nil {
		log.Fatalf("could not enqueue task: %v", err)
	}
	log.Printf("enqueued task: id=%s queue=%s", info.ID, info.Queue)
}
