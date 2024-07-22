package main

import (
	"log"

	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"

	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/internal/logging"
	"github.com/vultisig/vultisigner/internal/tasks"
	"github.com/vultisig/vultisigner/service"
)

func main() {
	workerServce := &service.WorkerService{}
	err := workerServce.Initialize()
	if err != nil {
		log.Fatalf("workerServce.Initialize failed: %v", err)
		return
	}

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
	mux.HandleFunc(tasks.TypeKeyGeneration, workerServce.HandleKeyGeneration)
	// mux.Handle(tasks.TypeKeyGeneration, tasks.I())
	// ...register other handlers...

	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
