package service

import (
	"simple-server/internal/config"
	"simple-server/internal/model"
	"simple-server/internal/service/currency"
	"simple-server/internal/service/note"
	storage "simple-server/internal/storage/memory"
)

type Service struct {
	noteService     *note.NoteService
	currencyService *currency.CurrencyService
}

func GetServices(config *config.Config) *Service {
	return &Service{
		noteService:     note.NewNoteService(config, storage.NewNoteStorage()),
		currencyService: currency.NewCurrencyService(config),
	}
}

func (s *Service) NoteService() model.NoteService {
	return s.noteService
}

func (s *Service) CurrencyService() model.CurrencyService {
	return s.currencyService
}
