package storage

import (
	"fmt"
	"simple-server/internal/model"
	"sync"
)

// хранилище для заметок
type NoteStorage struct {
	notes map[string]model.Note
	mtx   sync.RWMutex
}

func NewNoteStorage() *NoteStorage {
	return &NoteStorage{
		notes: make(map[string]model.Note),
	}
}

func NewNoteStorageWithData(notesList []model.Note) *NoteStorage {
	notes := make(map[string]model.Note, len(notesList))
	for _, elem := range notesList {
		notes[elem.Header] = elem
	}

	return &NoteStorage{
		notes: notes,
	}
}

func (s *NoteStorage) AddNote(note model.Note) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	// проверка существования осуществляется в хранилище, чтобы избежать гонки
	if _, exists := s.notes[note.Header]; exists {
		return fmt.Errorf("note with header '%s' already exists", note.Header)
	}
	s.notes[note.Header] = note
	return nil
}

func (s *NoteStorage) GetNotes() []model.Note {
	// заполняем список значениями из map
	notesList := make([]model.Note, 0, len(s.notes))
	s.mtx.RLock()
	for _, note := range s.notes {
		notesList = append(notesList, note)
	}
	s.mtx.RUnlock()
	return notesList
}

func (s *NoteStorage) GetNoteByHeader(header string) (*model.Note, error) {
	s.mtx.RLock()
	note, exists := s.notes[header]
	s.mtx.RUnlock()
	if !exists {
		return nil, fmt.Errorf("note with header '%s' not found", header)
	}
	return &note, nil
}

func (s *NoteStorage) UpdateNote(note model.Note) error {
	s.mtx.Lock()
	s.notes[note.Header] = note
	s.mtx.Unlock()
	return nil
}

func (s *NoteStorage) DeleteNoteByHeader(header string) error {
	s.mtx.Lock()
	delete(s.notes, header)
	s.mtx.Unlock()
	return nil
}

func (s *NoteStorage) NoteExists(header string) bool {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	_, exists := s.notes[header]
	return exists
}
