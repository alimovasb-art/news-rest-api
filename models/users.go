package models

import "time"

type User struct {
	ID        int        `json:"id"`
	FirstName string     `json:"first_name" validate:"required,min=3"`
	LastName  string     `json:"last_name" validate:"required,min=3"`
	Email     string     `json:"email" validate:"required,email"`
	Password  string     `json:"password" validate:"required,min=8,max=64"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type UpdateUser struct {
	FirstName string `json:"first_name" validate:"omitempty,min=3"`
	LastName  string `json:"last_name" validate:"omitempty,min=3"`
	Email     string `json:"email" validate:"omitempty,email"`
	Password  string `json:"password" validate:"omitempty,min=8,max=64"`
}
