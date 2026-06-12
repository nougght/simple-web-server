package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"simple-server/internal/model"
	"strconv"
)

// используется бесплатный api для получения курса валют - https://freecurrencyapi.com/docs/

type CurrencyHandler struct {
	service model.CurrencyService
	rootCtx context.Context
}

func NewCurrencyHandler(service model.CurrencyService, rootCtx context.Context) *CurrencyHandler {
	return &CurrencyHandler{
		service: service,
		rootCtx: rootCtx,
	}
}

// извлечение параметров из запроса
func (h *CurrencyHandler) parseConvertParameters(r *http.Request) (*model.ConvertCurrencyParams, error) {
	// параметры по умолчанию
	params := &model.ConvertCurrencyParams{
		Amount:           1.0,
		BaseCurrency:     "RUB",
		TargetCurrencies: []string{},
	}
	str := r.URL.Query().Get("amount")
	// если параметр не пуст - обрабатываем его
	if str != "" {
		if val, err := strconv.ParseFloat(str, 64); err != nil {
			return nil, fmt.Errorf("'amount' parsing failed: %w: %w", err, model.ErrBadRequest)
		} else {
			params.Amount = val
		}
	}
	params.BaseCurrency = r.URL.Query().Get("base")
	params.TargetCurrencies = r.URL.Query()["currencies"]

	return params, nil
}

// обработка запроса конвертации валют
func (h *CurrencyHandler) ConvertCurrency(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)

	params, err := h.parseConvertParameters(r)
	if err != nil {
		handleError(w, err)
		return
	}
	if err := params.Validate(); err != nil {
		handleError(w, err)
		return
	}

	var result model.ConvertCurrencyResponse
	result.TaskID, err = h.service.ConvertAndSaveAsync(r.Context(), params)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}
