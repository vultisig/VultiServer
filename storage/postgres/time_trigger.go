package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/vultisig/vultisigner/internal/types"
)

func (p *PostgresBackend) CreateTimeTriggerTx(ctx context.Context, tx pgx.Tx, trigger types.TimeTrigger) error {
	if p.pool == nil {
		return fmt.Errorf("database pool is nil")
	}

	query := `
		INSERT INTO time_triggers 
    (policy_id, cron_expression, start_time, end_time, frequency, interval) 
    VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := tx.Exec(ctx, query,
		trigger.PolicyID,
		trigger.CronExpression,
		trigger.StartTime,
		trigger.EndTime,
		trigger.Frequency,
		trigger.Interval,
	)
	return err
}

func (p *PostgresBackend) GetPendingTimeTriggers(ctx context.Context) ([]types.TimeTrigger, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("database pool is nil")
	}

	// TODO: add limit and proper index
	query := `
  	WITH active_triggers AS (
    		SELECT t.policy_id, t.cron_expression, t.start_time, t.end_time, t.frequency, t.interval, t.last_execution 
				FROM time_triggers t
				INNER JOIN plugin_policies p ON t.policy_id = p.id
				WHERE t.start_time <= $1
				AND (t.end_time IS NULL OR t.end_time > $1)
				AND p.active = true
				AND (t.last_execution IS NULL OR t.last_execution < $1)
    )
    SELECT * FROM active_triggers
    ORDER BY start_time ASC
	`

	rows, err := p.pool.Query(ctx, query, time.Now().UTC())
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

func (p *PostgresBackend) UpdateTimeTriggerLastExecution(ctx context.Context, policyID string) error {
	if p.pool == nil {
		return fmt.Errorf("database pool is nil")
	}

	query := `
		UPDATE time_triggers 
		SET last_execution = $2
		WHERE policy_id = $1
	`

	_, err := p.pool.Exec(ctx, query, policyID, time.Now().UTC())
	return err
}

func (p *PostgresBackend) UpdateTimeTriggerTx(ctx context.Context, policyID string, trigger types.TimeTrigger, tx pgx.Tx) error {
	if p.pool == nil {
		return fmt.Errorf("database pool is nil")
	}

	query := `
		UPDATE time_triggers 
		SET start_time = $2,
				frequency = $3,
				interval = $4,
				cron_expression = $5
		WHERE policy_id = $1
	`

	_, err := tx.Exec(ctx, query,
		policyID,
		trigger.StartTime,
		trigger.Frequency,
		trigger.Interval,
		trigger.CronExpression,
	)
	return err
}
