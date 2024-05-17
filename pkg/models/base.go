package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

// https://medium.com/@the.hasham.ali/how-to-use-uuid-key-type-with-gorm-cc00d4ec7100

type Base struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" sql:"index"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (base *Base) BeforeCreate(scope *gorm.Scope) error {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	return scope.SetColumn("ID", uuid)
}
