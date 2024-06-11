package main

import (
	"fmt"

	"github.com/hibiken/asynq"

	"github.com/vultisig/vultisigner/api"
	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/storage"
)

func main() {
	port := config.AppConfig.Server.Port

	redisStorage, err := storage.NewRedisStorage(config.AppConfig)
	if err != nil {
		panic(err)
	}
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     config.AppConfig.Redis.Host + ":" + config.AppConfig.Redis.Port,
		Username: config.AppConfig.Redis.User,
		Password: config.AppConfig.Redis.Password,
		DB:       config.AppConfig.Redis.DB,
	})
	defer func() {
		if err := client.Close(); err != nil {
			fmt.Println("fail to close asynq client,", err)
		}
	}()
	server := api.NewServer(port, redisStorage, client)
	if err := server.StartServer(); err != nil {
		panic(err)
	}
}
