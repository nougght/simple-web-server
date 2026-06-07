package model

import (
	"context"

	"github.com/google/uuid"
)

type CurrencyService interface {
	ConvertCurrency(ctx context.Context, params *ConvertCurrencyParams) (map[string]float64, error)
	ConvertAndSaveAsync(ctx context.Context, params *ConvertCurrencyParams) (uuid.UUID, error)
}

type NoteService interface {
	AddNote(ctx context.Context, note *Note) (*Note, error)
	GetNotes(ctx context.Context, filters GetNotesFilters) ([]Note, error)
	GetNoteByID(ctx context.Context, noteID uuid.UUID) (*Note, error)
	UpdateNote(ctx context.Context, note *Note) error
	DeleteNote(ctx context.Context, noteID uuid.UUID) error
}

type TaskService interface {
	GetTaskStatus(ctx context.Context, id uuid.UUID) (*TaskStatus, error)
	GetTaskByID(ctx context.Context, id uuid.UUID) (*Task, error)
	DeleteTask(ctx context.Context, id uuid.UUID) error
	ExecuteAndSaveAsync(ctx context.Context, taskFunc func(context.Context) (any, error)) (uuid.UUID, error)
}

type Service interface {
	NoteService() NoteService
	CurrencyService() CurrencyService
	TaskService() TaskService
}
