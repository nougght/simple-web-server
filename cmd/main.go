package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"simple-server/internal/handler"
	"simple-server/internal/service/currency"
	"simple-server/internal/service/notes"
	"simple-server/internal/storage"
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

	// обработка валют
	currencyService := currency.NewCurrencyService(apiKey)
	currencyHandler := handler.NewCurencyHandler(currencyService)

	// обработка заметок
	storage := storage.NewNotesStorage()
	notesService := notes.NewNotesService(storage)
	notesHandler := handler.NewNotesHandler(notesService)

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
