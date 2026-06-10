package model

import (
	"fmt"

	"github.com/google/uuid"
)

type ConvertCurrencyParams struct {
	Amount           float64  // сумма конвертации
	BaseCurrency     string   // исходная валюта
	TargetCurrencies []string // список запрашиваемых валют
}

// проверка на корректность параметров
func (p *ConvertCurrencyParams) Validate() error {
	if p.Amount < 0 {
		return fmt.Errorf("negative amount")
	}
	if len(p.BaseCurrency) != 3 && len(p.BaseCurrency) != 0 {
		return fmt.Errorf("invalid base currency: %w", ErrBadRequest)
	}
	for _, c := range p.TargetCurrencies {
		if len(c) != 3 {
			return fmt.Errorf("invalid target currency %s: %w", c, ErrBadRequest)
		}
	}
	return nil
}

// ответ внешнего API
type CurrencyRatesResponse struct {
	Data map[string]float64 `json:"data"`
}

// результат конвертации
type ConvertCurrencyResult map[string]float64

// ответ с указанием ID задачи
type ConvertCurrencyResponse struct {
	TaskID uuid.UUID `json:"task_id"`
}
