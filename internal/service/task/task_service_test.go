package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"simple-server/internal/config"
	"simple-server/internal/model"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockTaskStorage struct {
	CreateTaskFunc               func(ctx context.Context, task *model.Task) (*model.Task, error)
	GetTaskStatusFunc            func(ctx context.Context, id uuid.UUID) (*model.TaskStatus, error)
	GetTaskByIDFunc              func(ctx context.Context, id uuid.UUID) (*model.Task, error)
	UpdateTaskFunc               func(ctx context.Context, task *model.Task) error
	DeleteTaskFunc               func(ctx context.Context, id uuid.UUID) error
	GetPendingTasksWithLimitFunc func(ctx context.Context, limit uint) ([]model.Task, error)
	UpdateTaskStatusesFunc       func(ctx context.Context, status model.TaskStatus, ids []uuid.UUID) error
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
func (m *MockTaskStorage) GetPendingTasksWithLimit(ctx context.Context, limit uint) ([]model.Task, error) {
	return m.GetPendingTasksWithLimitFunc(ctx, limit)
}
func (m *MockTaskStorage) UpdateTaskStatuses(ctx context.Context, status model.TaskStatus, ids []uuid.UUID) error {
	return m.UpdateTaskStatusesFunc(ctx, status, ids)
}

func NewMockTaskStorage(tasks *sync.Map) *MockTaskStorage {
	return &MockTaskStorage{
		CreateTaskFunc: func(ctx context.Context, task *model.Task) (*model.Task, error) {
			task.ID = uuid.New()
			task.CreatedAt = time.Now()
			tasks.Store(task.ID, task)
			return task, nil
		},
		UpdateTaskFunc: func(ctx context.Context, task *model.Task) error {
			tasks.Store(task.ID, task)
			log.Printf("updating task: %s, %s", task.ID, task.Status)
			return nil
		},
		DeleteTaskFunc: func(ctx context.Context, id uuid.UUID) error {
			tasks.Delete(id)
			return nil
		},
		GetPendingTasksWithLimitFunc: func(ctx context.Context, limit uint) ([]model.Task, error) {
			tasksList := make([]model.Task, 0)
			tasks.Range(func(key, value any) bool {
				if value.(*model.Task).Status != model.TaskStatusPending {
					return true
				}
				tasksList = append(tasksList, *value.(*model.Task))
				return true
			})
			sort.Slice(tasksList, func(i, j int) bool {
				return tasksList[i].CreatedAt.Before(tasksList[j].CreatedAt)
			})
			for i := 0; i < len(tasksList) && i < int(limit); i++ {
				tasksList[i].Status = model.TaskStatusInProgress
				tasks.Store(tasksList[i].ID, &tasksList[i])
			}
			if len(tasksList) < int(limit) {
				return tasksList, nil
			}
			return tasksList[:limit], nil
		},
		UpdateTaskStatusesFunc: func(ctx context.Context, status model.TaskStatus, ids []uuid.UUID) error {
			log.Printf("updating task statuses: %s, %v", status, len(ids))
			for _, id := range ids {
				task, ok := tasks.Load(id)
				if !ok {
					return fmt.Errorf("task not found: %s", id)
				}
				task.(*model.Task).Status = status
				tasks.Store(id, task)
			}
			return nil
		},
	}
}
func TestExecuteAndSave(t *testing.T) {
	timeout := 3 * time.Second
	tests := []struct {
		name               string
		task               model.Task
		taskFunc           model.TaskHandler
		isContextCancelled bool
		expectedStatus     model.TaskStatus
		expectedResult     any
		isErrorExpected    bool
	}{
		{
			name: "too long task",
			task: model.Task{
				ID:     uuid.New(),
				Status: model.TaskStatusPending,
				Type:   "valid_type",
			},
			taskFunc: func(ctx context.Context, payload json.RawMessage) (any, error) {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(timeout + time.Second):
				}
				return "some result", nil
			},
			expectedStatus:  model.TaskStatusFailed,
			isErrorExpected: true,
		},
		{
			name: "default task",
			task: model.Task{
				ID:     uuid.New(),
				Status: model.TaskStatusPending,
				Type:   "valid_type",
			},
			taskFunc: func(ctx context.Context, payload json.RawMessage) (any, error) {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(timeout / 2):
				}
				return "default result", nil
			},
			expectedStatus:  model.TaskStatusSuccess,
			expectedResult:  "default result",
			isErrorExpected: false,
		},
		{
			name: "task with error",
			task: model.Task{
				ID:     uuid.New(),
				Status: model.TaskStatusPending,
				Type:   "valid_type",
			},
			taskFunc: func(ctx context.Context, payload json.RawMessage) (any, error) {
				return nil, assert.AnError
			},
			expectedStatus:  model.TaskStatusFailed,
			isErrorExpected: true,
		},
		{
			name: "task with panic",
			task: model.Task{
				ID:     uuid.New(),
				Status: model.TaskStatusPending,
				Type:   "valid_type",
			},
			taskFunc: func(ctx context.Context, payload json.RawMessage) (any, error) {
				panic("some panic")
			},
			expectedStatus:  model.TaskStatusFailed,
			isErrorExpected: true,
		},
		{
			name: "task with cancelled context",
			task: model.Task{
				ID:     uuid.New(),
				Status: model.TaskStatusPending,
				Type:   "valid_type",
			},
			taskFunc: func(ctx context.Context, payload json.RawMessage) (any, error) {
				return nil, ctx.Err()
			},
			isContextCancelled: true,
			expectedStatus:     model.TaskStatusCancelled,
			isErrorExpected:    true,
		},
		{
			name: "task with invalid type",
			task: model.Task{
				ID:     uuid.New(),
				Status: model.TaskStatusPending,
				Type:   "invalid_type",
			},
			taskFunc: func(ctx context.Context, payload json.RawMessage) (any, error) {
				return nil, nil
			},
			expectedStatus:  model.TaskStatusFailed,
			isErrorExpected: true,
		},
	}

	updatedTasks := sync.Map{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := NewTaskService(&config.Config{}, NewMockTaskStorage(&updatedTasks))
			service.RegisterHandler("valid_type", test.taskFunc)

			taskCtx, cancel := context.WithTimeout(context.Background(), timeout)
			if test.isContextCancelled {
				cancel()
			}
			defer cancel()

			service.executeAndSaveTask(taskCtx, &test.task)
			saved, ok := updatedTasks.Load(test.task.ID)
			require.True(t, ok)
			savedTask := saved.(*model.Task)
			assert.NotNil(t, savedTask.FinishedAt)
			assert.Equal(t, test.expectedStatus, savedTask.Status)
			if test.isErrorExpected {
				assert.Nil(t, savedTask.Result)
				assert.NotNil(t, savedTask.Error)
				return
			}
			assert.Nil(t, savedTask.Error)
			require.NotNil(t, savedTask.Result)
			var result any
			err := json.Unmarshal(*savedTask.Result, &result)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedResult, result)
		})
	}
}

