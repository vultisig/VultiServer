package models

type TransactionPolicy struct {
	Base

	Limit         float64 `json:"limit" validate:"required,gte=0"`
	Delay         int     `json:"delay" validate:"required,gte=0"`
	EmailApproval string  `json:"email_approval" validate:"required,email"`
	Password      string  `json:"password" validate:"required,min=8"`
}
