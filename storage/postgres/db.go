package postgres

import (
	"context"
	"embed"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/vultisigner/internal/types"
)

//go:embed migrations/*
var embeddedMigrations embed.FS

type PostgresBackend struct {
	pool *pgxpool.Pool
}

func NewPostgresBackend(readonly bool, dsn string) (*PostgresBackend, error) {
	logrus.Info("Connecting to database with DSN: ", dsn)
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	backend := &PostgresBackend{
		pool: pool,
	}

	if err := backend.Migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return backend, nil
}

func (d *PostgresBackend) Close() error {
	d.pool.Close()

	return nil
}

func (d *PostgresBackend) Migrate() error {
	logrus.Info("Starting database migration...")
	goose.SetBaseFS(embeddedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	db := stdlib.OpenDBFromPool(d.pool)
	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("failed to run goose up: %w", err)
	}
	logrus.Info("Database migration completed successfully")
	return nil
}

func (p *PostgresBackend) CreateTransactionHistory(tx types.TransactionHistory) (uuid.UUID, error) {
	query := `
        INSERT INTO transaction_history (
            policy_id, tx_body, status, metadata
        ) VALUES ($1, $2, $3, $4)
				RETURNING id
    `
	var txID uuid.UUID
	err := p.pool.QueryRow(context.Background(), query,
		tx.PolicyID,
		tx.TxBody,
		tx.Status,
		tx.Metadata,
	).Scan(&txID)

	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create transaction history: %w", err)
	}

	return txID, nil
}

func (p *PostgresBackend) UpdateTransactionStatus(txID uuid.UUID, status types.TransactionStatus, metadata map[string]interface{}) error {
	query := `
        UPDATE transaction_history 
        SET status = $1, metadata = metadata || $2::jsonb, updated_at = NOW()
        WHERE id = $3
    `

	_, err := p.pool.Exec(context.Background(), query, status, metadata, txID)
	return err
}

func (p *PostgresBackend) GetTransactionHistory(policyID uuid.UUID, take int, skip int) ([]types.TransactionHistory, error) {
	query := `
        SELECT id, policy_id, tx_body, status, created_at, updated_at, metadata, error_message
        FROM transaction_history
        WHERE policy_id = $1
        ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
    `

	rows, err := p.pool.Query(context.Background(), query, policyID, take, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []types.TransactionHistory
	for rows.Next() {
		var tx types.TransactionHistory
		err := rows.Scan(
			&tx.ID,
			&tx.PolicyID,
			&tx.TxBody,
			&tx.Status,
			&tx.CreatedAt,
			&tx.UpdatedAt,
			&tx.Metadata,
			&tx.ErrorMessage,
		)
		if err != nil {
			return nil, err
		}
		history = append(history, tx)
	}

	return history, nil
}

func (p *PostgresBackend) Pool() *pgxpool.Pool {
	return p.pool
}
