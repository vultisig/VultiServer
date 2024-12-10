package scheduler

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/storage/postgres"
)

type SchedulerService struct {
	db     *postgres.PostgresBackend
	logger *logrus.Logger
	client *asynq.Client
	done   chan struct{}
}

func NewSchedulerService(db *postgres.PostgresBackend, logger *logrus.Logger, client *asynq.Client) *SchedulerService {
	if db == nil {
		logger.Fatal("database connection is nil")
	}
	return &SchedulerService{
		db:     db,
		logger: logger,
		client: client,
		done:   make(chan struct{}),
	}
}

func (s *SchedulerService) Start() {
	go s.run()
}

func (s *SchedulerService) Stop() {
	close(s.done)
}

func (s *SchedulerService) run() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.checkAndEnqueueTasks(); err != nil {
				s.logger.Errorf("Failed to check and enqueue tasks: %v", err)
			}
		case <-s.done:
			return
		}
	}
}

func (s *SchedulerService) checkAndEnqueueTasks() error {
	triggers, err := s.db.GetPendingTriggers()
	if err != nil {
		return fmt.Errorf("failed to get pending triggers: %w", err)
	}

	for _, trigger := range triggers {
		// Parse cron expression
		schedule, err := cron.ParseStandard(trigger.CronExpression)
		if err != nil {
			s.logger.Errorf("Failed to parse cron expression: %v", err)
			continue
		}

		// Check if it's time to execute
		var nextTime time.Time
		if trigger.LastExecution != nil {
			nextTime = schedule.Next(*trigger.LastExecution)
		} else {
			nextTime = schedule.Next(time.Now().Add(-24 * time.Hour))
		}

		if time.Now().After(nextTime) {
			// Time to execute this policy
			if err := s.enqueuePolicyExecution(trigger.PolicyID); err != nil {
				s.logger.Errorf("Failed to enqueue policy execution: %v", err)
				continue
			}

			// Update last_execution
			if err := s.db.UpdateTriggerExecution(trigger.PolicyID); err != nil {
				s.logger.Errorf("Failed to update last execution: %v", err)
			}
		}
	}

	return nil
}

func (s *SchedulerService) enqueuePolicyExecution(policyID string) error {
	payload := ScheduledPluginSignPayload{
		PolicyID: policyID,
	}

	taskBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	_, err = s.client.Enqueue(
		asynq.NewTask(TypeScheduledPluginSign, taskBytes),
		asynq.Queue("scheduled_plugin_queue"),
	)
	return err
}

func (s *SchedulerService) CreateTimeTrigger(policy types.PluginPolicy) error {
	if s.db == nil {
		return fmt.Errorf("database backend is nil")
	}

	s.logger.Info("Attempting to parse policy schedule")

	var policySchedule struct {
		Schedule struct {
			Frequency string     `json:"frequency"`
			StartTime time.Time  `json:"start_time"`
			EndTime   *time.Time `json:"end_time,omitempty"`
		} `json:"schedule"`
	}

	if err := json.Unmarshal(policy.Policy, &policySchedule); err != nil {
		return fmt.Errorf("failed to parse policy schedule: %w", err)
	}

	s.logger.Info("Frequency to cron")

	cronExpr := frequencyToCron(policySchedule.Schedule.Frequency, policySchedule.Schedule.StartTime)

	trigger := postgres.TimeTrigger{
		PolicyID:       policy.ID,
		CronExpression: cronExpr,
		StartTime:      policySchedule.Schedule.StartTime,
		EndTime:        policySchedule.Schedule.EndTime,
		Frequency:      policySchedule.Schedule.Frequency,
	}

	return s.db.CreateTimeTrigger(trigger)
}

func frequencyToCron(frequency string, startTime time.Time) string {
	switch frequency {
	case "hourly":
		return fmt.Sprintf("%d * * * *", startTime.Minute())
	case "daily":
		return fmt.Sprintf("%d %d * * *", startTime.Minute(), startTime.Hour())
	case "weekly":
		return fmt.Sprintf("%d %d * * %d", startTime.Minute(), startTime.Hour(), startTime.Weekday())
	case "monthly":
		return fmt.Sprintf("%d %d %d * *", startTime.Minute(), startTime.Hour(), startTime.Day())
	default:
		return ""
	}
}
