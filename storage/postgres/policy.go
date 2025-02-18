package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vultisig/vultisigner/internal/types"
)

func (p *PostgresBackend) InsertPluginPolicy(policyDoc types.PluginPolicy) (types.PluginPolicy, error) {
	if p.pool == nil {
		return types.PluginPolicy{}, fmt.Errorf("database pool is nil")
	}
	policyJSON, err := json.Marshal(policyDoc.Policy)
	if err != nil {
		return types.PluginPolicy{}, fmt.Errorf("failed to marshal policy: %w", err)
	}

	query := `INSERT INTO plugin_policies
	(id, public_key, plugin_id, plugin_version, policy_version, plugin_type, signature, policy)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING *
	`

	var policy types.PluginPolicy
	err = p.pool.QueryRow(context.Background(), query, policyDoc.ID, policyDoc.PublicKey, policyDoc.PluginID, policyDoc.PluginVersion, policyDoc.PolicyVersion, policyDoc.PluginType, policyDoc.Signature, policyJSON).Scan(
		&policy.ID,
		&policy.PublicKey,
		&policy.PluginID,
		&policy.PluginVersion,
		&policy.PolicyVersion,
		&policy.PluginType,
		&policy.Signature,
		&policy.Policy,
	)

	if err != nil {
		return types.PluginPolicy{}, err
	}

	return policy, nil
}

func (p *PostgresBackend) UpdatePluginPolicy(policyDoc types.PluginPolicy) (types.PluginPolicy, error) {
	if p.pool == nil {
		return types.PluginPolicy{}, fmt.Errorf("database pool is nil")
	}

	policyJSON, err := json.Marshal(policyDoc.Policy)
	if err != nil {
		return types.PluginPolicy{}, fmt.Errorf("failed to marshal policy: %w", err)
	}

	query := `UPDATE plugin_policies 
	SET public_key = $2, plugin_type = $3, signature = $4, policy = $5
	WHERE id = $1
	RETURNING *
	`

	var policy types.PluginPolicy
	err = p.pool.QueryRow(context.Background(), query, policyDoc.ID, policyDoc.PublicKey, policyDoc.PluginType, policyDoc.Signature, policyJSON).Scan(
		&policy.ID,
		&policy.PublicKey,
		&policy.PluginID,
		&policy.PluginVersion,
		&policy.PolicyVersion,
		&policy.PluginType,
		&policy.Signature,
		&policy.Policy,
	)
	if err != nil {
		return types.PluginPolicy{}, err
	}

	return policy, nil
}

func (p *PostgresBackend) DeletePluginPolicy(id string) error {
	if p.pool == nil {
		return fmt.Errorf("database pool is nil")
	}

	// TODO: pass from outside
	ctx := context.Background()

	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin db transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		DELETE FROM transaction_history 
		WHERE policy_id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("failed to delete transaction history: %w", err)
	}

	_, err = tx.Exec(ctx, `
		DELETE FROM time_triggers
		WHERE policy_id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("failed to delete time triggers: %w", err)
	}

	_, err = tx.Exec(ctx, `
		DELETE FROM plugin_policies 
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit db transaction: %w", err)
	}

	return nil
}

func (p *PostgresBackend) GetPluginPolicy(id string) (types.PluginPolicy, error) {
	if p.pool == nil {
		return types.PluginPolicy{}, fmt.Errorf("database pool is nil")
	}

	var policy types.PluginPolicy
	var policyJSON []byte

	query := `
        SELECT id, public_key, plugin_id, plugin_version, policy_version, plugin_type, signature, policy 
        FROM plugin_policies 
        WHERE id = $1`

	err := p.pool.QueryRow(context.Background(), query, id).Scan(
		&policy.ID,
		&policy.PublicKey,
		&policy.PluginID,
		&policy.PluginVersion,
		&policy.PolicyVersion,
		&policy.PluginType,
		&policy.Signature,
		&policyJSON,
	)

	if err != nil {
		return types.PluginPolicy{}, fmt.Errorf("failed to get policy: %w", err)
	}

	policy.Policy = json.RawMessage(policyJSON)
	return policy, nil
}

func (p *PostgresBackend) GetAllPluginPolicies(publicKey string, pluginType string) ([]types.PluginPolicy, error) {
	if p.pool == nil {
		return []types.PluginPolicy{}, fmt.Errorf("database pool is nil")
	}

	query := `
        SELECT id, public_key, plugin_id, plugin_version, policy_version, plugin_type, signature, policy 
        FROM plugin_policies
		WHERE public_key = $1
		AND plugin_type = $2`

	rows, err := p.pool.Query(context.Background(), query, publicKey, pluginType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var policies []types.PluginPolicy
	for rows.Next() {
		var p types.PluginPolicy
		err := rows.Scan(
			&p.ID,
			&p.PublicKey,
			&p.PluginID,
			&p.PluginVersion,
			&p.PolicyVersion,
			&p.PluginType,
			&p.Signature,
			&p.Policy,
		)
		if err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}

	return policies, nil
}
