package note

import (
	"context"
	"simple-server/internal/config"
	"simple-server/internal/model"
	"simple-server/internal/storage/memory"
	"testing"

	"github.com/stretchr/testify/assert"
)

var cfg = &config.Config{
	StorageType: model.StorageTypeMemory,
}

func TestAddNote(t *testing.T) {
	storage := memory.NewNoteStorage()
	service := NewNoteService(cfg, storage)

	notes := []model.Note{
		{Header: "header1", Body: "some body"},
		{Header: "", Body: "another body"},
	}
	test := []struct {
		name          string
		note          model.Note
		errorExpected bool
	}{
		{"default add", notes[0], false},
		{"empty header add", notes[1], true},
	}

	for _, test := range test {
		t.Run(test.name, func(t *testing.T) {
			note, err := service.AddNote(context.Background(), &test.note)
			if test.errorExpected {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.note.Header, note.Header)
				assert.Equal(t, test.note.Body, note.Body)
			}
		})
	}
}
