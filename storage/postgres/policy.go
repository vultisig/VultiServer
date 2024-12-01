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
