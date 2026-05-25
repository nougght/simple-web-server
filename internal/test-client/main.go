// nolint
package main

import (
	"fmt"
	"log"
	"simple-server/internal/model"
	"simple-server/internal/test-client/api_client"
	"slices"
)

// проверка работы сервера
func main() {
	api := api_client.NewApiClient("http://127.0.0.1:8085")

	log.Println("Запрос списка заметок")
	notes, err := api.FetchAllNotes()
	if err != nil {
		fmt.Printf("Ошибка: %s\n\n", err.Error())
	} else {
		fmt.Println(notes)
	}

	log.Println("Добавление новой заметки")
	newNote := model.Note{Header: "my note 1", Body: "dfjsiefsdjflsehdjf"}
	addedNote, err := api.AddNote(&newNote)
	if err != nil {
		fmt.Printf("Ошибка: %s\n", err.Error())
	} else {
		fmt.Printf("Добавлена заметка: %v\n", addedNote)
	}

	if addedNote != nil {
		log.Println("Запрос заметки по Id")
		if note, err := api.FetchNoteById(addedNote.NoteId); err != nil {
			fmt.Printf("Ошибка: %s\n", err.Error())
		} else {
			fmt.Println(note)
		}

		log.Println("Запрос заметок по заголовку")
		if found, err := api.FetchNotesByHeader(addedNote.Header); err != nil {
			fmt.Printf("Ошибка: %s\n", err.Error())
		} else {
			fmt.Println(found)
			if !slices.ContainsFunc(found, func(e model.Note) bool { return e.NoteId == addedNote.NoteId }) {
				fmt.Print("Ответ не совпадает с ожидаемым\n\n")
			} else {
				fmt.Print("Ожидаемый результат\n\n")
			}
		}

		log.Println("Изменение заметки")
		updatedNote := model.Note{NoteId: addedNote.NoteId, Header: addedNote.Header, Body: "new body"}
		if err := api.UpdateNote(&updatedNote); err != nil {
			fmt.Printf("Ошибка: %s\n", err.Error())
		} else {
			if note, err := api.FetchNoteById(updatedNote.NoteId); err != nil {
				fmt.Printf("Ошибка: %s\n", err.Error())
			} else {
				if *note != updatedNote {
					fmt.Print("Ответ не совпадает с ожидаемым\n\n")
				} else {
					fmt.Print("Ожидаемый результат\n\n")
				}
			}
		}

		log.Println("Удаление заметки")
		if err := api.DeleteNote(updatedNote.NoteId); err != nil {
			fmt.Printf("Ошибка: %s\n", err.Error())
		} else {
			if note, err := api.FetchNoteById(updatedNote.NoteId); err != nil {
				fmt.Printf("Ошибка (ожидаемо после удаления): %s\n", err.Error())
			} else {
				fmt.Printf("Заметка не удалена: %v\n\n", note)
			}
		}
	}

	log.Println("Конвертация валют")
	if res, err := api.Convert(2000, "RUB", []string{"EUR", "USD", "CNY", "JPY"}); err != nil {
		fmt.Printf("Ошибка: %s\n", err.Error())
	} else {
		fmt.Println(res)
	}
}
