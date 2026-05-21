package service

import (
	"fmt"
	"simple-server/internal/config"
	"simple-server/internal/model"
	"simple-server/internal/service/currency"
	"simple-server/internal/service/note"
	"simple-server/internal/storage/memory"
	"simple-server/internal/storage/postgres"
)

type Service struct {
	noteService     *note.NoteService
	currencyService *currency.CurrencyService
}

func GetServices(config *config.Config) (*Service, error) {
	var storage note.NoteStorage
	if config.StorageType == "postgres" {
		var err error
		storage, err = postgres.NewNoteStorage(config.Postgres)
		if err != nil {
			return nil, fmt.Errorf("failed initialize postgres storage: %w", err)
		}
	} else {
		storage = memory.NewNoteStorage()
	}
	return &Service{
		noteService:     note.NewNoteService(config, storage),
		currencyService: currency.NewCurrencyService(config),
	}, nil
}

func (s *Service) NoteService() model.NoteService {
	return s.noteService
}

func (s *Service) CurrencyService() model.CurrencyService {
	return s.currencyService
}
