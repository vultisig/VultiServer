package models

// type PaillierSK struct {
// 	N       string `json:"N" validate:"required,numeric"`
// 	LambdaN string `json:"LambdaN" validate:"required,numeric"`
// 	PhiN    string `json:"PhiN" validate:"required,numeric"`
// 	P       string `json:"P" validate:"required,numeric"`
// 	Q       string `json:"Q" validate:"required,numeric"`
// }

// type ECDSAPub struct {
// 	Curve  string   `json:"Curve" validate:"required,oneof=secp256k1"`
// 	Coords []string `json:"Coords" validate:"required,len=2,hexadecimal"`
// }

// type EcdsaLocalData struct {
// 	PaillierSK  PaillierSK `json:"PaillierSK" validate:"required"`
// 	NTildei     string     `json:"NTildei" validate:"required,numeric"`
// 	H1i         string     `json:"H1i" validate:"required,numeric"`
// 	H2i         string     `json:"H2i" validate:"required,numeric"`
// 	Alpha       string     `json:"Alpha" validate:"required,numeric"`
// 	Beta        string     `json:"Beta" validate:"required,numeric"`
// 	P           string     `json:"P" validate:"required,numeric"`
// 	Q           string     `json:"Q" validate:"required,numeric"`
// 	Xi          string     `json:"Xi" validate:"required,numeric"`
// 	ShareID     string     `json:"ShareID" validate:"required,numeric"`
// 	Ks          []string   `json:"Ks" validate:"required,dive,numeric"`
// 	NTildej     []string   `json:"NTildej" validate:"required,dive,numeric"`
// 	H1j         []string   `json:"H1j" validate:"required,dive,numeric"`
// 	H2j         []string   `json:"H2j" validate:"required,dive,numeric"`
// 	BigXj       []ECDSAPub `json:"BigXj" validate:"required,dive,required"`
// 	PaillierPKs []string   `json:"PaillierPKs" validate:"required,dive,numeric"`
// 	ECDSAPub    ECDSAPub   `json:"ECDSAPub" validate:"required"`
// }

// type EddsaLocalData struct {
// 	Xi       *string `json:"Xi" validate:"omitempty,numeric"`
// 	ShareID  *string `json:"ShareID" validate:"omitempty,numeric"`
// 	Ks       *string `json:"Ks" validate:"omitempty,numeric"`
// 	BigXj    *string `json:"BigXj" validate:"omitempty,hexadecimal"`
// 	EDDSAPub *string `json:"EDDSAPub" validate:"omitempty,hexadecimal"`
// }

type VaultShare struct {
	Base

	PubKey              string         `json:"pub_key" validate:"required,len=66,hexadecimal"`
	EcdsaLocalData      EcdsaLocalData `json:"ecdsa_local_data" validate:"required"`
	EddsaLocalData      EddsaLocalData `json:"eddsa_local_data"`
	KeygenCommitteeKeys []string       `json:"keygen_committee_keys" validate:"required,dive,required"`
	LocalPartyKey       string         `json:"local_party_key" validate:"required"`
	ChainCodeHex        string         `json:"chain_code_hex" validate:"required,len=64,hexadecimal"`
	ResharePrefix       string         `json:"reshare_prefix"`
}
