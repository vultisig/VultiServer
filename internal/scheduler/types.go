package scheduler

import (
	"time"

	"github.com/vultisig/vultisigner/internal/types"
)

type Schedule struct {
	StartTime     time.Time  `json:"start_time"`
	EndTime       *time.Time `json:"end_time,omitempty"`
	Frequency     string     `json:"frequency,omitempty"` // "once", "hourly", "daily", "weekly", "monthly"
	CronExpr      string     `json:"cron_expr,omitempty"`
	ExecutionTime time.Time  `json:"execution_time,omitempty"`
}

type ScheduledPluginSignPayload struct {
	PolicyID    string                     `json:"policy_id"`
	SignRequest types.PluginKeysignRequest `json:"sign_request"`
	Schedule    Schedule                   `json:"schedule"`
}
