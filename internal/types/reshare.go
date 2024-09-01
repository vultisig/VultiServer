package types

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
