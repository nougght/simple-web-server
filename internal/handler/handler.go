package handler

import (
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

	// получение заметки по уникальному заголовку
	mux.HandleFunc("GET /notes/{header}", h.NoteHandler.GetNoteByHeader)
	// // получение всех заметок
	mux.HandleFunc("GET /notes", h.NoteHandler.GetAllNotes)
	// создание заметки
	mux.HandleFunc("POST /notes", h.NoteHandler.PostNote)
	// изменение
	mux.HandleFunc("PUT /notes/{header}", h.NoteHandler.PutNote)
	// удаление
	mux.HandleFunc("DELETE /notes/{header}", h.NoteHandler.DeleteNote)
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
