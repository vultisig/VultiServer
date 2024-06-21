package types

type KeyGeneration struct {
	Key              string `json:"key" validate:"required"` // should always be Vultisigner
	Session          string `json:"session" validate:"required"`
	ChainCode        string `json:"chain_code" validate:"required"`
	HexEncryptionKey string `json:"hex_encryption_key" validate:"required"`
}
