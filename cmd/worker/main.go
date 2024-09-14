package main

import (
	"fmt"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"

	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/internal/tasks"
	"github.com/vultisig/vultisigner/service"
)

func main() {
	cfg, err := config.GetConfigure()
	if err != nil {
		panic(err)
	}
	sdClient, err := statsd.New("127.0.0.1:8125")
	if err != nil {
		panic(err)
	}
	redisAddr := cfg.Redis.Host + ":" + cfg.Redis.Port
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	workerServce, err := service.NewWorker(*cfg, client, sdClient)
	if err != nil {
		panic(err)
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Logger:      logrus.StandardLogger(),
			Concurrency: 10,
			Queues: map[string]int{
				tasks.QUEUE_NAME:       10,
				tasks.EMAIL_QUEUE_NAME: 100,
			},
		},
	)

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeKeyGeneration, workerServce.HandleKeyGeneration)
	mux.HandleFunc(tasks.TypeKeySign, workerServce.HandleKeySign)
	mux.HandleFunc(tasks.TypeEmailVaultBackup, workerServce.HandleEmailVaultBackup)
	mux.HandleFunc(tasks.TypeReshare, workerServce.HandleReshare)
	if err := srv.Run(mux); err != nil {
		panic(fmt.Errorf("could not run server: %w", err))
	}
}
