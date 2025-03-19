package main

import (
	"fmt"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"

	"github.com/vultisig/vultisigner/api"
	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/storage"
	"github.com/vultisig/vultisigner/storage/postgres"
)

func main() {
	cfg, err := config.GetConfigure()
	if err != nil {
		panic(err)
	}

	logger := logrus.New()

	sdClient, err := statsd.New(fmt.Sprintf("%s:%s", cfg.Datadog.Host, cfg.Datadog.Port))
	if err != nil {
		panic(err)
	}

	redisStorage, err := storage.NewRedisStorage(*cfg)
	if err != nil {
		panic(err)
	}

	redisOptions := asynq.RedisClientOpt{
		Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
		Username: cfg.Redis.User,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	client := asynq.NewClient(redisOptions)
	defer func() {
		if err := client.Close(); err != nil {
			fmt.Println("fail to close asynq client,", err)
		}
	}()

	inspector := asynq.NewInspector(redisOptions)
	if cfg.Server.VaultsFilePath == "" {
		panic("vaults file path is empty")

	}
	blockStorage, err := storage.NewBlockStorage(*cfg)
	if err != nil {
		panic(err)
	}

	db, err := postgres.NewPostgresBackend(false, cfg.Server.Database.DSN)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}

	server := api.NewServer(
		cfg.Server.Port,
		db,
		redisStorage,
		blockStorage,
		redisOptions,
		client,
		inspector,
		sdClient,
		cfg.Server.VaultsFilePath,
		cfg.Server.Mode,
		cfg.Server.JWTSecret,
		cfg.Server.Plugin.Type,
		cfg.Server.Plugin.Eth.Rpc,
		logger,
	)
	if err := server.StartServer(); err != nil {
		panic(err)
	}
}
