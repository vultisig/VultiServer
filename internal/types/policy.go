package types

import "encoding/json"

type PluginPolicy struct {
	ID            string          `json:"id" validate:"required"`
	PublicKey     string          `json:"public_key" validate:"required"`
	PluginID      string          `json:"plugin_id" validate:"required"`
	PluginVersion string          `json:"plugin_version" validate:"required"`
	PolicyVersion string          `json:"policy_version"	validate:"required"`
	PluginType    string          `json:"plugin_type" validate:"required"`
	Signature     string          `json:"signature" validate:"required"`
	Policy        json.RawMessage `json:"policy" validate:"required"`
}

type PayrollPolicy struct {
	ChainID    string             `json:"chain_id"`
	TokenID    string             `json:"token_id"`
	Recipients []PayrollRecipient `json:"recipients"`
}

type PayrollRecipient struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
}
