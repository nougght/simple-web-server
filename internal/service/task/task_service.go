package task

import (
	"context"
	"errors"
	"fmt"
	"log"
	"simple-server/internal/config"
	"simple-server/internal/model"
	"simple-server/internal/util"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// объединяет модель задачи и функцию для выполнения
type job struct {
	Task model.Task
	Func func(context.Context) (any, error)
}

type TaskService struct {
	config         *config.Config
	storage        TaskStorage
	wg             sync.WaitGroup
	once           sync.Once
	mu             sync.Mutex
	isShuttingDown atomic.Bool
	jobs           chan *job // канал с очередью задач
}

func NewTaskService(cfg *config.Config, storage TaskStorage) *TaskService {
	return &TaskService{config: cfg, storage: storage,
		jobs: make(chan *job, cfg.TaskBufferSize)}
}

// запуск горутин с передачей контекста
func (s *TaskService) StartWorkers(ctx context.Context) {
	for i := 0; i < s.config.TaskWorkersCount; i++ {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()

			// не останавливается при ctx.Done(), чтобы после закрытия канала обработались оставшиеся задачи
			for job := range s.jobs {
				// контекст с таймаутом для задачи
				taskCtx, cancel := context.WithTimeout(ctx, model.TaskTimeout)

				s.executeAndSaveTask(taskCtx, &job.Task, job.Func)
				cancel()

				log.Printf("async task %s completed", job.Task.ID)

			}
		}()
	}
}

// закрытие канала и ожидание завершения задач
//
// если вызван до отмены контекста, задачи будут выполнены в обычном режиме
// если вызван после отмены контекста, задачи будут помечены как cancelled
func (s *TaskService) Stop() {
	s.isShuttingDown.Store(true)
	s.mu.Lock()
	s.once.Do(func() { close(s.jobs) })
	s.mu.Unlock()
	s.wg.Wait()
}

func (s *TaskService) createTask(ctx context.Context, task *model.Task) (*model.Task, error) {
	task, err := s.storage.CreateTask(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}
	return task, nil
}

func (s *TaskService) GetTaskByID(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	task, err := s.storage.GetTaskByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}

func (s *TaskService) GetTaskStatus(ctx context.Context, id uuid.UUID) (*model.TaskStatus, error) {
	status, err := s.storage.GetTaskStatus(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task status: %w", err)
	}
	return status, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, id uuid.UUID) error {
	err := s.storage.DeleteTask(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

func (s *TaskService) updateTask(ctx context.Context, task *model.Task) error {
	err := s.storage.UpdateTask(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	return nil
}

// выполняет задачу и сохраняет результат/ошибку в БД
// если контекст отменен, задача сохраняется со статусом cancelled
func (s *TaskService) executeAndSaveTask(ctx context.Context, task *model.Task,
	taskFunc func(context.Context) (any, error)) {
	log.Printf("starting task %s", task.ID)

	task.Status = model.TaskStatusFailed

	// отдельный контекст для БД, чтобы после отмены основного сохранить результат в БД
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer dbCancel()

	// сохранение в БД
	defer func() {
		if r := recover(); r != nil {
			if task.FinishedAt == nil {
				task.SetFinishedAt(time.Now())
			}
			task.Status = model.TaskStatusFailed
			task.SetError(fmt.Sprintf("task failed with panic: %v", r))
		}

		log.Printf("task %s finished at %v, status - %s", task.ID, *task.FinishedAt, task.Status)
		if task.Error != nil {
			log.Printf("task error: %s", *task.Error)
		}
		err := s.updateTask(dbCtx, task)
		if err != nil {
			log.Printf("failed to update task %s: %s", task.ID, err.Error())
			return
		}
		log.Printf("task %s updated successfully", task.ID)
	}()
	var (
		result  any
		taskErr error
	)

	select {
	// не запускаем задачу если контекст уже отменен
	case <-ctx.Done():
		taskErr = context.Canceled
	default:
		result, taskErr = taskFunc(ctx)
	}
	task.SetFinishedAt(time.Now())

	if taskErr != nil {
		// если контекст отменен, устанавливаем статус cancelled для задачи
		if errors.Is(ctx.Err(), context.Canceled) {
			task.Status = model.TaskStatusCancelled
			task.SetError("task cancelled")
			return
		}
		task.SetError(fmt.Sprintf("task execution failed: %s", taskErr.Error()))
		return
	}

	resultJson, err := util.EncodeJson(result)
	if err != nil {
		task.SetError(fmt.Sprintf("failed to encode result: %s", err.Error()))
		return
	}

	// если все ок - сохраняем результат с успешным статусом
	task.Status = model.TaskStatusSuccess
	task.SetResult(resultJson)
}

// добавляет задачу в очередь
func (s *TaskService) ExecuteAndSaveAsync(ctx context.Context, taskFunc func(context.Context) (any, error)) (uuid.UUID, error) {
	// не принимаем новые задачи после отмены
	if s.isShuttingDown.Load() {
		return uuid.Nil, fmt.Errorf("service is shutting down")
	}

	task := &model.Task{Status: model.TaskStatusInProgress}
	var err error
	task, err = s.createTask(ctx, task)
	if err != nil {
		return uuid.Nil, err
	}

	// проверка и отправка задачи в канал с блокировкой
	s.mu.Lock()
	if s.isShuttingDown.Load() {
		s.mu.Unlock()
		if err := s.DeleteTask(ctx, task.ID); err != nil {
			log.Printf("failed to delete task %s: %s", task.ID, err.Error())
		}
		return uuid.Nil, fmt.Errorf("service is shutting down")
	}

	// отправляем job с задачей и функцией через канал
	// если буффер канала заполнен - удаляем созданную запись и возвращаем ошибку
	select {
	case s.jobs <- &job{Task: *task, Func: taskFunc}:
		s.mu.Unlock()
		log.Printf("task added, buffer size = %d", len(s.jobs))
		return task.ID, nil
	default:
		s.mu.Unlock()
		log.Print("task buffer full")
		if err := s.DeleteTask(ctx, task.ID); err != nil {
			log.Printf("failed to delete task %s: %s", task.ID, err.Error())
		}
		return uuid.Nil, model.ErrTaskBufferFull
	}
}
