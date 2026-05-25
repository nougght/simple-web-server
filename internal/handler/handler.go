package handler

import (
	"errors"
	"log"
	"net/http"
	"simple-server/internal/model"
)

type Handler struct {
	NoteHandler     *NoteHandler
	CurrencyHandler *CurrencyHandler
}

func (h *Handler) registerRoutes(mux *http.ServeMux) {
	// конвертация валют с использованием внешнего api
	mux.HandleFunc("GET /currency", h.CurrencyHandler.ConvertCurrency)

	// // получение всех заметок
	mux.HandleFunc("GET /note", h.NoteHandler.GetAllNotes)

	// получение заметок с указанным заголовком
	mux.HandleFunc("GET /note/header/{header}", h.NoteHandler.GetNotesByHeader)

	// получение заметки по id
	mux.HandleFunc("GET /note/id/{id}", h.NoteHandler.GetNoteById)

	// создание заметки
	mux.HandleFunc("POST /note", h.NoteHandler.PostNote)
	// изменение
	mux.HandleFunc("PUT /note/{id}", h.NoteHandler.PutNote)
	// удаление
	mux.HandleFunc("DELETE /note/{id}", h.NoteHandler.DeleteNote)
}

func GetHandlers(services model.Service) (*http.ServeMux, *Handler) {
	handler := Handler{
		NoteHandler:     NewNoteHandler(services.NoteService()),
		CurrencyHandler: NewCurrencyHandler(services.CurrencyService()),
	}

	mux := http.NewServeMux()
	handler.registerRoutes(mux)
	return mux, &handler
}

func handleError(w http.ResponseWriter, err error) {
	log.Println(err.Error())
	switch {
	case errors.Is(err, model.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, model.ErrBadRequest):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
