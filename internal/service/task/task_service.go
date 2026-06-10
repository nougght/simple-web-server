package task

import (
	"context"
	"fmt"
	"log"
	"simple-server/internal/config"
	"simple-server/internal/model"
	"simple-server/internal/util"
	"sync"
	"time"

	"github.com/google/uuid"
)

type TaskService struct {
	config  *config.Config
	storage TaskStorage
	wg      *sync.WaitGroup
}

func NewTaskService(cfg *config.Config, storage TaskStorage, wg *sync.WaitGroup) *TaskService {
	return &TaskService{config: cfg, storage: storage, wg: wg}
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

	result, taskErr := taskFunc(ctx)
	task.SetFinishedAt(time.Now())

	if taskErr != nil {
		// если контекст отменен, устанавливаем статус cancelled для задачи
		if ctx.Err() == context.Canceled {
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

// запускает задачу в фоне с таймаутом, возвращает её ID
//
// при отмене rootCtx задачи сохраняется в БД со статусом cancelled, новые задачи не принимаются,
// клиент чаще всего не будет успевать получить статус отмененной задачи
func (s *TaskService) ExecuteAndSaveAsync(ctx context.Context, rootCtx context.Context, taskFunc func(context.Context) (any, error)) (uuid.UUID, error) {
	// не принимаем новые задачи после отмены
	select {
	case <-rootCtx.Done():
		return uuid.Nil, fmt.Errorf("service is shutting down")
	default:
	}

	task := &model.Task{Status: model.TaskStatusInProgress}
	var err error
	task, err = s.createTask(ctx, task)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create task: %w", err)
	}

	s.wg.Add(1)
	// запукаем задачу в фоне c отслеживанием сигнала отмены
	go func() {
		defer s.wg.Done()

		// контекст с таймаутом для задачи на основе rootCtx
		taskCtx, cancel := context.WithTimeout(rootCtx, 7*time.Second)
		defer cancel()

		s.executeAndSaveTask(taskCtx, task, taskFunc)
		log.Printf("async task %s completed", task.ID)
	}()
	return task.ID, nil
}
