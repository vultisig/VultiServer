package types

type KeyGeneration struct {
	Key string `json:"key" validate:"required"` // should always be vultisigner
	// Parties   []string `json:"parties" validate:"required,dive,required"`
	Session   string `json:"session" validate:"required"`
	ChainCode string `json:"chain_code" validate:"required"`
}
