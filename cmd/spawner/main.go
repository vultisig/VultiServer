package main

import (
	"log"
	"time"

	"github.com/vultisig/vultisigner/internal/tasks"

	"github.com/hibiken/asynq"
)

const redisAddr = "127.0.0.1:6371"

func main() {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	defer client.Close()

	session := "20"

	task, err := tasks.NewKeyGeneration("vultisigner", "test", session, "80871c0f885f953e5206e461630a9222148797e66276a83224c7b9b0f75b3ec0", "6e2ca8460d140f09f2879335af040af8ec8791f1a5588ea69612b1dbd701bb4b", "")
	if err != nil {
		log.Fatalf("could not create task: %v", err)
	}

	// this will make sure there is only one task active, the lock lasts 1 hour, if the job somehow lasts over 1 hour, the lock will be released
	// and then it will be possible to enqueue another task, this is useful for if it's stuck
	info, err := client.Enqueue(task, asynq.MaxRetry(1), asynq.Timeout(1*time.Minute), asynq.Unique(time.Hour))
	// info, err := client.Enqueue(task, asynq.MaxRetry(10), asynq.Timeout(3*time.Minute), asynq.Unique(time.Hour))
	if err != nil {
		log.Fatalf("could not enqueue task: %v", err)
	}
	log.Printf("enqueued task: id=%s queue=%s", info.ID, info.Queue)

	// 2
	task, err = tasks.NewKeyGeneration("test", "test", session, "80871c0f885f953e5206e461630a9222148797e66276a83224c7b9b0f75b3ec0", "6e2ca8460d140f09f2879335af040af8ec8791f1a5588ea69612b1dbd701bb4b", "")
	if err != nil {
		log.Fatalf("could not create task: %v", err)
	}

	// this will make sure there is only one task active, the lock lasts 1 hour, if the job somehow lasts over 1 hour, the lock will be released
	// and then it will be possible to enqueue another task, this is useful for if it's stuck
	info, err = client.Enqueue(task, asynq.MaxRetry(1), asynq.Timeout(1*time.Minute), asynq.Unique(time.Hour))
	// info, err := client.Enqueue(task, asynq.MaxRetry(10), asynq.Timeout(3*time.Minute), asynq.Unique(time.Hour))
	if err != nil {
		log.Fatalf("could not enqueue task: %v", err)
	}
	log.Printf("enqueued task: id=%s queue=%s", info.ID, info.Queue)
}
