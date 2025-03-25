package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/vultisig/vultisigner/internal/types"
)

const PRICINGS_TABLE = "pricings"

func (p *PostgresBackend) FindPricingById(ctx context.Context, id string) (*types.Pricing, error) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE id = $1 LIMIT 1;`, PRICINGS_TABLE)

	rows, err := p.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}

	pricing, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[types.Pricing])
	if err != nil {
		return nil, err
	}

	return &pricing, nil
}

func (p *PostgresBackend) CreatePricing(ctx context.Context, pricingDto types.PricingCreateDto) (*types.Pricing, error) {
	columns := []string{"type", "amount", "metric"}
	argNames := []string{"@Type", "@Amount", "@Metric"}
	args := pgx.NamedArgs{
		"Type":   pricingDto.Type,
		"Amount": pricingDto.Amount,
		"Metric": pricingDto.Metric,
	}

	if pricingDto.Frequency != "" {
		columns = append(columns, "frequency")
		argNames = append(argNames, "@Frequency")
		args["Frequency"] = pricingDto.Frequency
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s) RETURNING id;`,
		PRICINGS_TABLE,
		strings.Join(columns, ", "),
		strings.Join(argNames, ", "),
	)

	var createdId string
	err := p.pool.QueryRow(ctx, query, args).Scan(&createdId)
	if err != nil {
		return nil, err
	}

	return p.FindPricingById(ctx, createdId)
}

func (p *PostgresBackend) DeletePricingById(ctx context.Context, id string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1;`, PRICINGS_TABLE)

	_, err := p.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}
