package types

import "time"

type TimeTrigger struct {
	PolicyID       string
	CronExpression string
	StartTime      time.Time
	EndTime        *time.Time
	Frequency      string
	Interval       int
	LastExecution  *time.Time
}
