package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/contexthelper"
	"github.com/vultisig/vultisigner/internal/types"
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

func (r *RedisStorage) SetVaultCacheItem(ctx context.Context, vault *types.VaultCacheItem) error {
	if contexthelper.CheckCancellation(ctx) != nil {
		return ctx.Err()
	}
	key := vault.Key()
	vaultJSON, err := json.Marshal(vault)
	if err != nil {
		return fmt.Errorf("fail to serialize vault cache item to json, err: %w", err)
	}
	return r.client.Set(ctx, key, string(vaultJSON), 0).Err()
}

// GetVaultCacheItem returns a vault cache item by its key.
func (r *RedisStorage) GetVaultCacheItem(ctx context.Context, key string) (*types.VaultCacheItem, error) {
	if contexthelper.CheckCancellation(ctx) != nil {
		return nil, ctx.Err()
	}
	vaultJSON, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("fail to get vault cache item, err: %w", err)
	}
	var vault types.VaultCacheItem
	if err := json.Unmarshal([]byte(vaultJSON), &vault); err != nil {
		return nil, fmt.Errorf("fail to deserialize vault cache item, err: %w", err)
	}
	return &vault, nil
}

func (r *RedisStorage) Close() error {
	return r.client.Close()
}
