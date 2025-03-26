package storage

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vultisig/vultisigner/internal/types"
)

type DatabaseStorage interface {
	Close() error

	FindUserById(ctx context.Context, userId string) (*types.User, error)
	FindUserByName(ctx context.Context, username string) (*types.UserWithPassword, error)

	GetPluginPolicy(ctx context.Context, id string) (types.PluginPolicy, error)
	GetAllPluginPolicies(ctx context.Context, publicKey string, pluginType string) ([]types.PluginPolicy, error)
	DeletePluginPolicyTx(ctx context.Context, dbTx pgx.Tx, id string) error
	InsertPluginPolicyTx(ctx context.Context, dbTx pgx.Tx, policy types.PluginPolicy) (*types.PluginPolicy, error)
	UpdatePluginPolicyTx(ctx context.Context, dbTx pgx.Tx, policy types.PluginPolicy) (*types.PluginPolicy, error)

	FindPricingById(ctx context.Context, id string) (*types.Pricing, error)
	CreatePricing(ctx context.Context, pricingDto types.PricingCreateDto) (*types.Pricing, error)
	DeletePricingById(ctx context.Context, id string) error

	CreateTimeTriggerTx(ctx context.Context, dbTx pgx.Tx, trigger types.TimeTrigger) error
	GetPendingTimeTriggers(ctx context.Context) ([]types.TimeTrigger, error)
	UpdateTimeTriggerLastExecution(ctx context.Context, policyID string) error
	UpdateTimeTriggerTx(ctx context.Context, policyID string, trigger types.TimeTrigger, dbTx pgx.Tx) error

	DeleteTimeTrigger(ctx context.Context, policyID string) error
	UpdateTriggerStatus(ctx context.Context, policyID string, status types.TimeTriggerStatus) error
	GetTriggerStatus(ctx context.Context, policyID string) (types.TimeTriggerStatus, error)

	CountTransactions(ctx context.Context, policyID uuid.UUID, status types.TransactionStatus, txType string) (int64, error)
	CreateTransactionHistoryTx(ctx context.Context, dbTx pgx.Tx, tx types.TransactionHistory) (uuid.UUID, error)
	UpdateTransactionStatusTx(ctx context.Context, dbTx pgx.Tx, txID uuid.UUID, status types.TransactionStatus, metadata map[string]interface{}) error
	CreateTransactionHistory(ctx context.Context, tx types.TransactionHistory) (uuid.UUID, error)
	UpdateTransactionStatus(ctx context.Context, txID uuid.UUID, status types.TransactionStatus, metadata map[string]interface{}) error
	GetTransactionHistory(ctx context.Context, policyID uuid.UUID, transactionType string, take int, skip int) ([]types.TransactionHistory, error)
	GetTransactionByHash(ctx context.Context, txHash string) (*types.TransactionHistory, error)

	FindPlugins(ctx context.Context, take int, skip int, sort string) (types.PlugisDto, error)
	FindPluginById(ctx context.Context, id string) (*types.Plugin, error)
	CreatePlugin(ctx context.Context, pluginDto types.PluginCreateDto) (*types.Plugin, error)
	UpdatePlugin(ctx context.Context, id string, updates types.PluginUpdateDto) (*types.Plugin, error)
	DeletePluginById(ctx context.Context, id string) error

	Pool() *pgxpool.Pool
}
