package model

import (
	"context"

	"github.com/google/uuid"
)

type CurrencyService interface {
	ConvertCurrency(params *ConvertCurrencyParams) (map[string]float64, error)
}

type NoteService interface {
	AddNote(ctx context.Context, note *Note) (*Note, error)
	GetAllNotes(ctx context.Context) ([]Note, error)
	GetNoteById(ctx context.Context, noteId uuid.UUID) (*Note, error)
	GetNotesByHeader(ctx context.Context, header string) ([]Note, error)
	UpdateNote(ctx context.Context, note *Note) error
	DeleteNote(ctx context.Context, noteId uuid.UUID) error
}

type Service interface {
	NoteService() NoteService
	CurrencyService() CurrencyService
}
