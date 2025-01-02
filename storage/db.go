package storage

import "github.com/vultisig/vultisigner/internal/types"

type DatabaseStorage interface {
	Close() error

	InsertPluginPolicy(policyDoc types.PluginPolicy) error
	CreateTimeTrigger(trigger types.TimeTrigger) error
}
