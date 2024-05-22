package models

import (
	"github.com/jinzhu/gorm"
)

type KeyGeneration struct {
	gorm.Model

	Key       string
	Parties   string // csv
	Session   string
	ChainCode string
}
