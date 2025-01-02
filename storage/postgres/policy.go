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
