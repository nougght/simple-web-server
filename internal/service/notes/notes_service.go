package notes

import (
	"simple-server/internal/model"
	"simple-server/internal/storage"
)

type NotesService struct {
	storage *storage.NotesStorage
}

func NewNotesService(storage *storage.NotesStorage) *NotesService {
	return &NotesService{
		storage: storage,
	}
}

func (s *NotesService) AddNote(note *model.Note) error {
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
	return s.storage.DeleteNoteByHeader(header)
}
