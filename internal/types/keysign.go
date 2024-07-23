package types

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/vultisig/mobile-tss-lib/tss"

	"github.com/vultisig/vultisigner/internal/tasks"
)

type KeysignRequest struct {
	PublicKeyECDSA   string   `json:"public_key_ecdsa"`   // ECDSA public key, used to identify the backup file
	Messages         []string `json:"messages"`           // Messages need to be signed
	Key              string   `json:"key"`                // should always be Vultisigner
	Session          string   `json:"session"`            // Session ID , it should be an UUID
	HexEncryptionKey string   `json:"hex_encryption_key"` // Hex encryption key, used to encrypt the keysign messages
	DerivePath       string   `json:"derive_path"`        // Derive Path
	IsECDSA          bool     `json:"is_ecdsa"`           // indicate use ECDSA or EDDSA key to sign the messages
}

// IsValid checks if the keysign request is valid
func (r KeysignRequest) IsValid() error {
	if r.PublicKeyECDSA == "" {
		return errors.New("invalid public key ECDSA")
	}
	if len(r.Messages) == 0 {
		return errors.New("invalid messages")
	}
	if r.Key == "" {
		return errors.New("invalid key")
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
		PublicKeyECDSA:   r.PublicKeyECDSA,
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

type KeysignResponse struct {
	Signatures []tss.KeysignResponse `json:"signature"` // Signature of the messages
}
