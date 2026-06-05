package main

import (
	"fmt"
	"log"
	"simple-server/internal/model"
	"simple-server/internal/test-client/api_client"
	"slices"
	"time"
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
	newNotes := []model.Note{
		{Header: "header123", Body: "some body"},
		{Header: "header123", Body: "another body"},
		{Header: "other", Body: ""},
	}
	var addedNotes []*model.Note
	for _, newNote := range newNotes {
		addedNote, err := api.AddNote(&newNote)
		if err != nil {
			fmt.Printf("Ошибка: %s\n", err.Error())
		} else {
			addedNotes = append(addedNotes, addedNote)
		}
	}

	if len(addedNotes) > 0 {
		log.Println("Запрос заметки по ID")
		if note, err := api.FetchNoteByID(addedNotes[0].ID); err != nil {
			fmt.Printf("Ошибка: %s\n", err.Error())
		} else {
			fmt.Println(note)
		}

		log.Println("Запрос заметок по заголовку")
		if result, err := api.FetchNotesByHeader(addedNotes[2].Header); err != nil {
			fmt.Printf("Ошибка: %s\n", err.Error())
		} else {
			fmt.Println(result)
			if !slices.ContainsFunc(result, func(e model.Note) bool { return e.ID == addedNotes[2].ID }) {
				fmt.Print("Ответ не совпадает с ожидаемым\n\n")
			} else {
				fmt.Print("Ожидаемый результат\n\n")
			}
		}
		log.Println("Запрос нескольких заметок по заголовку")
		if result, err := api.FetchNotesByHeader(addedNotes[0].Header); err != nil {
			fmt.Printf("Ошибка: %s\n", err.Error())
		} else {
			fmt.Println(result)
			if !slices.ContainsFunc(result, func(e model.Note) bool { return e == *addedNotes[0] }) ||
				!slices.ContainsFunc(result, func(e model.Note) bool { return e == *addedNotes[1] }) {
				fmt.Print("Ответ не совпадает с ожидаемым\n\n")
			} else {
				fmt.Print("Ожидаемый результат\n\n")
			}
		}

		log.Println("Изменение заметки")
		updatedNote := model.Note{ID: addedNotes[0].ID, Header: addedNotes[0].Header, Body: "new body"}
		if err := api.UpdateNote(&updatedNote); err != nil {
			fmt.Printf("Ошибка: %s\n", err.Error())
		} else {
			if note, err := api.FetchNoteByID(updatedNote.ID); err != nil {
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
		if err := api.DeleteNote(updatedNote.ID); err != nil {
			fmt.Printf("Ошибка: %s\n", err.Error())
		} else {
			if note, err := api.FetchNoteByID(updatedNote.ID); err != nil {
				fmt.Printf("Ошибка (ожидаемо после удаления): %s\n", err.Error())
			} else {
				fmt.Printf("Заметка не удалена: %v\n\n", note)
			}
		}
	}

	log.Println("Конвертация валют")
	taskID, err := api.Convert(2000, "RUB", []string{"EUR", "USD", "CNY", "JPY"})
	if err != nil {
		fmt.Printf("Ошибка: %s\n", err.Error())
	} else {
		fmt.Println(taskID)
	}

	var status *model.TaskStatus
	// проверка статуса задачи с интервалом
	for i := 0; i < 10; i++ {
		status, err = api.FetchTaskStatus(taskID)
		if err != nil {
			fmt.Printf("Ошибка: %s\n", err.Error())
			break
		}
		log.Printf("Статус задачи: %s\n", *status)
		if *status != model.TaskStatusInProgress {
			break
		}

		time.Sleep(time.Second)
	}
	if status != nil && *status != model.TaskStatusInProgress {
		task, err := api.FetchTask(taskID)
		if err != nil {
			fmt.Printf("Ошибка: %s\n", err.Error())
		} else {
			fmt.Printf("Задача: %v\n", task)
			switch task.Status {
			case model.TaskStatusSuccess:
				fmt.Printf("Результат: %v\n", string(*task.Result))
			case model.TaskStatusFailed:
				fmt.Printf("Ошибка: %s\n", *task.Error)
			}
		}

		err = api.DeleteTask(taskID)
		if err != nil {
			fmt.Printf("Ошибка: %s\n", err.Error())
		} else {
			fmt.Printf("Задача %s удалена\n", taskID)
		}
	}

}
