package notes

import (
	"fmt"
	"simple-server/internal/model"
	"simple-server/internal/storage"
)

type NotesService struct {
	config  *model.Config
	storage *storage.NotesStorage
}

func NewNotesService(config *model.Config, storage *storage.NotesStorage) *NotesService {
	return &NotesService{
		config:  config,
		storage: storage,
	}
}

func (s *NotesService) AddNote(note *model.Note) error {
	if note.Header == "" {
		return fmt.Errorf("header can't be empty")
	}
	return s.storage.AddNote(*note)
}

func (s *NotesService) GetAllNotes() []model.Note {
	return s.storage.GetNotes()
}

func (s *NotesService) GetNoteByHeader(header string) (*model.Note, error) {
	return s.storage.GetNoteByHeader(header)
}

func (s *NotesService) UpdateNote(note *model.Note) error {
	return s.storage.UpdateNote(*note)
}

func (s *NotesService) DeleteNote(header string) error {
	// если заметки нет - ничего не делаем
	// (или возвращаем ошибку если надо)
	if !s.storage.NoteExists(header) {
		return nil
	}
	return s.storage.DeleteNoteByHeader(header)
}
