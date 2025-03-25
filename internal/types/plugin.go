package types

import "encoding/json"

type Plugin struct {
	ID             string          `json:"id" validate:"required"`
	Type           string          `json:"type" validate:"required"`
	Title          string          `json:"title" validate:"required"`
	Description    string          `json:"description" validate:"required"`
	Metadata       json.RawMessage `json:"metadata" validate:"required"`
	ServerEndpoint string          `json:"server_endpoint" validate:"required"`
	PricingID      string          `json:"pricing_id" validate:"required"`
}

type PluginCreateDto struct {
	Type           string          `json:"type" validate:"required"`
	Title          string          `json:"title" validate:"required"`
	Description    string          `json:"description" validate:"required"`
	Metadata       json.RawMessage `json:"metadata" validate:"required"`
	ServerEndpoint string          `json:"server_endpoint" validate:"required"`
	PricingID      string          `json:"pricing_id" validate:"required"`
}

// using references on struct fields allows us to process partially field DTOs
type PluginUpdateDto struct {
	Title          *string          `json:"title"`
	Description    *string          `json:"description"`
	Metadata       *json.RawMessage `json:"metadata"`
	ServerEndpoint *string          `json:"server_endpoint"`
	PricingID      *string          `json:"pricing_id"`
}
