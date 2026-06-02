package model

import "github.com/google/uuid"

type UpdateNoteRequestBody struct {
	Header string `json:"header" db:"header"`
	Body   string `json:"body" db:"body"`
}

type Note struct {
	ID     uuid.UUID `json:"id" db:"id"`
	Header string    `json:"header" db:"header"`
	Body   string    `json:"body" db:"body"`
}
