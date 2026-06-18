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

type TaskService struct {
	config         *config.Config
	storage        TaskStorage
	wg             sync.WaitGroup
	once           sync.Once
	isShuttingDown atomic.Bool
	tasks          chan *model.Task // канал с очередью задач
	handlers       sync.Map         // функции-обработчики для разных типов задач
	pollerDone     chan struct{}
	cancel         context.CancelFunc
}

func NewTaskService(cfg *config.Config, storage TaskStorage) *TaskService {
	return &TaskService{config: cfg, storage: storage,
		tasks:      make(chan *model.Task, cfg.TaskBufferSize),
		handlers:   sync.Map{},
		pollerDone: make(chan struct{}),
	}
}

// список задач из канала
func (s *TaskService) getTasksFromChannel() []model.Task {
	var tasks []model.Task
	for len(s.tasks) > 0 {
		task := <-s.tasks
		tasks = append(tasks, *task)
	}
	return tasks
}

// обновляет статусы оставшихся задач в БД
func (s *TaskService) updateRemainingTasks(unsent []model.Task) {
	ids := []uuid.UUID{}
	tasks := append(unsent, s.getTasksFromChannel()...)
	for _, task := range tasks {
		ids = append(ids, task.ID)
	}
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer dbCancel()

	s.updateTaskStatuses(dbCtx, model.TaskStatusPending, ids)
}

// периодический опрос БД на наличие новых задач с учетом доступного места в буфере
// при shutdown помечает задачи, которые остались в канале и в списке новых задач как pending
func (s *TaskService) startPolling(ctx context.Context, period time.Duration, pollerDone chan struct{}) {
	ticker := time.NewTicker(period)
	go func() {
		defer close(pollerDone)
		for {
			select {
			case <-ticker.C:
				queueSize := len(s.tasks)
				newTasks, err := s.storage.GetPendingTasksWithLimit(ctx, uint(s.config.TaskBufferSize-queueSize))
				if err != nil {
					log.Printf("failed to get pending tasks: %s", err.Error())
					continue
				}

				for i := range newTasks {
					select {
					// может заблокироваться если количество задач в канале увеличилось
					case s.tasks <- &newTasks[i]:
					case <-ctx.Done():
						// возвращаем оставшиеся задачи в статус pending
						ticker.Stop()
						s.updateRemainingTasks(newTasks[i:])
						return
					}
				}
				if len(newTasks) > 0 {
					log.Printf("queue size = %d, found %d new tasks", queueSize, len(newTasks))
				}
			case <-ctx.Done():
				ticker.Stop()
				s.updateRemainingTasks([]model.Task{})
				return
			}
		}
	}()
}

// регистрация функции-обработчика для конкретного типа задачи
func (s *TaskService) RegisterHandler(taskType model.TaskType, handler model.TaskHandler) {
	s.handlers.Store(taskType, handler)
}

// запуск горутин с передачей контекста
func (s *TaskService) StartWorkers(ctx context.Context) {
	ctx, s.cancel = context.WithCancel(ctx)
	s.startPolling(ctx, s.config.TaskPollingPeriod, s.pollerDone)

	for i := 0; i < s.config.TaskWorkersCount; i++ {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case task, ok := <-s.tasks:
					if !ok {
						return
					}
					// контекст с таймаутом для задачи
					taskCtx, cancel := context.WithTimeout(ctx, model.TaskTimeout)
					s.executeAndSaveTask(taskCtx, task)
					cancel()
					log.Printf("async task %s completed", task.ID)
				}
			}
		}()
	}
}

// остановка сервиса и завершение выполнения задач
//
// отменяет контекст(дочерний от rootCtx) и закрывает канал с задачами;
// задачи в очереди возвращаются в статус pending;
// задачи в воркерах помечаются как cancelled
func (s *TaskService) Stop() {
	s.isShuttingDown.Store(true)
	s.once.Do(func() {
		s.cancel()
		// закрываем канал после завершения записывающей горутины
		<-s.pollerDone
		close(s.tasks)
	})
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

func (s *TaskService) updateTaskStatuses(ctx context.Context, status model.TaskStatus, tasks []uuid.UUID) {

	err := s.storage.UpdateTaskStatuses(ctx, status, tasks)
	if err != nil {
		log.Printf("failed to update task statuses: %s", err.Error())
	}
}

// выполняет задачу и сохраняет результат/ошибку в БД
// если контекст отменен, задача сохраняется со статусом cancelled
func (s *TaskService) executeAndSaveTask(ctx context.Context, task *model.Task) {
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

	val, handlerExists := s.handlers.Load(task.Type)
	taskHandler, ok := val.(model.TaskHandler)
	if !handlerExists || !ok {
		task.SetFinishedAt(time.Now())
		task.SetError(fmt.Sprintf("invalid task type: %s", task.Type))
		log.Println(*task.Error)
		return
	}

	select {
	// не запускаем задачу если контекст уже отменен
	case <-ctx.Done():
		taskErr = context.Canceled
	default:
		result, taskErr = taskHandler(ctx, task.Payload)
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

// добавляет новую задачу в БД, возвращает ее ID
func (s *TaskService) ExecuteAndSaveAsync(ctx context.Context, taskType model.TaskType, payload any) (uuid.UUID, error) {
	// не принимаем новые задачи после отмены
	if s.isShuttingDown.Load() {
		return uuid.Nil, fmt.Errorf("service is shutting down")
	}

	// проверка наличия обработчика для типа задачи
	if _, ok := s.handlers.Load(taskType); !ok {
		return uuid.Nil, fmt.Errorf("invalid task type '%s': %w", taskType, model.ErrBadRequest)
	}

	payloadJson, err := util.EncodeJson(payload)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to encode payload: %w", err)
	}

	task := &model.Task{Status: model.TaskStatusPending, Type: taskType, Payload: payloadJson}
	task, err = s.createTask(ctx, task)
	if err != nil {
		return uuid.Nil, err
	}

	return task.ID, nil
}
