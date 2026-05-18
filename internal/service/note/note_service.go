package note

import (
	"fmt"
	"simple-server/internal/model"
	"simple-server/internal/storage"
)

type NoteService struct {
	config  *model.Config
	storage *storage.NoteStorage
}

func NewNoteService(config *model.Config, storage *storage.NoteStorage) *NoteService {
	return &NoteService{
		config:  config,
		storage: storage,
	}
}

func (s *NoteService) AddNote(note *model.Note) error {
	if note.Header == "" {
		return fmt.Errorf("header can't be empty")
	}
	return s.storage.AddNote(*note)
}

func (s *NoteService) GetAllNotes() []model.Note {
	return s.storage.GetNotes()
}

func (s *NoteService) GetNoteByHeader(header string) (*model.Note, error) {
	return s.storage.GetNoteByHeader(header)
}

func (s *NoteService) UpdateNote(note *model.Note) error {
	return s.storage.UpdateNote(*note)
}

func (s *NoteService) DeleteNote(header string) error {
	// если заметки нет - ничего не делаем
	// (или возвращаем ошибку если надо)
	if !s.storage.NoteExists(header) {
		return nil
	}
	return s.storage.DeleteNoteByHeader(header)
}
