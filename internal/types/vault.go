package types

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"

	"github.com/vultisig/vultisigner/internal/tasks"
)

// VaultCreateRequest is a struct that represents a request to create a new vault from integration.
type VaultCreateRequest struct {
	Name               string `json:"name" validate:"required"`
	EncryptionPassword string `json:"encryption_password" validate:"required"`
}

// VaultCreateResponse is a struct that represents a response to create a new vault
// integration partner need to use this information to construct a QR Code , so vultisig device can participate in the vault creation process.
type VaultCreateResponse struct {
	Name             string `json:"name"`
	SessionID        string `json:"session_id"`
	HexEncryptionKey string `json:"hex_encryption_key"`
	HexChainCode     string `json:"hex_chain_code"`
}

func (v *VaultCreateResponse) Task() (*asynq.Task, error) {
	payload, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(tasks.TypeKeyGeneration, payload), nil
}

// VaultCacheItem is a struct that represents the vault information stored in cache
type VaultCacheItem struct {
	Name               string `json:"name"`
	SessionID          string `json:"session_id"`
	HexEncryptionKey   string `json:"hex_encryption_key"`
	HexChainCode       string `json:"hex_chain_code"`
	EncryptionPassword string `json:"encryption_password"` // this is the password used to encrypt the vault file
}

func (v VaultCacheItem) Key() string {
	return fmt.Sprintf("vault-%s-%s", v.Name, v.SessionID)
}
