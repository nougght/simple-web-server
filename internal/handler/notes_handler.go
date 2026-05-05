package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"simple-server/internal/model"
)

type NotesHandler struct {
	notes map[string]model.Note
}

func NewNotesHandler() *NotesHandler {
	return &NotesHandler{
		notes: make(map[string]model.Note),
	}
}

// Добавление заметки
func (h *NotesHandler) PostNote(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)

	// читаем тело запроса
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Print(err.Error() + "\n\n")
		http.Error(w, "body reading error", http.StatusBadRequest)
		return
	}
	// декодируем тело запроса
	var note model.Note
	if err := json.Unmarshal(rawBody, &note); err != nil {
		log.Print(err.Error() + "\n\n")
		http.Error(w, "json decoding error", http.StatusBadRequest)
		return
	}
	// если заголовок уже используется в другой заметке - возвращаем ошибку
	if _, exists := h.notes[note.Header]; exists == true {
		log.Print("Заметка с указанным заголовком уже существует\n\n")
		http.Error(w, fmt.Sprintf("note with header '%s' already exists", note.Header), http.StatusBadRequest)
		return
	}
	// сохраняем новую заметку
	h.notes[note.Header] = note

	log.Print("Заметка создана\n\n")
	w.WriteHeader(http.StatusOK)
}

// получение списка заметок
func (h *NotesHandler) GetAllNotes(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)

	// заполняем список значениями из map
	notesList := make([]model.Note, 0, len(h.notes))
	for _, note := range h.notes {
		notesList = append(notesList, note)
	}

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
func (h *NotesHandler) GetNoteByHeader(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)

	header := r.PathValue("header")
	// если заметки с подобным заголовком нет
	note, exists := h.notes[header]
	if exists == false {
		log.Print("Заметка не найдена\n\n")
		http.Error(w, fmt.Sprintf("note with header '%s' not found", header), http.StatusNotFound)
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

// Изменение заметки по его заголовку
func (h *NotesHandler) PutNote(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)

	header := r.PathValue("header")
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		log.Print(err.Error() + "\n\n")
		http.Error(w, "body reading error", http.StatusBadRequest)
		return
	}

	var note model.Note
	if err := json.Unmarshal(raw, &note); err != nil {
		log.Print(err.Error() + "\n\n")
		http.Error(w, "json decoding error", http.StatusBadRequest)
		return
	}

	// если заголовок в URL и теле запроса не совпадает - возвращаем ошибку
	if header != note.Header {
		log.Println("Заголовк в URL не совпадает с телом")
		http.Error(w, fmt.Sprintf("header in URL '%s' doesn't match header in body '%s'", header, note.Header), http.StatusBadRequest)
		return
	}
	// если все ок - обновляем заметку
	h.notes[header] = note

	w.WriteHeader(http.StatusOK)
	log.Print("Заметка обновлена\n\n")
}

// удаление заметки
func (h *NotesHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)

	header := r.PathValue("header")

	// удаляем заметку по ключу из map
	delete(h.notes, header)

	w.WriteHeader(http.StatusOK)
	log.Print("Заметка удалена\n\n")
}
