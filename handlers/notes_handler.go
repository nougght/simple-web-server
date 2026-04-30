package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Note struct {
	Header string `json:"header"`
	Body   string `json:"body"`
}

type NotesHandler struct {
	notes map[string]Note
}

func NewNotesHandler() *NotesHandler {
	return &NotesHandler{
		notes: make(map[string]Note),
	}
}

func (h *NotesHandler) PostNote(w http.ResponseWriter, r *http.Request) {
	// читаем тело запроса
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "body reading error", http.StatusBadRequest)
		return
	}
	fmt.Printf("post: \n%s\n", string(rawBody))
	// декодируем тело запроса
	var note Note
	if err := json.Unmarshal(rawBody, &note); err != nil {
		fmt.Println(err)
		http.Error(w, "json decoding error", http.StatusBadRequest)
		return
	}
	// если заголовок уже используется в другой заметке - возвращаем ошибку
	if _, exists := h.notes[note.Header]; exists == true {
		fmt.Println("Заметка с указанным заголовком уже существует\n")
		http.Error(w, fmt.Sprintf("note with header '%s' already exists", note.Header), http.StatusBadRequest)
		return
	}
	// сохраняем новую заметку
	h.notes[note.Header] = note

	fmt.Println("Заметка создана\n")
	w.WriteHeader(http.StatusOK)
}

func (h *NotesHandler) GetNoteByHeader(w http.ResponseWriter, r *http.Request) {
	header := r.PathValue("header")
	fmt.Printf("get: %s\n", header)
	// если заметки с подобным заголовком нет
	if note, exists := h.notes[header]; exists == false {
		fmt.Println("Заметка не найдена\n")
		http.Error(w, fmt.Sprintf("note with header '%s' not found", header), http.StatusNotFound)
		return
	} else {
		// конвертируем структуру в json и отправляем ответ
		jsonResponse, err := json.Marshal(note)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "json encoding error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
		fmt.Println("Ответ успешно отправлен\n")
	}
}