// тестирование выполнения с помощью пула

// выполнение нескольких задач параллельно
func TestExecuteAndSave_parallelTasks(t *testing.T) {
	t.Parallel()
	tasks := sync.Map{}
	storage := NewMockTaskStorage(&tasks)

	workersCount := 2
	taskDuration := 100 * time.Millisecond
	taskCount := 5

	svc := NewTaskService(&config.Config{TaskWorkersCount: workersCount, TaskBufferSize: 10, TaskPollingPeriod: 100 * time.Millisecond}, storage)
	svc.StartWorkers(context.Background())

	// фиксирование максимального количества одновременно работающих задач
	var (
		current, mx int
		mtx         sync.Mutex
	)
	ids := make([]uuid.UUID, 0, taskCount)
	for i := 0; i < taskCount; i++ {
		taskFunc := func(ctx context.Context, payload json.RawMessage) (any, error) {
			mtx.Lock()
			current++
			if current > mx {
				mx = current
			}
			mtx.Unlock()

			select {
			case <-time.After(taskDuration):
			case <-ctx.Done():
			}

			mtx.Lock()
			current--
			mtx.Unlock()
			log.Println("task completed: ", i)
			return i, nil
		}
		taskType := model.TaskType(fmt.Sprintf("task_%d", i))
		svc.RegisterHandler(taskType, taskFunc)
		id, err := svc.ExecuteAndSaveAsync(context.Background(), taskType, nil)
		require.NoError(t, err)
		ids = append(ids, id)
		time.Sleep(1 * time.Millisecond)
	}

	time.Sleep(300 * time.Millisecond)
	svc.Stop()

	mtx.Lock()
	assert.LessOrEqual(t, mx, workersCount, "workers count doesn't match")
	assert.Equal(t, 0, current)
	mtx.Unlock()

	results := make([]bool, taskCount)
	// проверка статусов и результатов задач
	for _, id := range ids {
		val, ok := tasks.Load(id)
		require.True(t, ok)
		task := val.(*model.Task)
		assert.Equal(t, model.TaskStatusSuccess, task.Status)
		require.NotNil(t, task.Result)
		var result int
		err := json.Unmarshal(*task.Result, &result)
		require.NoError(t, err)
		require.Less(t, result, taskCount)
		results[result] = true
	}
	require.NotContains(t, results, false)
}

