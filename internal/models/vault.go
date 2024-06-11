package models

// KeyShare is a struct that represents a keyshare.
type KeyShare struct {
	PubKey   string `json:"pubkey"`
	Keyshare string `json:"keyshare"`
}

// Vault is a struct that represents a vault. It is used to store the information about the vault.
// Here keep the json field name consistent with IOS and android, so we could use the backup file from IOS and android to restore the vault.
type Vault struct {
	Name          string     `json:"name"`
	PubKeyECDSA   string     `json:"pubKeyECDSA"`
	PubKeyEdDSA   string     `json:"pubKeyEdDSA"`
	LocalPartyID  string     `json:"localPartyID"`
	HexChainCode  string     `json:"hexChainCode"`
	ResharePrefix string     `json:"resharePrefix"`
	Signers       []string   `json:"signers"`
	Keyshares     []KeyShare `json:"keyshares"`
}

type VaultBackup struct {
	Version string `json:"version"`
	Vault   Vault  `json:"vault"`
}
