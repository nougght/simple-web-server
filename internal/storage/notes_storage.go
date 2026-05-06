package storage

import (
	"fmt"
	"log"
	"simple-server/internal/model"
)

// хранилище для заметок
type NotesStorage struct {
	notes map[string]model.Note
}

func NewNotesStorage() *NotesStorage {
	return &NotesStorage{
		notes: make(map[string]model.Note),
	}
}

func (s *NotesStorage) AddNote(note model.Note) error {
	if _, exists := s.notes[note.Header]; exists == true {
		log.Print("Заметка с указанным заголовком уже существует\n\n")
		return fmt.Errorf("note with header '%s' already exists", note.Header)
	}
	s.notes[note.Header] = note
	return nil
}

func (s *NotesStorage) GetNotes() []model.Note {
	// заполняем список значениями из map
	notesList := make([]model.Note, 0, len(s.notes))
	for _, note := range s.notes {
		notesList = append(notesList, note)
	}
	return notesList
}

func (s *NotesStorage) GetNoteByHeader(header string) (*model.Note, error) {
	note, exists := s.notes[header]
	if exists == false {
		log.Print("Запрашиваемая заметка не найдена\n\n")
		return nil, fmt.Errorf("note not with header '%s' found", header)
	}
	return &note, nil
}

func (s *NotesStorage) UpdateNote(note model.Note) error {
	s.notes[note.Header] = note
	return nil
}

func (s *NotesStorage) DeleteNoteByHeader(header string) error {
	delete(s.notes, header)
	return nil
}
