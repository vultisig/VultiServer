package storage

import (
	"context"

	"github.com/redis/go-redis/v9"

	"github.com/vultisig/vultisigner/config"
)

type RedisStorage struct {
	cfg    config.Config
	client *redis.Client
}

func NewRedisStorage(cfg config.Config) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
		Username: cfg.Redis.User,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	status := client.Ping(context.Background())
	if status.Err() != nil {
		return nil, status.Err()
	}
	return &RedisStorage{
		cfg:    cfg,
		client: client,
	}, nil
}

func (r *RedisStorage) Close() error {
	return r.client.Close()
}
