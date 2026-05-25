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

func (s *NoteService) GetAllNotes(ctx context.Context) ([]model.Note, error) {
	notes, err := s.storage.GetNotes(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all notes: %w", err)
	}
	return notes, err
}

func (s *NoteService) GetNoteById(ctx context.Context, noteId uuid.UUID) (*model.Note, error) {
	note, err := s.storage.GetNoteById(ctx, noteId)
	if err != nil {
		return nil, fmt.Errorf("get note by id: %w", err)
	}
	return note, err
}

func (s *NoteService) GetNotesByHeader(ctx context.Context, header string) ([]model.Note, error) {
	notes, err := s.storage.GetNotesByHeader(ctx, header)
	if err != nil {
		return nil, fmt.Errorf("get notes by header: %w", err)
	}
	return notes, err
}

func (s *NoteService) UpdateNote(ctx context.Context, note *model.Note) error {
	err := s.storage.UpdateNote(ctx, note)
	if err != nil {
		return fmt.Errorf("update note: %w", err)
	}
	return nil
}

func (s *NoteService) DeleteNote(ctx context.Context, noteId uuid.UUID) error {
	return s.storage.DeleteNote(ctx, noteId)
}
