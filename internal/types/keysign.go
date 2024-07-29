package types

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hibiken/asynq"

	keysignTypes "github.com/vultisig/commondata/go/vultisig/keysign/v1"
	"github.com/vultisig/vultisigner/internal/tasks"
)

type KeysignAPIRequest struct {
	PublicKey  string                       `json:"public_key"`  // public key, used to identify the backup file
	Messages   []string                     `json:"messages"`    // Messages need to be signed
	DerivePath string                       `json:"derive_path"` // Derive Path
	IsECDSA    bool                         `json:"is_ecdsa"`    // indicate use ECDSA or EDDSA key to sign the messages
	Payload    *keysignTypes.KeysignPayload `json:"payload"`     // keysign payload
}

// IsValid checks if the keysign request is valid
func (r KeysignAPIRequest) IsValid() error {
	if r.PublicKey == "" {
		return errors.New("invalid public key ECDSA")
	}
	if len(r.Messages) == 0 {
		return errors.New("invalid messages")
	}
	if r.DerivePath == "" {
		return errors.New("invalid derive path")
	}
	if r.Payload == nil {
		return errors.New("invalid payload")
	}
	return nil
}

// NewKeysignTask creates a new task to sign the messages
func (r KeysignAPIRequest) NewKeysignTask(vaultPassword, sessionID, encryptionKey string) (*asynq.Task, error) {
	buf, err := json.Marshal(tasks.KeysignPayload{
		PublicKey:        r.PublicKey,
		Messages:         r.Messages,
		SessionID:        sessionID,
		HexEncryptionKey: encryptionKey,
		DerivePath:       r.DerivePath,
		IsECDSA:          r.IsECDSA,
		VaultPassword:    vaultPassword,
	})
	if err != nil {
		return nil, fmt.Errorf("fail to marshal keysign payload to json,err: %w", err)
	}
	return asynq.NewTask(tasks.TypeKeySign, buf), nil

}

type KeysignAPIResponse struct {
	SessionID        string `json:"session_id"`
	HexEncryptionKey string `json:"hex_encryption_key"`
	HexChainCode     string `json:"hex_chain_code"`
	KeysignMsg       string `json:"keysign_msg"`
	TaskId           string `json:"task_id"`
}

type KeysignRequest struct {
	PublicKey        string                      `json:"public_key"`         // ECDSA public key, used to identify the backup file
	Messages         []string                    `json:"messages"`           // Messages need to be signed
	Session          string                      `json:"session"`            // Session ID , it should be an UUID
	HexEncryptionKey string                      `json:"hex_encryption_key"` // Hex encryption key, used to encrypt the keysign messages
	DerivePath       string                      `json:"derive_path"`        // Derive Path
	IsECDSA          bool                        `json:"is_ecdsa"`           // indicate use ECDSA or EDDSA key to sign the messages
	VaultPassword    string                      `json:"vault_password"`     // password used to decrypt the vault file
	Payload          keysignTypes.KeysignPayload `json:"payload"`            // keysign payload
}
