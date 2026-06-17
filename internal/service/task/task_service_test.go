package task

import (
	"context"
	"encoding/json"
	"simple-server/internal/config"
	"simple-server/internal/model"
	"sync"
	"sync/atomic"
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

func NewMockTaskStorage(tasks *sync.Map) *MockTaskStorage {
	return &MockTaskStorage{
		CreateTaskFunc: func(ctx context.Context, task *model.Task) (*model.Task, error) {
			task.ID = uuid.New()
			tasks.Store(task.ID, task)
			return task, nil
		},
		UpdateTaskFunc: func(ctx context.Context, task *model.Task) error {
			tasks.Store(task.ID, task)
			return nil
		},
		DeleteTaskFunc: func(ctx context.Context, id uuid.UUID) error {
			tasks.Delete(id)
			return nil
		},
	}
}
func TestExecuteAndSave(t *testing.T) {
	timeout := 3 * time.Second
	tests := []struct {
		name               string
		task               model.Task
		taskFunc           func(context.Context) (any, error)
		isContextCancelled bool
		expectedStatus     model.TaskStatus
		expectedResult     any
		isErrorExpected    bool
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
			expectedStatus:  model.TaskStatusFailed,
			isErrorExpected: true,
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
			expectedStatus:  model.TaskStatusSuccess,
			expectedResult:  "default result",
			isErrorExpected: false,
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
			expectedStatus:  model.TaskStatusFailed,
			isErrorExpected: true,
		},
		{
			name: "task with panic",
			task: model.Task{
				ID:     uuid.New(),
				Status: model.TaskStatusInProgress,
			},
			taskFunc: func(context.Context) (any, error) {
				panic("some panic")
			},
			expectedStatus:  model.TaskStatusFailed,
			isErrorExpected: true,
		},
		{
			name: "task with cancelled context",
			task: model.Task{
				ID:     uuid.New(),
				Status: model.TaskStatusInProgress,
			},
			taskFunc: func(ctx context.Context) (any, error) {
				return nil, ctx.Err()
			},
			isContextCancelled: true,
			expectedStatus:     model.TaskStatusCancelled,
			isErrorExpected:    true,
		},
	}

	updatedTasks := sync.Map{}
	service := NewTaskService(&config.Config{}, NewMockTaskStorage(&updatedTasks))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			taskCtx, cancel := context.WithTimeout(context.Background(), timeout)
			if test.isContextCancelled {
				cancel()
			}
			defer cancel()

			service.executeAndSaveTask(taskCtx, &test.task, test.taskFunc)
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

	svc := NewTaskService(&config.Config{TaskWorkersCount: workersCount, TaskBufferSize: 10}, storage)
	svc.StartWorkers(context.Background())

	// фиксирование максимального количества одновременно работающих задач
	var (
		current, mx int
		mtx         sync.Mutex
	)
	ids := make([]uuid.UUID, 0, taskCount)
	for i := 0; i < taskCount; i++ {
		taskFunc := func(ctx context.Context) (any, error) {
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
			return i, nil
		}

		id, err := svc.ExecuteAndSaveAsync(context.Background(), taskFunc)
		require.NoError(t, err)
		ids = append(ids, id)
	}

	svc.Stop()

	mtx.Lock()
	assert.LessOrEqual(t, mx, workersCount, "workers count doesn't match")
	assert.Equal(t, 0, current)
	mtx.Unlock()

	// проверка статусов и результатов задач
	for i, id := range ids {
		val, ok := tasks.Load(id)
		require.True(t, ok)
		task := val.(*model.Task)
		assert.Equal(t, model.TaskStatusSuccess, task.Status)
		require.NotNil(t, task.Result)
		var result int
		err := json.Unmarshal(*task.Result, &result)
		require.NoError(t, err)
		assert.Equal(t, i, result)
	}
}

