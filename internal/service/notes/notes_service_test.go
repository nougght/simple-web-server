package notes

import (
	"simple-server/internal/model"
	"simple-server/internal/storage"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddNote(t *testing.T) {
	storage := storage.NewNotesStorage()
	service := NewNotesService(storage)

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
			err := service.AddNote(&test.note)
			if test.errorExpected {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
