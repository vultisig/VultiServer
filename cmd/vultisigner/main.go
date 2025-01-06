package main

import (
	"fmt"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/hibiken/asynq"

	"github.com/vultisig/vultisigner/api"
	"github.com/vultisig/vultisigner/config"
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
	server := api.NewServer(
		cfg.Server.Port,
		redisStorage,
		redisOptions,
		client,
		inspector,
		cfg.Server.VaultsFilePath,
		sdClient,
		blockStorage,
		cfg.Server.Mode,
		cfg.Plugin.Type,
		cfg.Server.Database.DSN,
	)
	if err := server.StartServer(); err != nil {
		panic(err)
	}
}
