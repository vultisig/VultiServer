package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vultisig/vultisigner/internal/types"
)

func (d *PostgresBackend) InsertPluginPolicy(policyDoc types.PluginPolicy) error {
	policyJSON, err := json.Marshal(policyDoc.Policy)
	if err != nil {
		return fmt.Errorf("failed to marshal policy: %w", err)
	}

	_, err = d.pool.Exec(context.Background(), "INSERT INTO plugin_policies (id, public_key, plugin_id, plugin_version, policy_version, plugin_type, signature, policy) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", policyDoc.ID, policyDoc.PublicKey, policyDoc.PluginID, policyDoc.PluginVersion, policyDoc.PolicyVersion, policyDoc.PluginType, policyDoc.Signature, policyJSON)

	return err
}

func (d *PostgresBackend) UpdatePluginPolicy(policyDoc types.PluginPolicy) error {
	policyJSON, err := json.Marshal(policyDoc.Policy)
	if err != nil {
		return fmt.Errorf("failed to marshal policy: %w", err)
	}

	_, err = d.pool.Exec(context.Background(), "UPDATE plugin_policies SET (id, public_key, plugin_id, plugin_version, policy_version, plugin_type, signature, policy) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) WHERE id = $1", policyDoc.ID, policyDoc.PublicKey, policyDoc.PluginID, policyDoc.PluginVersion, policyDoc.PolicyVersion, policyDoc.PluginType, policyDoc.Signature, policyJSON)

	return err
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

func (p *PostgresBackend) GetAllPluginPolicies(public_key string) ([]types.PluginPolicy, error) {
	if p.pool == nil {
		return []types.PluginPolicy{}, fmt.Errorf("database pool is nil")
	}

	query := `
        SELECT id, public_key, plugin_id, plugin_version, policy_version, plugin_type, signature, policy 
        FROM plugin_policies
		WHERE public_key = $1`

	rows, err := p.pool.Query(context.Background(), query, public_key)
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
