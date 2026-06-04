package model

import "github.com/google/uuid"

const (
	NoteFilterHeader string = "header"
)

// необязательные фильтры для получения заметок
type GetNotesFilters struct {
	Header *string
}

type UpdateNoteRequestBody struct {
	Header string `json:"header" db:"header"`
	Body   string `json:"body" db:"body"`
}

type Note struct {
	ID     uuid.UUID `json:"id" db:"id"`
	Header string    `json:"header" db:"header"`
	Body   string    `json:"body" db:"body"`
}
