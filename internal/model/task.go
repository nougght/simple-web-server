package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

// статусы задач
const (
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusSuccess    TaskStatus = "success"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

const (
	TaskTimeout             time.Duration = 30 * time.Second
	DefaultTaskWorkersCount               = 10
	DefaultTaskBufferSize                 = 100
)

// асинхронная задача
type Task struct {
	ID         uuid.UUID        `json:"id" db:"id"`
	Status     TaskStatus       `json:"status" db:"status"`
	Result     *json.RawMessage `json:"result" db:"result"`
	Error      *string          `json:"error" db:"error"`
	CreatedAt  time.Time        `json:"created_at" db:"created_at"`
	FinishedAt *time.Time       `json:"finished_at,omitempty" db:"finished_at"`
}

// методы для заполнения необязательных полей с указателями
func (t *Task) SetError(err string) {
	t.Error = &err
}

func (t *Task) SetFinishedAt(finishedAt time.Time) {
	t.FinishedAt = &finishedAt
}

func (t *Task) SetResult(result json.RawMessage) {
	t.Result = &result
}
