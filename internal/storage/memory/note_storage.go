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
		notes[elem.NoteId] = elem
	}

	return &NoteStorage{
		notes: notes,
	}
}

func (s *NoteStorage) AddNote(ctx context.Context, note *model.Note) (*model.Note, error) {
	note.NoteId = uuid.New()
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.notes[note.NoteId] = *note
	return note, nil
}

func (s *NoteStorage) GetNotes(ctx context.Context) ([]model.Note, error) {
	// заполняем список значениями из map
	notesList := make([]model.Note, 0, len(s.notes))
	s.mtx.RLock()
	for _, note := range s.notes {
		notesList = append(notesList, note)
	}
	s.mtx.RUnlock()
	return notesList, nil
}

func (s *NoteStorage) GetNotesByHeader(ctx context.Context, header string) ([]model.Note, error) {
	s.mtx.RLock()
	notesList := make([]model.Note, 0)
	for _, note := range s.notes {
		if note.Header == header {
			notesList = append(notesList, note)
		}
	}
	s.mtx.RUnlock()
	return notesList, nil
}

func (s *NoteStorage) GetNoteById(ctx context.Context, id uuid.UUID) (*model.Note, error) {
	var note model.Note
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	note, exists := s.notes[id]
	if !exists {
		return nil, model.ErrNotFound
	}
	return &note, nil
}

func (s *NoteStorage) UpdateNote(ctx context.Context, note *model.Note) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if _, exists := s.notes[note.NoteId]; !exists {
		return model.ErrNotFound
	}
	s.notes[note.NoteId] = *note
	return nil
}

func (s *NoteStorage) DeleteNote(ctx context.Context, id uuid.UUID) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	delete(s.notes, id)
	return nil
}

func (s *NoteStorage) NoteExists(ctx context.Context, id uuid.UUID) (bool, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	_, exists := s.notes[id]
	return exists, nil
}
