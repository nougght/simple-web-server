package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"simple-server/handlers"
)

func main() {
	// загрузка переменных окружения
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Не найден .env файл")
		return
	}
	apiKey, exists := os.LookupEnv("FREECURRENCY_API_KEY")
	if exists == false {
		log.Println("Не найден api ключ")
		return
	}

	// обработчик для курсов валют
	currencyHandler := handlers.NewCurencyHandler(apiKey)
	// обработчик для заметок
	notesHandler := handlers.NewNotesHandler()

	mux := http.NewServeMux()

	// регистрация эндпоинтов

	// конвертация валют с использованием внешнего api
	mux.HandleFunc("GET /currency", currencyHandler.ConvertCurrency)

	// получение заметки по уникальному заголовку
	mux.HandleFunc("GET /notes/{header}", notesHandler.GetNoteByHeader)
	// // получение всех заметок
	mux.HandleFunc("GET /notes", notesHandler.GetAllNotes)
	// создание заметки
	mux.HandleFunc("POST /notes", notesHandler.PostNote)
	// изменение
	mux.HandleFunc("PUT /notes/{header}", notesHandler.PutNote)
	// удаление
	mux.HandleFunc("DELETE /notes/{header}", notesHandler.DeleteNote)

	log.Println("Сервер запущен")
	err := http.ListenAndServe(":8080", mux)
	log.Println(err)
}
