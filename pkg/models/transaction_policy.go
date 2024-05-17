package models

import (
	"github.com/jinzhu/gorm"
)

type TransactionPolicy struct {
	gorm.Model
	Limit         float64 `json:"limit"`
	Delay         int     `json:"delay"`
	EmailApproval string  `json:"email_approval"`
	Password      string  `json:"password"`
}
