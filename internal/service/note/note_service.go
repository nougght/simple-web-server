package note

import (
	"context"
	"fmt"
	"simple-server/internal/config"
	"simple-server/internal/model"

	"github.com/google/uuid"
)

type NoteService struct {
	config  *config.Config
	storage NoteStorage
}

func NewNoteService(config *config.Config, storage NoteStorage) *NoteService {
	return &NoteService{
		config:  config,
		storage: storage,
	}
}

func (s *NoteService) AddNote(ctx context.Context, note *model.Note) (*model.Note, error) {
	if note.Header == "" {
		return nil, fmt.Errorf("header can't be empty: %w", model.ErrBadRequest)
	}
	created, err := s.storage.AddNote(ctx, note)
	if err != nil {
		return nil, fmt.Errorf("add note: %w", err)
	}
	return created, nil
}

func (s *NoteService) GetNotes(ctx context.Context, filters map[string]interface{}) ([]model.Note, error) {
	notes, err := s.storage.GetNotes(ctx, filters)
	if header, ok := filters["header"]; ok && header == "" {
		return nil, fmt.Errorf("header filter can't be empty: %w", model.ErrBadRequest)
	}
	if err != nil {
		return nil, fmt.Errorf("get notes with filters %v: %w", filters, err)
	}
	return notes, err
}

func (s *NoteService) GetNoteByID(ctx context.Context, noteID uuid.UUID) (*model.Note, error) {
	note, err := s.storage.GetNoteByID(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("get note by ID: %w", err)
	}
	return note, err
}

func (s *NoteService) UpdateNote(ctx context.Context, note *model.Note) error {
	err := s.storage.UpdateNote(ctx, note)
	if err != nil {
		return fmt.Errorf("update note: %w", err)
	}
	return nil
}

func (s *NoteService) DeleteNote(ctx context.Context, noteID uuid.UUID) error {
	return s.storage.DeleteNote(ctx, noteID)
}
