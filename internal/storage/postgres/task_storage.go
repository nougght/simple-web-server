package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"simple-server/internal/model"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func NewTaskStorage(db *sqlx.DB) *TaskStorage {
	return &TaskStorage{db: db}
}

type TaskStorage struct {
	db *sqlx.DB
}

func (s *TaskStorage) CreateTask(ctx context.Context, task *model.Task) (*model.Task, error) {
	query := `INSERT INTO tasks(status, type, payload) VALUES($1, $2, $3) 
			  RETURNING id, created_at`

	err := s.db.QueryRowxContext(ctx, query, task.Status, task.Type, task.Payload).Scan(&task.ID, &task.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert failed: %w", err)
	}
	return task, nil
}

func (s *TaskStorage) GetTaskStatus(ctx context.Context, id uuid.UUID) (*model.TaskStatus, error) {
	query := `SELECT status FROM tasks WHERE id = $1`

	var status model.TaskStatus

	if err := s.db.GetContext(ctx, &status, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("select failed: %w", err)
	}
	return &status, nil
}

func (s *TaskStorage) GetTaskByID(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	query := `SELECT * FROM tasks WHERE id = $1`

	var task model.Task

	if err := s.db.GetContext(ctx, &task, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("select failed: %w", err)
	}
	return &task, nil
}

// забирает задачи со статусом pending и меняет их статус на in_progress,
// но не больше переданного лимита
func (s *TaskStorage) GetPendingTasksWithLimit(ctx context.Context, limit uint) ([]model.Task, error) {
	tasks := []model.Task{}
	// забирает определенное количество задач из бд и меняет их статус
	query := `UPDATE tasks
				SET status = $1
				WHERE id IN (
					SELECT id
					FROM tasks
					WHERE status = $2
					ORDER BY created_at ASC
					LIMIT $3
				)
				RETURNING *`

	rows, err := s.db.QueryxContext(ctx, query, model.TaskStatusInProgress, model.TaskStatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %s", err.Error())
		}
	}()
	err = sqlx.StructScan(rows, &tasks)
	if err != nil {
		return nil, fmt.Errorf("struct scan failed: %w", err)
	}
	return tasks, nil
}

func (s *TaskStorage) UpdateTask(ctx context.Context, task *model.Task) error {
	query := `UPDATE tasks SET status = $1, result = $2, error = $3, finished_at = $4
			  WHERE id = $5`

	_, err := s.db.ExecContext(ctx, query, task.Status, task.Result, task.Error, task.FinishedAt, task.ID)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}
	return nil
}

// обновление статусов задач по списку ID
func (s *TaskStorage) UpdateTaskStatuses(ctx context.Context, status model.TaskStatus, ids []uuid.UUID) error {
	query := `UPDATE tasks SET status = $1
			  WHERE id = ANY($2)`
	_, err := s.db.ExecContext(ctx, query, status, pq.Array(ids))
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}
	return nil
}

func (s *TaskStorage) DeleteTask(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tasks WHERE id = $1`

	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	return nil
}
