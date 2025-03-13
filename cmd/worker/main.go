package main

import (
	"fmt"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"

	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/internal/syncer"
	"github.com/vultisig/vultisigner/internal/tasks"
	"github.com/vultisig/vultisigner/service"
	"github.com/vultisig/vultisigner/storage"
)

func main() {
	cfg, err := config.GetConfigure()
	if err != nil {
		panic(err)
	}
	sdClient, err := statsd.New(cfg.Datadog.Host + ":" + cfg.Datadog.Port)
	if err != nil {
		panic(err)
	}
	blockStorage, err := storage.NewBlockStorage(*cfg)
	if err != nil {
		panic(err)
	}
	redisOptions := asynq.RedisClientOpt{
		Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
		Username: cfg.Redis.User,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}
	logger := logrus.StandardLogger()
	verifierConfig, err := config.ReadConfig("config-verifier")
	if err != nil {
		panic(err)
	}
	syncer := syncer.NewPolicySyncer(logger, verifierConfig)
	client := asynq.NewClient(redisOptions)
	inspector := asynq.NewInspector(redisOptions)
	workerService, err := service.NewWorker(*cfg, client, sdClient, syncer, blockStorage, inspector)
	if err != nil {
		panic(err)
	}

	srv := asynq.NewServer(
		redisOptions,
		asynq.Config{
			Logger:      logger,
			Concurrency: 10,
			Queues: map[string]int{
				tasks.QUEUE_NAME:         10,
				tasks.EMAIL_QUEUE_NAME:   100,
				"scheduled_plugin_queue": 10, // new queue
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeKeyGeneration, workerService.HandleKeyGeneration)
	mux.HandleFunc(tasks.TypeKeySign, workerService.HandleKeySign)
	mux.HandleFunc(tasks.TypeEmailVaultBackup, workerService.HandleEmailVaultBackup)
	mux.HandleFunc(tasks.TypeReshare, workerService.HandleReshare)
	mux.HandleFunc(tasks.TypePluginTransaction, workerService.HandlePluginTransaction)

	if err := srv.Run(mux); err != nil {
		panic(fmt.Errorf("could not run server: %w", err))
	}
}
