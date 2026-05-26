package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"simple-server/internal/model"
	"simple-server/internal/util"

	"github.com/google/uuid"
)

type NoteHandler struct {
	service model.NoteService
}

func NewNoteHandler(service model.NoteService) *NoteHandler {
	return &NoteHandler{
		service: service,
	}
}

func (h *NoteHandler) parseNoteFromRequest(r *http.Request) (*model.Note, error) {
	raw, err := io.ReadAll(r.Body)
	_ = r.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("body reading error: %w", err)
	}
	var note model.Note
	if err = util.DecodeJson(raw, &note); err != nil {
		return nil, fmt.Errorf("%w: %w", err, model.ErrBadRequest)
	}
	return &note, nil
}

func (h *NoteHandler) parsePutNoteRequest(r *http.Request) (*model.Note, error) {
	ID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return nil, fmt.Errorf("invalid note id: %w: %w", err, model.ErrBadRequest)
	}
	note, err := h.parseNoteFromRequest(r)
	if err != nil {
		return nil, err
	}

	// если id из URL и тела не совпадают
	if ID != note.ID {
		return nil, fmt.Errorf("id in URL '%s' doesn't match id in body'%s': %w", ID, note.ID, model.ErrBadRequest)
	}

	return note, nil
}

// Добавление заметки
func (h *NoteHandler) PostNote(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)

	note, err := h.parseNoteFromRequest(r)
	if err != nil {
		handleError(w, err)
		return
	}

	note, err = h.service.AddNote(r.Context(), note)
	if err != nil {
		handleError(w, err)
		return
	}

	jsonResponse, err := json.Marshal(note)
	if err != nil {
		handleError(w, fmt.Errorf("json encoding error: %w", err))
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(jsonResponse); err == nil {
		log.Print("Ответ успешно отправлен\n\n")
	} else {
		log.Println(err.Error())
	}
}

// получение списка заметок
func (h *NoteHandler) GetNotes(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)
	var (
		notesList []model.Note
		err       error
	)
	if header := r.URL.Query().Get("header"); header != "" {
		notesList, err = h.service.GetNotesByHeader(r.Context(), header)
	} else {
		notesList, err = h.service.GetAllNotes(r.Context())
	}
	if err != nil {
		handleError(w, err)
		return
	}

	jsonResponse, err := json.Marshal(notesList)
	if err != nil {
		handleError(w, fmt.Errorf("json encoding error: %w", err))
		return
	}

	if _, err := w.Write(jsonResponse); err == nil {
		log.Print("Ответ успешно отправлен\n\n")
	} else {
		log.Print(err.Error() + "\n\n")
	}
}

// получение заметки по его ID
func (h *NoteHandler) GetNoteByID(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)

	ID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		handleError(w, fmt.Errorf("invalid note ID: %w: %w", err, model.ErrBadRequest))
		return
	}

	note, err := h.service.GetNoteByID(r.Context(), ID)
	if err != nil {
		handleError(w, err)
		return
	}

	// конвертируем структуру в json и отправляем ответ
	jsonResponse, err := json.Marshal(note)
	if err != nil {
		handleError(w, fmt.Errorf("json encoding error: %w", err))
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(jsonResponse); err == nil {
		log.Print("Ответ успешно отправлен\n\n")
	} else {
		log.Print(err.Error() + "\n\n")
	}
}

// Изменение заметки по его заголовку
func (h *NoteHandler) PutNote(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)

	note, err := h.parsePutNoteRequest(r)
	if err != nil {
		handleError(w, err)
		return
	}
	// если все ок - обновляем заметку
	if err := h.service.UpdateNote(r.Context(), note); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Print("Заметка обновлена\n\n")
}

// удаление заметки
func (h *NoteHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)
	ID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		handleError(w, fmt.Errorf("invalid note ID: %w: %w", err, model.ErrBadRequest))
		return
	}

	if err := h.service.DeleteNote(r.Context(), ID); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Print("Заметка удалена\n\n")
}
