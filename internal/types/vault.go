package types

type VaultCreateRequest struct {
	Name      string `json:"name" validate:"required"`
	ChainCode string `json:"chain_code" validate:"required"`
	// Key       string   `json:"key" validate:"required"` // this is the public key of the vault
	// Parties   []string `json:"parties" validate:"required,dive,required"`
	// Session   string   `json:"session" validate:"required"`
}

type Vault struct {
	Name string `json:"name" validate:"required"`
	// ChainCode string `json:"chain_code" validate:"required"`
	// Key       string   `json:"key" validate:"required"` // this is the public key of the vault
	// Parties   []string `json:"parties" validate:"required,dive,required"`
	// Session   string   `json:"session" validate:"required"`
}
