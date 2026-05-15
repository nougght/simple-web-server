package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"simple-server/internal/model"
	"simple-server/internal/util"
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
	if err != nil {
		return nil, fmt.Errorf("body reading error")
	}
	var note model.Note
	if err = util.DecodeJson(raw, &note); err != nil {
		return nil, err
	}
	return &note, nil
}

// Добавление заметки
func (h *NoteHandler) PostNote(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)

	note, err := h.parseNoteFromRequest(r)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = h.service.AddNote(note); err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Print("Заметка создана\n\n")
	w.WriteHeader(http.StatusOK)
}

// получение списка заметок
func (h *NoteHandler) GetAllNotes(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)

	notesList := h.service.GetAllNotes()

	jsonResponse, err := json.Marshal(notesList)
	if err != nil {
		http.Error(w, "json encoding error", http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	if _, err := w.Write(jsonResponse); err == nil {
		log.Print("Ответ успешно отправлен\n\n")
	} else {
		log.Print(err.Error() + "\n\n")
	}
}

// получение заметки по его заголовку
func (h *NoteHandler) GetNoteByHeader(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)

	header := r.PathValue("header")

	note, err := h.service.GetNoteByHeader(header)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// конвертируем структуру в json и отправляем ответ
	jsonResponse, err := json.Marshal(note)
	if err != nil {
		log.Print(err.Error() + "\n\n")
		http.Error(w, "json encoding error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(jsonResponse); err == nil {
		log.Print("Ответ успешно отправлен\n\n")
	} else {
		log.Print(err.Error() + "\n\n")
	}

}

func (h *NoteHandler) parsePutNoteRequest(r *http.Request) (*model.Note, error) {
	header := r.PathValue("header")
	note, err := h.parseNoteFromRequest(r)
	if err != nil {
		return nil, err
	}

	// если заголовок из URL и тела не совпадает
	if header != note.Header {
		return nil, fmt.Errorf("header in URL '%s' doesn't match header in body '%s'", header, note.Header)
	}

	return note, nil
}

// Изменение заметки по его заголовку
func (h *NoteHandler) PutNote(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)

	note, err := h.parsePutNoteRequest(r)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// если все ок - обновляем заметку
	if err := h.service.UpdateNote(note); err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Print("Заметка обновлена\n\n")
}

// удаление заметки
func (h *NoteHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)
	header := r.PathValue("header")

	if err := h.service.DeleteNote(header); err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Print("Заметка удалена\n\n")
}
