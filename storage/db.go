package storage

import "github.com/vultisig/vultisigner/internal/types"

type DatabaseStorage interface {
	Close() error

	InsertPluginPolicy(policyDoc types.PluginPolicy) error
	GetPluginPolicy(id string) (types.PluginPolicy, error)

	CreateTimeTrigger(trigger types.TimeTrigger) error
	GetPendingTriggers() ([]types.TimeTrigger, error)
	UpdateTriggerExecution(policyID string) error
}
