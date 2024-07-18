package tasks

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

func NewKeyGeneration(
	localKey string,
	name string,
	sessionID string,
	chainCode string,
	hexEncryptionKey string,
) (*asynq.Task, error) {
	payload, err := json.Marshal(KeyGenerationPayload{LocalKey: localKey, Name: name, SessionID: sessionID, ChainCode: chainCode, HexEncryptionKey: hexEncryptionKey})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeKeyGeneration, payload), nil
}
