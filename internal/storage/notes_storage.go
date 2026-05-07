package storage

import (
	"fmt"
	"simple-server/internal/model"
	"sync"
)

// хранилище для заметок
type NotesStorage struct {
	notes map[string]model.Note
	mtx   sync.RWMutex
}

func NewNotesStorage() *NotesStorage {
	return &NotesStorage{
		notes: make(map[string]model.Note),
	}
}

func (s *NotesStorage) AddNote(note model.Note) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	// проверка существования осуществляется в хранилище, чтобы избежать гонки
	if _, exists := s.notes[note.Header]; exists {
		return fmt.Errorf("note with header '%s' already exists", note.Header)
	}
	s.notes[note.Header] = note
	return nil
}

func (s *NotesStorage) GetNotes() []model.Note {
	// заполняем список значениями из map
	notesList := make([]model.Note, 0, len(s.notes))
	s.mtx.RLock()
	for _, note := range s.notes {
		notesList = append(notesList, note)
	}
	s.mtx.RUnlock()
	return notesList
}

func (s *NotesStorage) GetNoteByHeader(header string) (*model.Note, error) {
	s.mtx.RLock()
	note, exists := s.notes[header]
	s.mtx.RUnlock()
	if !exists {
		return nil, fmt.Errorf("note with header '%s' not found", header)
	}
	return &note, nil
}

func (s *NotesStorage) UpdateNote(note model.Note) error {
	s.mtx.Lock()
	s.notes[note.Header] = note
	s.mtx.Unlock()
	return nil
}

func (s *NotesStorage) DeleteNoteByHeader(header string) error {
	s.mtx.Lock()
	delete(s.notes, header)
	s.mtx.Unlock()
	return nil
}

func (s *NotesStorage) NoteExists(header string) bool {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	_, exists := s.notes[header]
	return exists
}
