package types

type VaultResendRequest struct {
	PublicKeyECDSA string `json:"public_key_ecdsa"`
	Password       string `json:"password"`
	Email          string `json:"email"`
}
