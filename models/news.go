package models

import "time"

type News struct {
	ID               int    `json:"id"`
	Title            string `json:"title" validate:"required,min=3"`
	ShortDescription string `json:"short_description" validate:"required,min=10"`
	Description      string `json:"description" validate:"required,min=15"`
	Views            int    `json:"views"`

	AuthorID int    `json:"author_id"`
	ImageURL string `json:"image_url,omitempty"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type NewsResponse struct {
	ID               int    `json:"id"`
	Title            string `json:"title"`
	ShortDescription string `json:"short_description"`
	Description      string `json:"description"`
	Views            int    `json:"views"`

	AuthorID int     `json:"author_id"`
	Author   *Author `json:"author"`

	ImageURL string `json:"image_url,omitempty"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type UpdateNews struct {
	Title            string `json:"title" validate:"omitempty,min=3"`
	ShortDescription string `json:"short_description" validate:"omitempty,min=10"`
	Description      string `json:"description" validate:"omitempty,min=15"`
}
