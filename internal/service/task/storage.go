package task

import (
	"context"

	"simple-server/internal/model"

	"github.com/google/uuid"
)

type TaskStorage interface {
	CreateTask(ctx context.Context, task *model.Task) (*model.Task, error)
	GetTaskStatus(ctx context.Context, id uuid.UUID) (*model.TaskStatus, error)
	GetTaskByID(ctx context.Context, id uuid.UUID) (*model.Task, error)
	UpdateTask(ctx context.Context, task *model.Task) error
	DeleteTask(ctx context.Context, id uuid.UUID) error
}
