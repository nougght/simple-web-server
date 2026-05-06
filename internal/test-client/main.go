package main

import (
	"fmt"
	"log"
	"slices"
	. "test-client/api_client"
)

// проверка работы сервера
func main() {
	api := NewApiClient("http://127.0.0.1:8081")

	log.Println("Запрос списка заметок")
	notes, err := api.FetchAllNotes()
	if err != nil {
		fmt.Printf("Ошибка: %s\n\n", err.Error())
	} else {
		fmt.Println(notes)
	}

	log.Println("Добавление новых заметок")
	newNotes := []Note{Note{Header: "my note 1", Body: "dfjsiefsdjflsehdjf"},
		Note{Header: "my note 2", Body: "4536 dfdfj 343704"},
		Note{Header: "my note 3", Body: ""}}
	notes = append(notes, newNotes...)

	if err := api.AddNote(&notes[0]); err != nil {
		fmt.Printf("Ошибка: %s\n", err.Error())
	}
	if err := api.AddNote(&notes[1]); err != nil {
		fmt.Printf("Ошибка: %s\n", err.Error())
	}
	if err := api.AddNote(&notes[2]); err != nil {
		fmt.Printf("Ошибка: %s\n", err.Error())
	}
	// проверка добавленных элементов
	if resultNotes, err := api.FetchAllNotes(); err != nil {
		fmt.Printf("Ошибка: %s\n", err.Error())
	} else {
		fmt.Println(resultNotes)

		i := 0
		for ; i < len(notes) && slices.ContainsFunc(resultNotes,
			func(e Note) bool { return e.Header == notes[i].Header }); i++ {
		}

		if i < len(notes) {
			fmt.Print("Ответ не совпадает с ожидаемым\n\n")
		} else {
			fmt.Print("Ожидаемый результат\n\n")
		}
	}

	log.Println("Запрос одной из заметок")
	if note, err := api.FetchNote("my note 2"); err != nil {
		fmt.Printf("Ошибка: %s\n", err.Error())
	} else {
		fmt.Println(note)
		// проверка соответствия
		if *note != newNotes[2] {
			fmt.Print("Ответ не совпадает с ожидаемым\n\n")
		} else {
			fmt.Print("Ожидаемый результат\n\n")
		}
	}

	log.Println("Добавление заметки с уже существующим заголовком (ожидается ошибка)")
	if err := api.AddNote(&Note{Header: "my note 1", Body: "1234"}); err != nil {
		fmt.Printf("Ошибка: %s\n\n", err.Error())
	} else {
		fmt.Print("Ответ не совпадает с ожидаемым\n\n")
	}

	log.Println("Изменение существующей заметки")
	newNote := Note{Header: "my note 1", Body: "new body"}
	if err := api.UpdateNote(&newNote); err != nil {
		fmt.Printf("Ошибка: %s\n", err.Error())
	}
	// проверка изменения
	if note, err := api.FetchNote("my note 1"); err != nil {
		fmt.Printf("Ошибка: %s\n", err.Error())
	} else {
		if *note != newNote {
			fmt.Print("Ответ не совпадает с ожидаемым\n\n")
		} else {
			fmt.Print("Ожидаемый результат\n\n")
		}
	}

	log.Println("Конвертация валют")
	if res, err := api.Convert(2000, "RUB", []string{"EUR", "USD", "CNY", "JPY"}); err != nil {
		fmt.Printf("Ошибка: %s\n", err.Error())
	} else {
		fmt.Println(res)
	}

}
