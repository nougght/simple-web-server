package note

import "simple-server/internal/model"

type NoteStorage interface {
	AddNote(note model.Note) error
	GetNotes() []model.Note
	GetNoteByHeader(header string) (*model.Note, error)
	UpdateNote(note model.Note) error
	DeleteNoteByHeader(header string) error
	NoteExists(header string) bool
}
