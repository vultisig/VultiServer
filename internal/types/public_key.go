package types

type GetDerivedPublicKeyRequest struct {
	PublicKey    string `json:"public_key"`
	HexChainCode string `json:"hex_chain_code"`
	DerivePath   string `json:"derive_path"`
	IsEdDSA      bool   `json:"is_eddsa"`
}
