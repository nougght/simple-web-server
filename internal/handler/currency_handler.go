package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"simple-server/internal/model"
	"strconv"
)

// используется бесплатный api для получения курса валют - https://freecurrencyapi.com/docs/

type CurrencyHandler struct {
	// сервис с бизнес-логикой
	service model.CurrencyService
}

func NewCurrencyHandler(service model.CurrencyService) *CurrencyHandler {
	return &CurrencyHandler{
		service: service,
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
		// если параметр не является числом - возвращаем ошибку
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

	// извлекаем параметры из запроса
	params, err := h.parseConvertParameters(r)
	if err != nil {
		handleError(w, err)
		return
	}
	// проверяем параметры
	if err := params.Validate(); err != nil {
		handleError(w, err)
		return
	}

	var result model.ConvertCurrencyResponse
	result, err = h.service.ConvertCurrency(params)
	if err != nil {
		handleError(w, err)
		return
	}

	// кодируем результат в json
	jsonResponse, err := json.Marshal(result)
	if err != nil {
		handleError(w, fmt.Errorf("json encoding error: %w", err))
		return
	}
	// отправляем ответ с успешным статусом
	_, err = w.Write(jsonResponse)
	if err != nil {
		log.Print(err.Error() + "\n\n")
	} else {
		log.Print("Ответ успешно отправлен\n\n")
	}
}
