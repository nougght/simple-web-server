package note

import (
	"context"
	"fmt"
	"simple-server/internal/config"
	"simple-server/internal/model"
	"strings"

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

func (s *NoteService) validateHeader(header string) error {
	if strings.TrimSpace(header) == "" {
		return fmt.Errorf("header can't be empty: %w", model.ErrBadRequest)
	}
	return nil
}

func (s *NoteService) AddNote(ctx context.Context, note *model.Note) (*model.Note, error) {
	if err := s.validateHeader(note.Header); err != nil {
		return nil, err
	}
	created, err := s.storage.AddNote(ctx, note)
	if err != nil {
		return nil, fmt.Errorf("add note: %w", err)
	}
	return created, nil
}

func (s *NoteService) GetNotes(ctx context.Context, filters model.GetNotesFilters) ([]model.Note, error) {
	if filters.Header != nil {
		if err := s.validateHeader(*filters.Header); err != nil {
			return nil, err
		}
	}
	notes, err := s.storage.GetNotes(ctx, filters)
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
	if err := s.validateHeader(note.Header); err != nil {
		return err
	}
	err := s.storage.UpdateNote(ctx, note)
	if err != nil {
		return fmt.Errorf("update note: %w", err)
	}
	return nil
}

func (s *NoteService) DeleteNote(ctx context.Context, noteID uuid.UUID) error {
	return s.storage.DeleteNote(ctx, noteID)
}
