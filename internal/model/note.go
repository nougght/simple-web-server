package model

import "github.com/google/uuid"

type Note struct {
	NoteId uuid.UUID `json:"note_id" db:"note_id"`
	Header string    `json:"header" db:"header"`
	Body   string    `json:"body" db:"body"`
}
