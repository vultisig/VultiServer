package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/vultisigner/internal/tasks"
)

const (
	TypeScheduledPluginSign = "scheduled:plugin:sign"
)

type TaskHandler struct {
	client *asynq.Client
	logger *logrus.Logger
	queues map[string]int
}

func NewTaskHandler(client *asynq.Client, logger *logrus.Logger) *TaskHandler {
	return &TaskHandler{
		client: client,
		logger: logger,
		queues: map[string]int{
			tasks.QUEUE_NAME:         10,
			tasks.EMAIL_QUEUE_NAME:   100,
			"scheduled_plugin_queue": 10,
		},
	}
}

func (h *TaskHandler) HandleScheduledPluginSign(ctx context.Context, t *asynq.Task) error {
	var payload ScheduledPluginSignPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// check time
	now := time.Now()
	if now.Before(payload.Schedule.StartTime) {
		return fmt.Errorf("too early to execute, will retry at %v", payload.Schedule.StartTime)
	}

	//create sign request
	signBytes, err := json.Marshal(payload.SignRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal sign request: %w", err)
	}

	// execute it
	_, err = h.client.Enqueue(
		asynq.NewTask(tasks.TypeKeySign, signBytes),
		asynq.MaxRetry(-1),
		asynq.Timeout(2*time.Minute),
		asynq.Retention(5*time.Minute),
		asynq.Queue(tasks.QUEUE_NAME),
	)
	if err != nil {
		return fmt.Errorf("failed to enqueue signing task: %w", err)
	}

	// if recurring, schedule next execution
	if payload.Schedule.Frequency != "once" {
		nextTime := h.calculateNextExecution(payload.Schedule)
		if nextTime != nil {
			newPayload := payload
			newPayload.Schedule.StartTime = *nextTime

			taskBytes, err := json.Marshal(newPayload)
			if err != nil {
				return fmt.Errorf("failed to marshal next task: %w", err)
			}

			_, err = h.client.Enqueue(
				asynq.NewTask(TypeScheduledPluginSign, taskBytes),
				asynq.ProcessAt(*nextTime),
				asynq.Queue("scheduled_plugin_queue"),
			)
			if err != nil {
				h.logger.Errorf("Failed to schedule next task: %v", err)
			} else {
				h.logger.Infof("Scheduled next execution for %v", nextTime)
			}
		}
	}

	return nil
}

func (h *TaskHandler) calculateNextExecution(schedule Schedule) *time.Time {
	now := time.Now()
	var next time.Time

	if schedule.CronExpr != "" {
		// TODO: Implement cron expression parsing
		return nil
	}

	switch schedule.Frequency {
	case "hourly":
		next = now.Add(time.Hour)
	case "daily":
		next = now.AddDate(0, 0, 1)
		next = time.Date(next.Year(), next.Month(), next.Day(),
			schedule.StartTime.Hour(), schedule.StartTime.Minute(), 0, 0, next.Location())
	case "weekly":
		next = now.AddDate(0, 0, 7)
		next = time.Date(next.Year(), next.Month(), next.Day(),
			schedule.StartTime.Hour(), schedule.StartTime.Minute(), 0, 0, next.Location())
	case "monthly":
		next = now.AddDate(0, 1, 0)
		next = time.Date(next.Year(), next.Month(), schedule.StartTime.Day(),
			schedule.StartTime.Hour(), schedule.StartTime.Minute(), 0, 0, next.Location())
	case "once":
		return nil
	default:
		h.logger.Errorf("Unknown frequency: %s", schedule.Frequency)
		return nil
	}

	if schedule.EndTime != nil && next.After(*schedule.EndTime) {
		return nil
	}

	return &next
}
