package scheduler

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
)

type Service struct {
	client *asynq.Client
	logger *logrus.Logger
}

func NewService(client *asynq.Client, logger *logrus.Logger) *Service {
	return &Service{
		client: client,
		logger: logger,
	}
}

func (s *Service) SchedulePluginSign(payload ScheduledPluginSignPayload) error {
	taskBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	_, err = s.client.Enqueue(
		asynq.NewTask(TypeScheduledPluginSign, taskBytes),
		asynq.ProcessAt(payload.Schedule.StartTime),
		asynq.Queue("scheduled_plugin_queue"),
	)
	return err
}
