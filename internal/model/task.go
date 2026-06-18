package model

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// статус задачи
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusSuccess    TaskStatus = "success"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// тип задачи
type TaskType string

const (
	TaskTypeCurrencyConversion TaskType = "currency_conversion"
)

const (
	TaskTimeout             time.Duration = 30 * time.Second
	TaskPollingPeriod       time.Duration = 1 * time.Second
	DefaultTaskWorkersCount               = 10
	DefaultTaskBufferSize                 = 100
)

// обработчик задачи
type TaskHandler func(ctx context.Context, payload json.RawMessage) (any, error)

// асинхронная задача
type Task struct {
	ID     uuid.UUID  `json:"id" db:"id"`
	Status TaskStatus `json:"status" db:"status"`

	// тип задачи
	Type TaskType `json:"type" db:"type"`
	// входные данные
	Payload json.RawMessage `json:"payload" db:"payload"`

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
