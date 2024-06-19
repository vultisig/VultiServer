package types

type KeyGeneration struct {
	Key       string `json:"key" validate:"required"` // should always be vultisigner
	Session   string `json:"session" validate:"required"`
	ChainCode string `json:"chain_code" validate:"required"`
}
