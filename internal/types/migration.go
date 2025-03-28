package types

import (
	"fmt"

	"github.com/google/uuid"
)

// MigrationRequest is a struct that represents a request to reshare a vault
type MigrationRequest struct {
	PublicKey          string `json:"public_key"`          // public key ecdsa
	SessionID          string `json:"session_id"`          // session id
	HexEncryptionKey   string `json:"hex_encryption_key"`  // hex encryption key
	EncryptionPassword string `json:"encryption_password"` // password used to encrypt the vault file
	Email              string `json:"email"`
}

func (req *MigrationRequest) IsValid() error {
	if req.SessionID == "" {
		return fmt.Errorf("session_id is required")
	}
	if _, err := uuid.Parse(req.SessionID); err != nil {
		return fmt.Errorf("session_id is not valid")
	}
	if req.HexEncryptionKey == "" {
		return fmt.Errorf("hex_encryption_key is required")
	}
	if !isValidHexString(req.HexEncryptionKey) {
		return fmt.Errorf("hex_encryption_key is not valid")
	}
	if req.EncryptionPassword == "" {
		return fmt.Errorf("encryption_password is required")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	return nil
}
