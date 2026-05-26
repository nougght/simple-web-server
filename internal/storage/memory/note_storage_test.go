package memory

import (
	"context"
	"log"
	"simple-server/internal/model"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// проверка добавления заметки
func TestAdd(t *testing.T) {
	storage := NewNoteStorage()

	notes := []model.Note{
		{Header: "header123", Body: "some body"},
		{Header: "other", Body: ""},
	}

	addTests := []struct {
		name          string
		note          *model.Note
		errorExpected bool
	}{
		{"default add", &notes[0], false},
		{"empty body", &notes[1], false},
	}

	for _, test := range addTests {

		t.Run(test.name, func(t *testing.T) {
			result, err := storage.AddNote(context.Background(), test.note)
			// проверка ошибки при добавлении
			if test.errorExpected {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				require.NotNil(t, result)
				assert.Equal(t, test.note.Header, result.Header)
				assert.Equal(t, test.note.Body, result.Body)
			}

			note, ok := storage.notes[test.note.ID]
			if test.errorExpected {
				assert.False(t, ok)
				assert.Nil(t, note)
			} else {
				assert.True(t, ok)
				require.NotNil(t, note)
				assert.Equal(t, test.note.Header, note.Header)
				assert.Equal(t, test.note.Body, note.Body)
			}
		})
	}
	if len(storage.notes) != 2 {
		t.Errorf("Количество добавленных заметок(%d) не совпадает с ожиданием", len(storage.notes))
	}

}

// проверка получения заметки
func TestGetNotes(t *testing.T) {
	notes := []model.Note{
		{ID: uuid.New(), Header: "header1", Body: "some body"},
		{ID: uuid.New(), Header: "header2", Body: "another body"},
		{ID: uuid.New(), Header: "header2", Body: "sdsfsdfds"},
	}
	// создаем хранилище с заполненными данными
	storage := NewNoteStorageWithData(notes)
	log.Println(storage.notes)

	randomHeader := "header" + uuid.New().String()
	tests := []struct {
		name     string
		header   *string
		expected []*model.Note
	}{
		{"get all notes", nil, []*model.Note{&notes[0], &notes[1], &notes[2]}},
		{"get by header", &notes[0].Header, []*model.Note{&notes[0]}},
		{"get by header with multiple notes", &notes[1].Header, []*model.Note{&notes[1], &notes[2]}},
		{"get by header that doesn't exist", &randomHeader, []*model.Note{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filters := make(map[string]interface{})
			if test.header != nil {
				filters["header"] = *test.header
			}
			result, err := storage.GetNotes(context.Background(), filters)

			assert.Nil(t, err)
			assert.NotNil(t, result)
			assert.Len(t, result, len(test.expected))
			// проверяем, что результат содержит все ожидаемые заметки
			for _, expectedNote := range test.expected {
				assert.True(t, slices.ContainsFunc(result, func(note model.Note) bool {
					return note == *expectedNote
				}))
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	notes := []model.Note{
		{ID: uuid.New(), Header: "header1", Body: "some body"},
	}
	storage := NewNoteStorageWithData(notes)
	log.Println(storage.notes)

	newNote := model.Note{ID: notes[0].ID, Header: "header1", Body: "new body"}

	// обновляем заметку (с тем же заголовком)
	err := storage.UpdateNote(context.Background(), &newNote)
	log.Println(storage.notes)
	assert.Nil(t, err)

	// заметка должна замениться
	result, ok := storage.notes[newNote.ID]
	assert.True(t, ok)
	require.NotNil(t, result)
	assert.Equal(t, newNote, result)
}

func TestDelete(t *testing.T) {
	notes := []model.Note{
		{ID: uuid.New(), Header: "header1", Body: "some body"},
		{ID: uuid.New(), Header: "header2", Body: "sfjdsiofj"},
	}
	storage := NewNoteStorageWithData(notes)

	// обновляем заметку (с тем же заголовком)
	err := storage.DeleteNote(context.Background(), notes[0].ID)
	assert.Nil(t, err)

	// заметка должна остсутствовать
	_, ok := storage.notes[notes[0].ID]
	assert.False(t, ok)

	if len(storage.notes) != 1 {
		t.Errorf("Количество добавленных заметок(%d) не совпадает с ожиданием", len(storage.notes))
	}
}
