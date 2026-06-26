package model

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

type CurrencyService interface {
	ConvertCurrency(ctx context.Context, params *ConvertCurrencyParams) (map[string]float64, error)
	ConvertCurrencyHandler(ctx context.Context, payload json.RawMessage) (any, error)
	ConvertAndSaveAsync(ctx context.Context, params *ConvertCurrencyParams) (uuid.UUID, error)
}

type NoteService interface {
	AddNote(ctx context.Context, note *Note) (*Note, error)
	GetNotes(ctx context.Context, filters NotesFilters) ([]Note, error)
	GetNoteByID(ctx context.Context, noteID uuid.UUID) (*Note, error)
	UpdateNote(ctx context.Context, note *Note) error
	DeleteNote(ctx context.Context, noteID uuid.UUID) error
}

type TaskService interface {
	RegisterHandler(taskType TaskType, handler TaskHandler)
	StartWorkers(ctx context.Context)
	Stop()
	GetTaskStatus(ctx context.Context, id uuid.UUID) (*TaskStatus, error)
	GetTaskByID(ctx context.Context, id uuid.UUID) (*Task, error)
	DeleteTask(ctx context.Context, id uuid.UUID) error
	ExecuteAndSaveAsync(ctx context.Context, taskType TaskType, payload any) (uuid.UUID, error)
}

type Service interface {
	NoteService() NoteService
	CurrencyService() CurrencyService
	TaskService() TaskService
}
