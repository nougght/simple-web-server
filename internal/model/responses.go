package model

type CurrencyRatesResponse struct {
	Data map[string]float64 `json:"data"`
}

type ConvertCurrencyResponse map[string]float64
