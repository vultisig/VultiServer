package types

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
