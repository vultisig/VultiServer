package tasks

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

func NewKeyGeneration(
	localKey string,
	sessionID string,
	chainCode string,
	parties []string,
) (*asynq.Task, error) {
	payload, err := json.Marshal(KeyGenerationPayload{LocalKey: localKey, SessionID: sessionID, ChainCode: chainCode, Parties: parties})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeKeyGeneration, payload), nil
}
