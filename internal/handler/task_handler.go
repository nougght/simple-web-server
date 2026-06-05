package handler

import (
	"fmt"
	"log"
	"net/http"
	"simple-server/internal/model"

	"github.com/google/uuid"
)

type TaskHandler struct {
	service model.TaskService
}

func NewTaskHandler(service model.TaskService) *TaskHandler {
	return &TaskHandler{
		service: service,
	}
}

func (h *TaskHandler) GetTaskStatus(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		handleError(w, fmt.Errorf("invalid task ID: %w: %w", err, model.ErrBadRequest))
		return
	}

	status, err := h.service.GetTaskStatus(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		handleError(w, fmt.Errorf("invalid task ID: %w: %w", err, model.ErrBadRequest))
		return
	}
	result, err := h.service.GetTaskByID(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		handleError(w, fmt.Errorf("invalid task ID: %w: %w", err, model.ErrBadRequest))
		return
	}
	if err := h.service.DeleteTask(r.Context(), id); err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nil)
}
