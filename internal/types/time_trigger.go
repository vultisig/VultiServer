package types

import "time"

type TimeTrigger struct {
	PolicyID       string     `json:"policy_id"`
	CronExpression string     `json:"cron_expression"`
	StartTime      time.Time  `json:"start_time"`
	EndTime        *time.Time `json:"end_time"`
	Frequency      string     `json:"frequency"`
	Interval       int        `json:"interval"`
	LastExecution  *time.Time `json:"last_execution"`
	Status         string     `json:"status"`
}
