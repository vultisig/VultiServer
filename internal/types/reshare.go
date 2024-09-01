package types

import (
	"fmt"
)

// ReshareRequest is a struct that represents a request to reshare a vault
type ReshareRequest struct {
	Name               string `json:"name"`                // name of the vault
	PublicKey          string `json:"public_key"`          // public key ecdsa
	SessionID          string `json:"session_id"`          // session id
	HexEncryptionKey   string `json:"hex_encryption_key"`  // hex encryption key
	HexChainCode       string `json:"hex_chain_code"`      // hex chain code
	LocalPartyId       string `json:"local_party_id"`      // local party id
	EncryptionPassword string `json:"encryption_password"` // password used to encrypt the vault file
	Email              string `json:"email"`
}

func (req *ReshareRequest) IsValid() error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.SessionID == "" {
		return fmt.Errorf("session_id is required")
	}
	if req.HexEncryptionKey == "" {
		return fmt.Errorf("hex_encryption_key is required")
	}
	if req.HexChainCode == "" {
		return fmt.Errorf("hex_chain_code is required")
	}
	if req.EncryptionPassword == "" {
		return fmt.Errorf("encryption_password is required")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	return nil
}
