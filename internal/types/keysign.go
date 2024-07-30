package types

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hibiken/asynq"

	"github.com/vultisig/vultisigner/internal/tasks"
)

type KeysignRequest struct {
	PublicKey        string   `json:"public_key"`         // public key, used to identify the backup file
	Messages         []string `json:"messages"`           // Messages need to be signed
	Session          string   `json:"session"`            // Session ID , it should be an UUID
	HexEncryptionKey string   `json:"hex_encryption_key"` // Hex encryption key, used to encrypt the keysign messages
	DerivePath       string   `json:"derive_path"`        // Derive Path
	IsECDSA          bool     `json:"is_ecdsa"`           // indicate use ECDSA or EDDSA key to sign the messages
	VaultPassword    string   `json:"vault_password"`     // password used to decrypt the vault file
}

// IsValid checks if the keysign request is valid
func (r KeysignRequest) IsValid() error {
	if r.PublicKey == "" {
		return errors.New("invalid public key ECDSA")
	}
	if len(r.Messages) == 0 {
		return errors.New("invalid messages")
	}
	if r.Session == "" {
		return errors.New("invalid session")
	}
	if r.HexEncryptionKey == "" {
		return errors.New("invalid hex encryption key")
	}
	if r.DerivePath == "" {
		return errors.New("invalid derive path")
	}

	return nil
}

// NewKeysignTask creates a new task to sign the messages
func (r KeysignRequest) NewKeysignTask(vaultPassword string) (*asynq.Task, error) {
	buf, err := json.Marshal(tasks.KeysignPayload{
		PublicKey:        r.PublicKey,
		Messages:         r.Messages,
		SessionID:        r.Session,
		HexEncryptionKey: r.HexEncryptionKey,
		DerivePath:       r.DerivePath,
		IsECDSA:          r.IsECDSA,
		VaultPassword:    vaultPassword,
	})
	if err != nil {
		return nil, fmt.Errorf("fail to marshal keysign payload to json,err: %w", err)
	}
	return asynq.NewTask(tasks.TypeKeySign, buf), nil

}
