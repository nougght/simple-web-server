package task

import (
	"context"
	"fmt"
	"log"
	"simple-server/internal/config"
	"simple-server/internal/model"
	"simple-server/internal/util"
	"time"

	"github.com/google/uuid"
)

type TaskService struct {
	config  *config.Config
	storage TaskStorage
}

func NewTaskService(cfg *config.Config, storage TaskStorage) *TaskService {
	return &TaskService{config: cfg, storage: storage}
}

func (s *TaskService) CreateTask(ctx context.Context, task *model.Task) (*model.Task, error) {
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
func (s *TaskService) executeAndSaveTask(ctx context.Context, task *model.Task, taskFunc func(context.Context) (any, error)) {
	log.Printf("starting task %s", task.ID)

	// ограничение времени выполнения задачи
	taskContext, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	result, err := taskFunc(taskContext)
	task.SetFinishedAt(time.Now())

	defer func() {
		log.Printf("task %s finished at %v, status - %s", task.ID, *task.FinishedAt, task.Status)
		if task.Error != nil {
			log.Printf("error: %s", *task.Error)
		}
		if err := s.updateTask(ctx, task); err != nil {
			log.Printf("failed to update task %s: %s", task.ID, err.Error())
			return
		}
		log.Printf("task %s updated successfully", task.ID)
	}()

	if err != nil {
		task.Status = model.TaskStatusFailed
		task.SetError(fmt.Sprintf("task execution failed: %s", err.Error()))
		return
	}

	resultJson, err := util.EncodeJson(result)
	if err != nil {
		task.Status = model.TaskStatusFailed
		task.SetError(fmt.Sprintf("failed to encode result: %s", err.Error()))
		return
	}

	task.Status = model.TaskStatusSuccess
	task.SetResult(resultJson)
}

// запускает задачу в фоне, возвращает её ID
func (s *TaskService) ExecuteAndSaveAsync(ctx context.Context, taskFunc func(context.Context) (any, error)) (uuid.UUID, error) {
	task := &model.Task{Status: model.TaskStatusInProgress}
	var err error
	task, err = s.CreateTask(ctx, task)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create task: %w", err)
	}

	// запукаем задачу в фоне с отдельным контекстом
	go s.executeAndSaveTask(context.Background(), task, taskFunc)
	return task.ID, nil
}
