package storage

import (
	"simple-server/internal/model"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// проверка добавления заметки
func TestAdd(t *testing.T) {
	storage := NewNotesStorage()

	notes := []model.Note{
		model.Note{Header: "header123", Body: "some body"},
		model.Note{Header: "header123", Body: ""},
		model.Note{Header: "other", Body: "dfdfs"},
	}

	addTests := []struct {
		name          string
		note          model.Note
		errorExpected bool // ожидается ли ошибка
	}{
		{"default add", notes[0], false},
		{"duplicate error add", notes[1], true}, // ошибка из-за дублирования заголовка
		{"default add 2", notes[2], false},
	}

	for _, test := range addTests {

		t.Run(test.name, func(t *testing.T) {
			err := storage.AddNote(test.note)
			// проверка ошибки при добавлении
			if test.errorExpected {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
			}

			result, ok := storage.notes[test.note.Header]
			// проверка наличия добавленной заметки
			assert.True(t, ok)
			require.NotNil(t, result)

			if test.errorExpected {
				// дубликат не должен был сохраниться
				assert.NotEqual(t, test.note, result)
			} else {
				assert.Equal(t, test.note, result)
			}
		})
	}
	if len(storage.notes) != 2 {
		t.Errorf("Количество добавленных заметок(%d) не совпадает с ожиданием", len(storage.notes))
	}

}

// проверка получения заметки
func TestGet(t *testing.T) {
	notes := []model.Note{
		model.Note{Header: "header1", Body: "some body"},
		model.Note{Header: "header2", Body: "sdsfsdfds"},
	}
	// создаем хранилище с заполненными данными
	storage := NewNotesStorageWithData(notes)

	tests := []struct {
		name          string
		header        string
		expected      *model.Note
		errorExpected bool // ожидается ли ошибка
	}{
		{"default get", notes[0].Header, &notes[0], false},
		{"default get 2", notes[1].Header, &notes[1], false},
		{"not found error get", "header3", nil, true}, // ошибка,т.к. такой заметки нет
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := storage.GetNoteByHeader(test.header)

			if test.errorExpected {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				require.NotNil(t, result)

				assert.Equal(t, *test.expected, *result)
			}

		})
	}

}

func TestUpdate(t *testing.T) {
	notes := []model.Note{
		model.Note{Header: "header1", Body: "some body"},
	}
	storage := NewNotesStorageWithData(notes)

	newNote := model.Note{Header: "header1", Body: "new body"}

	// обновляем заметку (с тем же заголовком)
	err := storage.UpdateNote(newNote)
	assert.Nil(t, err)

	// заметка должна замениться
	result, ok := storage.notes[newNote.Header]
	assert.True(t, ok)
	require.NotNil(t, result)
	assert.Equal(t, newNote, result)
}

func TestDelete(t *testing.T) {
	notes := []model.Note{
		model.Note{Header: "header1", Body: "some body"},
		model.Note{Header: "header2", Body: "sfjdsiofj"},
	}
	storage := NewNotesStorageWithData(notes)

	// обновляем заметку (с тем же заголовком)
	err := storage.DeleteNoteByHeader(notes[0].Header)
	assert.Nil(t, err)

	// заметка должна остсутствовать
	_, ok := storage.notes[notes[0].Header]
	assert.False(t, ok)

	if len(storage.notes) != 1 {
		t.Errorf("Количество добавленных заметок(%d) не совпадает с ожиданием", len(storage.notes))
	}
}
