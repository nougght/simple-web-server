package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"simple-server/internal/config"
	"simple-server/internal/model"
	"simple-server/internal/service/currency"
	"simple-server/internal/service/note"
	"simple-server/internal/service/task"
	"simple-server/internal/storage/memory"
	"simple-server/internal/storage/postgres"
)

type Service struct {
	noteService     model.NoteService
	currencyService model.CurrencyService
	taskService     model.TaskService
}

func GetServices(config *config.Config, httpClient *http.Client, rootCtx context.Context) (*Service, error) {
	// общее подключение к БД
	db, err := postgres.ConnectDB(config.Postgres)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	var noteStorage note.NoteStorage
	if config.NoteStorageType == model.StorageTypePostgres {
		var err error
		noteStorage, err = postgres.NewNoteStorage(db)
		if err != nil {
			return nil, fmt.Errorf("failed initialize postgres storage: %w", err)
		}
	} else {
		noteStorage = memory.NewNoteStorage()
	}
	log.Printf("%s note storage initialized", config.NoteStorageType)

	taskStorage := postgres.NewTaskStorage(db)

	taskService := task.NewTaskService(config, taskStorage)
	taskService.StartWorkers(rootCtx)

	return &Service{
		noteService:     note.NewNoteService(config, noteStorage),
		currencyService: currency.NewCurrencyService(config, httpClient, taskService),
		taskService:     taskService,
	}, nil
}

func (s *Service) NoteService() model.NoteService {
	return s.noteService
}

func (s *Service) CurrencyService() model.CurrencyService {
	return s.currencyService
}

func (s *Service) TaskService() model.TaskService {
	return s.taskService
}
