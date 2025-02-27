package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"time"

	"github.com/vultisig/vultisigner/internal/types"
)

func (p *PostgresBackend) CreateTimeTriggerTx(ctx context.Context, tx pgx.Tx, trigger types.TimeTrigger) error {
	_, err := tx.Exec(ctx, `
        INSERT INTO time_triggers 
        (policy_id, cron_expression, start_time, end_time, frequency, interval) 
        VALUES ($1, $2, $3, $4, $5, $6)`,
		trigger.PolicyID,
		trigger.CronExpression,
		trigger.StartTime,
		trigger.EndTime,
		trigger.Frequency,
		trigger.Interval,
	)
	return err
}

func (p *PostgresBackend) GetPendingTriggers() ([]types.TimeTrigger, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("database pool is nil")
	}

	query := `
        SELECT policy_id, cron_expression, start_time, end_time, frequency, interval, last_execution 
        FROM time_triggers 
        WHERE start_time <= NOW() 
        AND (end_time IS NULL OR end_time > NOW())`

	rows, err := p.pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var triggers []types.TimeTrigger
	for rows.Next() {
		var t types.TimeTrigger
		err := rows.Scan(
			&t.PolicyID,
			&t.CronExpression,
			&t.StartTime,
			&t.EndTime,
			&t.Frequency,
			&t.Interval,
			&t.LastExecution)
		if err != nil {
			return nil, err
		}
		triggers = append(triggers, t)
	}

	return triggers, nil
}

func (p *PostgresBackend) UpdateTriggerExecution(policyID string) error {
	if p.pool == nil {
		return fmt.Errorf("database pool is nil")
	}

	query := `
        UPDATE time_triggers 
        SET last_execution = $2
        WHERE policy_id = $1`

	_, err := p.pool.Exec(context.Background(), query, policyID, time.Now().UTC())
	return err
}

func (p *PostgresBackend) UpdateTriggerTx(ctx context.Context, policyID string, trigger types.TimeTrigger, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `
        UPDATE time_triggers 
        SET start_time = $2,
            frequency = $3,
            interval = $4,
            cron_expression = $5
        WHERE policy_id = $1`,
		policyID,
		trigger.StartTime,
		trigger.Frequency,
		trigger.Interval,
		trigger.CronExpression,
	)
	return err
}
