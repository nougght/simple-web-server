package handler

import (
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
	defer util.CloseRequestBody(r)
	raw, err := io.ReadAll(r.Body)
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
	defer util.CloseRequestBody(r)
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("body reading error: %w", err)
	}
	var updateNote model.UpdateNoteRequestBody
	if err = util.DecodeJson(raw, &updateNote); err != nil {
		return nil, fmt.Errorf("%w: %w", err, model.ErrBadRequest)
	}
	note := &model.Note{
		ID:     ID,
		Header: updateNote.Header,
		Body:   updateNote.Body,
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

	writeJSON(w, http.StatusOK, note)
}

// получение списка заметок c фильтрацией
func (h *NoteHandler) GetNotes(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)
	var (
		notesList []model.Note
		err       error
	)
	filters := make(map[string]interface{})
	if header := r.URL.Query().Get("header"); header != "" {
		filters["header"] = header
	}
	notesList, err = h.service.GetNotes(r.Context(), filters)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, notesList)
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

	writeJSON(w, http.StatusOK, note)
}

// изменение заметки
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

	writeJSON(w, http.StatusOK, nil)
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

	writeJSON(w, http.StatusOK, nil)
}
