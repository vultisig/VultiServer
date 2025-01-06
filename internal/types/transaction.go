package types

import (
	"time"

	"github.com/google/uuid"
)

type TransactionStatus string

const (
	StatusPending       TransactionStatus = "PENDING"
	StatusSigningFailed TransactionStatus = "SIGNING_FAILED"
	StatusSigned        TransactionStatus = "SIGNED"
	StatusBroadcast     TransactionStatus = "BROADCAST"
	StatusMined         TransactionStatus = "MINED"
	StatusRejected      TransactionStatus = "REJECTED"
)

type TransactionHistory struct {
	ID           uuid.UUID              `json:"id"`
	PolicyID     uuid.UUID              `json:"policy_id"`
	TxBody       string                 `json:"tx_body"`
	Status       TransactionStatus      `json:"status"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Metadata     map[string]interface{} `json:"metadata"`
	ErrorMessage *string                `json:"error_message,omitempty"`
}
