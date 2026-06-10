package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"simple-server/internal/model"
	"simple-server/internal/util"
)

type Handler struct {
	NoteHandler     *NoteHandler
	TaskHandler     *TaskHandler
	CurrencyHandler *CurrencyHandler
}

func (h *Handler) registerRoutes(mux *http.ServeMux) {
	// конвертация валют с использованием внешнего api
	mux.HandleFunc("GET /currency", h.CurrencyHandler.ConvertCurrency)

	// получение cписка заметок
	mux.HandleFunc("GET /notes", h.NoteHandler.GetNotes)

	// получение заметки по ID
	mux.HandleFunc("GET /note/{id}", h.NoteHandler.GetNoteByID)
	// создание заметки
	mux.HandleFunc("POST /note", h.NoteHandler.PostNote)
	// изменение
	mux.HandleFunc("PUT /note/{id}", h.NoteHandler.PutNote)
	// удаление
	mux.HandleFunc("DELETE /note/{id}", h.NoteHandler.DeleteNote)

	mux.HandleFunc("GET /task/{id}/status", h.TaskHandler.GetTaskStatus)
	mux.HandleFunc("GET /task/{id}", h.TaskHandler.GetTask)
	mux.HandleFunc("DELETE /task/{id}", h.TaskHandler.DeleteTask)
}

func GetHandlers(services model.Service, rootCtx context.Context) (*http.ServeMux, *Handler) {
	handler := Handler{
		NoteHandler:     NewNoteHandler(services.NoteService()),
		TaskHandler:     NewTaskHandler(services.TaskService()),
		CurrencyHandler: NewCurrencyHandler(services.CurrencyService(), rootCtx),
	}

	mux := http.NewServeMux()
	handler.registerRoutes(mux)
	return mux, &handler
}

func handleError(w http.ResponseWriter, err error) {
	log.Printf("error: %s", err.Error())

	switch {
	case errors.Is(err, model.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, model.ErrBadRequest):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// отправка JSON ответа с передачей статуса и тела(может быть nil)
func writeJSON(w http.ResponseWriter, status int, body any) {
	if body != nil {
		jsonResponse, err := util.EncodeJson(body)
		if err != nil {
			handleError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if _, err := w.Write(jsonResponse); err != nil {
			log.Println(err.Error())
			return
		}
	} else {
		w.WriteHeader(status)
	}
	log.Println("Ответ успешно отправлен")
}
