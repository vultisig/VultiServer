package types

import "time"

type UserAuthDto struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type User struct {
	ID        string    `json:"id" validate:"required"`
	Username  string    `json:"username" validate:"required"`
	CreatedAt time.Time `json:"created_at"`
}

type UserWithPassword struct {
	ID        string    `json:"id" validate:"required"`
	Username  string    `json:"username" validate:"required"`
	Password  string    `json:"password" validate:"required"`
	CreatedAt time.Time `json:"created_at"`
}
