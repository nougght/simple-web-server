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
	// забирает задачи со статусом pending и меняет их статус на in_progress,
	// но не больше переданного лимита
	GetPendingTasksWithLimit(ctx context.Context, limit uint) ([]model.Task, error)
	UpdateTask(ctx context.Context, task *model.Task) error
	UpdateTaskStatuses(ctx context.Context, status model.TaskStatus, ids []uuid.UUID) error
	DeleteTask(ctx context.Context, id uuid.UUID) error
}