// отклонение новых задач после остановки сервиса
func TestStop_rejectsAfterStop(t *testing.T) {
	t.Parallel()
	tasks := sync.Map{}
	storage := NewMockTaskStorage(&tasks)

	svc := NewTaskService(&config.Config{TaskWorkersCount: 1, TaskBufferSize: 1, TaskPollingPeriod: 100 * time.Millisecond}, storage)
	svc.RegisterHandler(model.TaskTypeCurrencyConversion, func(ctx context.Context, payload json.RawMessage) (any, error) {
		return "ok", nil
	})
	svc.StartWorkers(context.Background())
	svc.Stop()

	id, err := svc.ExecuteAndSaveAsync(context.Background(), model.TaskTypeCurrencyConversion, nil)
	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, id)
}

// при остановке воркеры отмечают выполняющиеся задачи как cancelled
// а оставшиеся задачи в очереди возвращаются в статус pending
func TestStop_processTasksInBuffer(t *testing.T) {
	t.Parallel()
	tasks := sync.Map{}
	storage := NewMockTaskStorage(&tasks)

	svc := NewTaskService(&config.Config{TaskWorkersCount: 2, TaskBufferSize: 10, TaskPollingPeriod: 100 * time.Millisecond}, storage)
	svc.StartWorkers(context.Background())

	taskCount := 5
	ids := make([]uuid.UUID, 0, taskCount)
	wg := sync.WaitGroup{}
	wg.Add(svc.config.TaskWorkersCount)
	for i := 0; i < taskCount; i++ {
		taskType := model.TaskType(fmt.Sprintf("task_%d", i))
		svc.RegisterHandler(taskType, func(ctx context.Context, payload json.RawMessage) (any, error) {
			// сохраняем задачи выполняемые воркерами
			wg.Done()
			<-ctx.Done()

			return nil, ctx.Err()

		})
		id, err := svc.ExecuteAndSaveAsync(context.Background(), taskType,
			nil)
		require.NoError(t, err)
		ids = append(ids, id)
	}

	wg.Wait()
	svc.Stop()

	var cancelled int
	for _, id := range ids {
		val, ok := tasks.Load(id)
		require.True(t, ok)
		task := val.(*model.Task)

		// если задача выполнялась вокером - она отмечена как cancelled
		// иначе она остается в статусе pending
		switch task.Status {
		case model.TaskStatusCancelled:
			cancelled++
			require.NotNil(t, task.FinishedAt)
			require.NotNil(t, task.Error)
			assert.Contains(t, *task.Error, "cancelled")
			assert.Nil(t, task.Result)
		case model.TaskStatusPending:
			assert.Nil(t, task.FinishedAt)
			assert.Nil(t, task.Error)
			assert.Nil(t, task.Result)
		default:
			t.Fatalf("unexpected task status: %s", task.Status)
		}
	}
	assert.Equal(t, cancelled, svc.config.TaskWorkersCount)
}
