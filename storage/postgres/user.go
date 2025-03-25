package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/vultisig/vultisigner/internal/types"
)

const USERS_TABLE = "users"

func (p *PostgresBackend) FindUserById(ctx context.Context, userId string) (*types.User, error) {
	query := fmt.Sprintf(`SELECT id, username, created_at FROM %s WHERE id = $1 LIMIT 1;`, USERS_TABLE)

	rows, err := p.pool.Query(ctx, query, userId)
	if err != nil {
		return nil, err
	}

	order, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[types.User])
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (p *PostgresBackend) FindUserByName(ctx context.Context, username string) (*types.UserWithPassword, error) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE username = $1 LIMIT 1;`, USERS_TABLE)

	rows, err := p.pool.Query(ctx, query, username)
	if err != nil {
		return nil, err
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[types.UserWithPassword])
	if err != nil {
		return nil, err
	}

	return &user, nil
}
