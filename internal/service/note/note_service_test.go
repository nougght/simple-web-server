package note

import (
	"context"
	"simple-server/internal/config"
	"simple-server/internal/model"
	"simple-server/internal/storage/memory"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var cfg = &config.Config{
	NoteStorageType: model.StorageTypeMemory,
}

func TestAddNote(t *testing.T) {
	storage := memory.NewNoteStorage()
	service := NewNoteService(cfg, storage)

	notes := []model.Note{
		{Header: "header1", Body: "some body"},
		{Header: "", Body: "another body"},
	}
	tests := []struct {
		name          string
		note          model.Note
		errorExpected bool
	}{
		{"default add", notes[0], false},
		{"empty header add", notes[1], true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			note, err := service.AddNote(context.Background(), &test.note)
			if test.errorExpected {
				assert.Error(t, err)
				assert.Nil(t, note)
				return
			}
			assert.Nil(t, err)
			require.NotNil(t, note)
			assert.NotEqual(t, uuid.Nil, note.ID)
			assert.Equal(t, test.note.Header, note.Header)
			assert.Equal(t, test.note.Body, note.Body)

		})
	}
}

func TestGetNotes(t *testing.T) {
	storage := memory.NewNoteStorageWithData([]model.Note{
		{ID: uuid.New(), Header: "header1", Body: "some body"},
		{ID: uuid.New(), Header: "header2", Body: "another body"},
	})
	service := NewNoteService(cfg, storage)

	headers := []string{"header1", "header3843", " "}
	tests := []struct {
		name          string
		filters       model.NotesFilters
		expectedCount int
		errorExpected bool
	}{
		{"no filters", model.NotesFilters{}, 2, false},
		{"filter by header", model.NotesFilters{Header: &headers[0]}, 1, false},
		{"filter by header that doesn't exist", model.NotesFilters{Header: &headers[1]}, 0, false},
		{"filter by empty header", model.NotesFilters{Header: &headers[2]}, 0, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			notes, err := service.GetNotes(context.Background(), test.filters)
			if test.errorExpected {
				assert.Error(t, err)
				assert.Nil(t, notes)
				return
			}
			assert.Nil(t, err)
			assert.Len(t, notes, test.expectedCount)
		})
	}
}
