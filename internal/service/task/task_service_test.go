package task

import (
	"context"
	"encoding/json"
	"simple-server/internal/model"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockTaskStorage struct {
	CreateTaskFunc    func(ctx context.Context, task *model.Task) (*model.Task, error)
	GetTaskStatusFunc func(ctx context.Context, id uuid.UUID) (*model.TaskStatus, error)
	GetTaskByIDFunc   func(ctx context.Context, id uuid.UUID) (*model.Task, error)
	UpdateTaskFunc    func(ctx context.Context, task *model.Task) error
	DeleteTaskFunc    func(ctx context.Context, id uuid.UUID) error
}

func (m *MockTaskStorage) CreateTask(ctx context.Context, task *model.Task) (*model.Task, error) {
	return m.CreateTaskFunc(ctx, task)
}
func (m *MockTaskStorage) GetTaskStatus(ctx context.Context, id uuid.UUID) (*model.TaskStatus, error) {
	return m.GetTaskStatusFunc(ctx, id)
}
func (m *MockTaskStorage) GetTaskByID(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	return m.GetTaskByIDFunc(ctx, id)
}
func (m *MockTaskStorage) UpdateTask(ctx context.Context, task *model.Task) error {
	return m.UpdateTaskFunc(ctx, task)
}
func (m *MockTaskStorage) DeleteTask(ctx context.Context, id uuid.UUID) error {
	return m.DeleteTaskFunc(ctx, id)
}

func TestExecuteAndSave(t *testing.T) {
	timeout := 3 * time.Second
	tests := []struct {
		name          string
		task          model.Task
		taskFunc      func(context.Context) (any, error)
		result        any
		errorExpected bool
	}{
		{
			name: "too long task",
			task: model.Task{
				ID:     uuid.New(),
				Status: model.TaskStatusInProgress,
			},
			taskFunc: func(ctx context.Context) (any, error) {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(timeout + time.Second):
				}
				return "some result", nil
			},
			result:        "some result",
			errorExpected: true,
		},
		{
			name: "default task",
			task: model.Task{
				ID:     uuid.New(),
				Status: model.TaskStatusInProgress,
			},
			taskFunc: func(ctx context.Context) (any, error) {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(timeout / 2):
				}
				return "default result", nil
			},
			result:        "default result",
			errorExpected: false,
		},
		{
			name: "task with error",
			task: model.Task{
				ID:     uuid.New(),
				Status: model.TaskStatusInProgress,
			},
			taskFunc: func(ctx context.Context) (any, error) {
				return nil, assert.AnError
			},
			result:        nil,
			errorExpected: true,
		},
	}

	updatedTasks := make(map[uuid.UUID]*model.Task)
	service := NewTaskService(nil, &MockTaskStorage{
		// сохраняем обновленные задачи для проверки
		UpdateTaskFunc: func(ctx context.Context, task *model.Task) error {
			updatedTasks[task.ID] = task
			return nil
		},
	}, context.Background(), &sync.WaitGroup{})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			taskCtx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			service.executeAndSaveTask(taskCtx, &test.task, test.taskFunc)
			savedTask := updatedTasks[test.task.ID]
			assert.NotNil(t, savedTask.FinishedAt)
			if test.errorExpected {
				assert.Equal(t, model.TaskStatusFailed, savedTask.Status)
				assert.Nil(t, savedTask.Result)
				assert.NotNil(t, savedTask.Error)
				return
			}
			assert.Equal(t, model.TaskStatusSuccess, savedTask.Status)
			assert.Nil(t, savedTask.Error)
			require.NotNil(t, savedTask.Result)
			var result any
			err := json.Unmarshal(*savedTask.Result, &result)
			assert.NoError(t, err)
			assert.Equal(t, test.result, result.(string))
		})
	}
}
