package scheduler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hibiken/asynq"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/vultisigner/internal/request"
	"github.com/vultisig/vultisigner/internal/tasks"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/storage/postgres"
)

type SchedulerService struct {
	db        *postgres.PostgresBackend
	logger    *logrus.Logger
	client    *asynq.Client
	inspector *asynq.Inspector
	done      chan struct{}
}

func NewSchedulerService(db *postgres.PostgresBackend, logger *logrus.Logger, client *asynq.Client) *SchedulerService {
	if db == nil {
		logger.Fatal("database connection is nil")
	}

	// create inspector using the same Redis configuration as the client
	inspector := asynq.NewInspector(asynq.RedisClientOpt{
		Addr: "localhost:6379", // redis configuration
		DB:   1,                // redis DB number
	})

	return &SchedulerService{
		db:        db,
		logger:    logger,
		client:    client,
		inspector: inspector,
		done:      make(chan struct{}),
	}
}

func (s *SchedulerService) Start() {
	go s.run()
}

func (s *SchedulerService) Stop() {
	close(s.done)
}

func (s *SchedulerService) run() {
	//ticker := time.NewTicker(1 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.logger.Info("Checking and enqueuing tasks")
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
	s.logger.Info("Triggers: ", triggers)

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

		s.logger.WithFields(logrus.Fields{
			"current_time": time.Now(),
			"next_time":    nextTime,
			"policy_id":    trigger.PolicyID,
			"last_exec":    trigger.LastExecution,
		}).Info("Checking execution time")
		if time.Now().After(nextTime) {
			// Get policy details
			policy, err := s.db.GetPluginPolicy(trigger.PolicyID)
			if err != nil {
				s.logger.Errorf("Failed to get policy: %v", err)
				continue
			}

			s.logger.WithFields(logrus.Fields{
				"policy_id":   policy.ID,
				"public_key":  policy.PublicKey,
				"plugin_type": policy.PluginType,
			}).Info("Retrieved policy for signing")

			signRequest, err := request.CreateSigningRequest(policy)
			if err != nil {
				s.logger.Errorf("Failed to create signing request: %v", err)
				continue
			}
			for _, signRequest := range signRequest {

				signBytes, err := json.Marshal(signRequest)
				if err != nil {
					s.logger.Errorf("Failed to marshal sign request: %v", err)
					continue
				}

				signResp, err := http.Post(
					fmt.Sprintf("http://localhost:%d/signFromPlugin", 8080),
					"application/json",
					bytes.NewBuffer(signBytes),
				)
				if err != nil {
					s.logger.Errorf("Failed to make sign request: %v", err)
					return err
				}
				defer signResp.Body.Close()

				// Read and log response
				respBody, err := io.ReadAll(signResp.Body)
				if err != nil {
					s.logger.Errorf("Failed to read response: %v", err)
					return err
				}

				if signResp.StatusCode == http.StatusOK {
					// Enqueue the same signing request locally
					signRequest.KeysignRequest.StartSession = true
					signRequest.KeysignRequest.Parties = []string{"1", "2"}
					buf, err := json.Marshal(signRequest.KeysignRequest)
					if err != nil {
						s.logger.Errorf("Failed to marshal local sign request: %v", err)
						return err
					}

					// Enqueue TypeKeySign directly
					ti, err := s.client.Enqueue(
						asynq.NewTask(tasks.TypeKeySign, buf),
						asynq.MaxRetry(-1),
						asynq.Timeout(2*time.Minute),
						asynq.Retention(5*time.Minute),
						asynq.Queue(tasks.QUEUE_NAME),
					)
					if err != nil {
						s.logger.Errorf("Failed to enqueue signing task: %v", err)
						continue
					}

					// wait for result with timeout
					result, err := s.waitForTaskResult(ti.ID, 120*time.Second) // adjust timeout as needed
					if err != nil {                                            //do we consider that the signature is always valid if err = nil?
						s.logger.Errorf("Failed to get task result: %v", err)
						return err
					}
					//do we store the result in db? or boradcast it directly?

					s.logger.WithFields(logrus.Fields{
						"task_id": ti.ID,
						"result":  string(result),
					}).Info("Successfully retrieved task result")

					/*if result.isValid      {

					}*/
					if err := s.db.UpdateTriggerExecution(trigger.PolicyID); err != nil {
						s.logger.Errorf("Failed to update last execution: %v", err)
					}
					s.logger.Infof("Local signing task enqueued with ID: %s", ti.ID)
				}
				s.logger.Infof("Plugin signing test complete. Status: %d, Response: %s",
					signResp.StatusCode, string(respBody))
			}

			//only when valid signture is stored, update last execution in trigger.
			//then, other part will be responsible for broadcasting the signature to the network. and retry if needed

		}
	}

	return nil
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

func (s *SchedulerService) waitForTaskResult(taskID string, timeout time.Duration) ([]byte, error) {
	start := time.Now()
	pollInterval := time.Second // Poll every second

	for {
		// check if we've exceeded timeout
		if time.Since(start) > timeout {
			return nil, fmt.Errorf("timeout waiting for task result after %v", timeout)
		}

		// try to get the result
		task, err := s.inspector.GetTaskInfo(tasks.QUEUE_NAME, taskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get task info: %w", err)
		}

		switch task.State {
		case asynq.TaskStateCompleted:
			s.logger.Info("Task completed successfully")
			return task.Result, nil
		case asynq.TaskStateArchived:
			return nil, fmt.Errorf("task archived: %s", task.LastErr)
		case asynq.TaskStateRetry:
			s.logger.Debug("Task scheduled for retry...")
			time.Sleep(pollInterval)
			continue
		case asynq.TaskStatePending, asynq.TaskStateActive, asynq.TaskStateScheduled:
			s.logger.Debug("Task still in progress, waiting...")
			time.Sleep(pollInterval)
			continue
		case asynq.TaskStateAggregating:
			s.logger.Debug("Task aggregating, waiting...")
			time.Sleep(pollInterval)
			continue
		default:
			return nil, fmt.Errorf("unexpected task state: %s", task.State)
		}
	}
}
