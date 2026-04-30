package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"simple_server/handlers"
)

func main() {
	// загрузка переменных окружения
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("Не найден .env файл")
		return
	}
	apiKey, exists := os.LookupEnv("FREECURRENCY_API_KEY")
	if exists == false {
		fmt.Println("Не найден api ключ")
		return
	}

	// обработчик для курсов валют
	currencyHandler := handlers.NewCurencyHandler(apiKey)
	// обработчик для заметок
	notesHandler := handlers.NewNotesHandler()
	// регистрация эндпоинтов
	mux := http.NewServeMux()

	// конвертация валют с использованием внешнего api
	mux.HandleFunc("GET /currency", currencyHandler.ConvertCurrency)

	// получение заметки по уникальному заголовку
	mux.HandleFunc("GET /notes/{header}", notesHandler.GetNoteByHeader)

	// // получение всех заметок
	// mux.HandleFunc("GET /notes", notesHandler.GetAllNotes)

	// создание заметки
	mux.HandleFunc("POST /notes", notesHandler.PostNote)

	// // изменение
	// mux.HandleFunc("PUT /notes/{header}", notesHandler.PutNote)

	// // удаление
	// mux.HandleFunc("DELETE /notes/{header}", notesHandler.DeleteNote)

	
	fmt.Println("Сервер запущен")
	err := http.ListenAndServe(":9000", mux)
	fmt.Println(err)
}
