package memory

import (
	"context"
	"simple-server/internal/model"
	"sync"

	"github.com/google/uuid"
)

// хранилище для заметок
type NoteStorage struct {
	notes map[uuid.UUID]model.Note
	mtx   sync.RWMutex
}

func NewNoteStorage() *NoteStorage {
	return &NoteStorage{
		notes: make(map[uuid.UUID]model.Note),
	}
}

func NewNoteStorageWithData(notesList []model.Note) *NoteStorage {
	notes := make(map[uuid.UUID]model.Note, len(notesList))
	for _, elem := range notesList {
		notes[elem.ID] = elem
	}

	return &NoteStorage{
		notes: notes,
	}
}

func (s *NoteStorage) AddNote(ctx context.Context, note *model.Note) (*model.Note, error) {
	note.ID = uuid.New()
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.notes[note.ID] = *note
	return note, nil
}

func (s *NoteStorage) GetNotes(ctx context.Context, filters model.GetNotesFilters) ([]model.Note, error) {
	notesList := make([]model.Note, 0, len(s.notes))
	s.mtx.RLock()
	headerFilterExists := filters.Header != nil
	for _, note := range s.notes {
		if headerFilterExists && note.Header != *filters.Header {
			continue
		}
		notesList = append(notesList, note)
	}
	s.mtx.RUnlock()
	return notesList, nil
}

func (s *NoteStorage) GetNoteByID(ctx context.Context, noteID uuid.UUID) (*model.Note, error) {
	var note model.Note
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	note, exists := s.notes[noteID]
	if !exists {
		return nil, model.ErrNotFound
	}
	return &note, nil
}

func (s *NoteStorage) UpdateNote(ctx context.Context, note *model.Note) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if _, exists := s.notes[note.ID]; !exists {
		return model.ErrNotFound
	}
	s.notes[note.ID] = *note
	return nil
}

func (s *NoteStorage) DeleteNote(ctx context.Context, noteID uuid.UUID) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	delete(s.notes, noteID)
	return nil
}

func (s *NoteStorage) NoteExists(ctx context.Context, noteID uuid.UUID) (bool, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	_, exists := s.notes[noteID]
	return exists, nil
}
