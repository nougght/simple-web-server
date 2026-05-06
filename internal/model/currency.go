package model

import (
	"fmt"
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
		return fmt.Errorf("invalid base currency")
	}
	for _, c := range p.TargetCurrencies {
		if len(c) != 3 && len(c) != 0 {
			return fmt.Errorf("invalid target currency %s", c)
		}
	}
	return nil
}

type CurrencyRatesResponse struct {
	Data map[string]float64 `json:"data"`
}

type ConvertCurrencyResponse map[string]float64