// проверка поведения при заполненном буфере
func TestExecuteAndSaveAsync_bufferFull(t *testing.T) {
	t.Parallel()
	tasks := sync.Map{}
	storage := NewMockTaskStorage(&tasks)

	svc := NewTaskService(&config.Config{TaskWorkersCount: 1, TaskBufferSize: 1}, storage)
	ctx, cancel := context.WithCancel(context.Background())
	svc.StartWorkers(ctx)

	started := make(chan struct{})
	defer svc.Stop()

	// долгая задача занимает воркера
	_, err := svc.ExecuteAndSaveAsync(context.Background(), func(ctx context.Context) (any, error) {
		close(started)
		<-ctx.Done()
		return "ok", nil
	})
	require.NoError(t, err)
	<-started

	_, err = svc.ExecuteAndSaveAsync(context.Background(), func(ctx context.Context) (any, error) {
		return "ok", nil
	})
	require.NoError(t, err)

	// третья задача не помещается в буфер
	id, err := svc.ExecuteAndSaveAsync(context.Background(), func(ctx context.Context) (any, error) {
		return "ok", nil
	})
	require.ErrorIs(t, err, model.ErrTaskBufferFull)
	assert.Equal(t, uuid.Nil, id)
	cancel()
}

// отклонение новых задач после остановки сервиса
func TestStop_rejectsAfterStop(t *testing.T) {
	t.Parallel()
	tasks := sync.Map{}
	storage := NewMockTaskStorage(&tasks)

	svc := NewTaskService(&config.Config{TaskWorkersCount: 1, TaskBufferSize: 1}, storage)
	svc.StartWorkers(context.Background())
	svc.Stop()

	id, err := svc.ExecuteAndSaveAsync(context.Background(), func(ctx context.Context) (any, error) {
		return "ok", nil
	})
	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, id)
}

// при остановке воркеры дорабатывают оставшиеся в очереди задачи
func TestStop_processTasksInBuffer(t *testing.T) {
	t.Parallel()
	tasks := sync.Map{}
	storage := NewMockTaskStorage(&tasks)
	var processed atomic.Int32

	svc := NewTaskService(&config.Config{TaskWorkersCount: 2, TaskBufferSize: 10}, storage)
	svc.StartWorkers(context.Background())

	taskCount := 9
	ids := make([]uuid.UUID, 0, taskCount)
	for i := 0; i < taskCount; i++ {
		id, err := svc.ExecuteAndSaveAsync(context.Background(), func(ctx context.Context) (any, error) {
			processed.Add(1)
			return i, nil
		})
		require.NoError(t, err)
		ids = append(ids, id)
	}

	svc.Stop()

	assert.Equal(t, int32(taskCount), processed.Load())
	for i, id := range ids {
		val, ok := tasks.Load(id)
		require.True(t, ok)
		task := val.(*model.Task)
		assert.Equal(t, model.TaskStatusSuccess, task.Status)
		require.NotNil(t, task.Result)
		var result int
		err := json.Unmarshal(*task.Result, &result)
		require.NoError(t, err)
		assert.Equal(t, i, result)
	}
}

// задачи в очереди и выполняемые задачи получают статус cancelled при отмене контекста воркеров
func TestStop_cancelAndStop(t *testing.T) {
	t.Parallel()
	tasks := sync.Map{}
	storage := NewMockTaskStorage(&tasks)

	// 1 воркер — чтобы в буфере всегда были ожидающие задачи
	svc := NewTaskService(&config.Config{TaskWorkersCount: 1, TaskBufferSize: 10}, storage)
	ctx, cancel := context.WithCancel(context.Background())
	svc.StartWorkers(ctx)

	taskCount := 5
	ids := make([]uuid.UUID, 0, taskCount)
	for i := 0; i < taskCount; i++ {
		id, err := svc.ExecuteAndSaveAsync(context.Background(), func(ctx context.Context) (any, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		})
		require.NoError(t, err)
		ids = append(ids, id)
	}

	cancel()
	svc.Stop()

	for _, id := range ids {
		val, ok := tasks.Load(id)
		require.True(t, ok)
		task := val.(*model.Task)
		assert.Equal(t, model.TaskStatusCancelled, task.Status)
		assert.Nil(t, task.Result)
		require.NotNil(t, task.Error)
		assert.Contains(t, *task.Error, "cancelled")
	}
}
