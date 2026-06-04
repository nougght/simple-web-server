package note

import (
	"context"
	"simple-server/internal/model"

	"github.com/google/uuid"
)

type NoteStorage interface {
	AddNote(ctx context.Context, note *model.Note) (*model.Note, error)
	GetNotes(ctx context.Context, filters model.GetNotesFilters) ([]model.Note, error)
	GetNoteByID(ctx context.Context, noteID uuid.UUID) (*model.Note, error)
	UpdateNote(ctx context.Context, note *model.Note) error
	DeleteNote(ctx context.Context, noteID uuid.UUID) error
	NoteExists(ctx context.Context, noteID uuid.UUID) (bool, error)
}
