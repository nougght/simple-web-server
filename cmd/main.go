package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"simple-server/internal/handler"
	"simple-server/internal/service/currency"
	"simple-server/internal/service/notes"
	"simple-server/internal/storage"

	"github.com/joho/godotenv"
)

func registerRoutes(mux *http.ServeMux, currencyHandler *handler.CurrencyHandler, notesHandler *handler.NotesHandler) {
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
}

func main() {
	// загрузка переменных окружения
	if err := godotenv.Load(); err != nil {
		log.Println(err.Error())
	}
	apiKey, exists := os.LookupEnv("FREECURRENCY_API_KEY")
	if !exists {
		log.Println("Не найден api ключ")
		return
	}

	// обработка валют
	currencyService := currency.NewCurrencyService(apiKey)
	currencyHandler := handler.NewCurrencyHandler(currencyService)

	// обработка заметок
	storage := storage.NewNotesStorage()
	notesService := notes.NewNotesService(storage)
	notesHandler := handler.NewNotesHandler(notesService)

	mux := http.NewServeMux()

	// регистрация эндпоинтов
	registerRoutes(mux, currencyHandler, notesHandler)

	// перехват сигналов завершения работы
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverCtx, stopServer := context.WithCancel(context.Background())
	server := &http.Server{
		Addr: ":8085",
		BaseContext: func(_ net.Listener) context.Context {
			return serverCtx
		},
		Handler: mux,
	}

	go func() {
		log.Println("Сервер запущен")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	// ожидание сигнала завершения работы
	<-rootCtx.Done()
	// выключаем перехват сигналов
	stop()

	log.Println("Остановка сервара, ожидание завершения текущих запросов")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	// отключаем неактивные соединения и ждем завершения запросов
	err := server.Shutdown(shutdownCtx)

	// если запросы не завершились, отменяем контекст сервера
	stopServer()
	if err != nil {
		log.Println("Ошибка при остановке сервера, ожидаем еще 3 секунды")
		time.Sleep(time.Second * 3)
	}
	log.Println("Сервер остановлен")

}
