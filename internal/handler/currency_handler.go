package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"simple-server/internal/model"
	"simple-server/internal/util"
	"strconv"
)

// используется бесплатный api для получения курса валют - https://freecurrencyapi.com/docs/

// параметры для запроса

type CurrencyHandler struct {
	apiBaseUrl string
	apiKey     string
}

func NewCurencyHandler(key string) *CurrencyHandler {
	return &CurrencyHandler{
		apiKey:     key,
		apiBaseUrl: "https://api.freecurrencyapi.com/v1/latest",
	}
}

// внутренний метод - запрос курса валют через внешний api
func (h *CurrencyHandler) requestCurrencyRates(baseCurrency string, targetCurrencies []string) (map[string]float64, error) {
	params := url.Values{}
	params.Add("apikey", h.apiKey)
	params.Add("base_currency", baseCurrency)
	params.Add("currencies", util.SliceToCommaString(targetCurrencies))

	// добавляем параметры к url и отправляем запрос
	resp, err := http.Get(fmt.Sprintf("%s?%s", h.apiBaseUrl, params.Encode()))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// читаем тело запроса
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println("Получены курсы валют: " + string(raw))

	var respBody model.CurrencyRatesResponse
	// декодируем json
	if err := json.Unmarshal(raw, &respBody); err != nil {
		return nil, err
	}

	return respBody.Data, nil
}

func (h *CurrencyHandler) parseConvertParameters(r *http.Request) (*model.ConvertCurrencyRequestParams, error) {
	// параметры по умолчанию
	params := &model.ConvertCurrencyRequestParams{
		Amount:           1.0,
		BaseCurrency:     "RUB",
		TargetCurrencies: []string{},
	}
	str := r.URL.Query().Get("amount")
	// если параметр не пуст - обрабатываем его
	if str != "" {
		// если параметр не является числом - возвращаем ошибку
		if val, err := strconv.ParseFloat(str, 64); err != nil {
			fmt.Print(err.Error() + "\n\n")
			return nil, fmt.Errorf("'amount' parsing failed: %e", err)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// проверяем параметры
	if err := params.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// получаем курс нужных валют для конвертации
	rates, err := h.requestCurrencyRates(params.BaseCurrency, params.TargetCurrencies)
	if err != nil {
		fmt.Print(err.Error() + "\n\n")
		// возвращаем ответ с ошибкой
		http.Error(w, "currency rates request error", http.StatusInternalServerError)
		return
	}

	// инициализируем словарь для результатов
	response := model.ConvertCurrencyResponse{}
	// конвертация по курсу
	for currency, rate := range rates {
		response[currency] = params.Amount * rate
	}

	// кодируем результат в json
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		fmt.Print(err.Error() + "\n\n")
		http.Error(w, "json encoding error", http.StatusInternalServerError)
		return
	}
	// отправляем ответ с успешным статусом
	w.Write(jsonResponse)

	fmt.Print("Ответ успешно отправлен\n\n")
}
