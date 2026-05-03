package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// используется бесплатный api для получения курса валют - https://freecurrencyapi.com/docs/

// параметры для запроса
type ApiRequestParams struct {
	BaseCurrency     string   // исходная валюта
	TargetCurrencies []string // запрашиваемые валюты
}

// преобразование списка валют в строку с разделителем
func (p *ApiRequestParams) targetCurrenciesToString() string {
	return strings.Join(p.TargetCurrencies, ",")
}

// ответ на запрос курсов валют
type ApiResponse struct {
	Data map[string]float64 `json:"data"`
}

type CurrencyHandler struct {
	apiBaseUrl string // url
	apiKey     string

	defaultApiParams *ApiRequestParams
}

func NewCurencyHandler(key string) *CurrencyHandler {
	return &CurrencyHandler{
		apiKey:     key,
		apiBaseUrl: "https://api.freecurrencyapi.com/v1/latest",

		defaultApiParams: &ApiRequestParams{
			BaseCurrency:     "RUB",
			TargetCurrencies: []string{},
		},
	}
}

// внутренний метод - запрос курса валют через внешний api
func (h *CurrencyHandler) requestCurrencyRates(requestParams ApiRequestParams) (map[string]float64, error) {
	params := url.Values{}
	params.Add("apikey", h.apiKey)
	params.Add("base_currency", requestParams.BaseCurrency)
	params.Add("currencies", requestParams.targetCurrenciesToString())

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

	var respBody ApiResponse
	// декодируем json
	if err := json.Unmarshal(raw, &respBody); err != nil {
		return nil, err
	}

	return respBody.Data, nil
}

// обработка запроса конвертации валют
func (h *CurrencyHandler) ConvertCurrency(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)
	amount := 1.0
	str := r.URL.Query().Get("amount")
	// если параметр не пуст - обрабатываем его, иначе используем 1.0 по умолчанию
	if str != "" {
		// если параметр не является числом - возвращаем ошибку
		if val, err := strconv.ParseFloat(str, 64); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("incorrect 'amount' parameter"))
			fmt.Print(err.Error() + "\n\n")
			return
		} else {
			amount = val
		}
	}
	requestParams := *h.defaultApiParams
	// меняем параметры по умолчанию на полученные в запросе
	if r.URL.Query().Has("base") {
		requestParams.BaseCurrency = r.URL.Query().Get("base")
	}
	if r.URL.Query().Has("currencies") {
		requestParams.TargetCurrencies = r.URL.Query()["currencies"]
	}

	// получаем курс нужных валют для конвертации
	rates, err := h.requestCurrencyRates(requestParams)
	if err != nil {
		fmt.Print(err.Error() + "\n\n")
		// возвращаем ответ с ошибкой
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("currency rates request error"))
		return
	}

	// инициализируем словарь для результатов
	response := make(map[string]float64, len(rates))
	// конвертация по курсу
	for currency, rate := range rates {
		response[currency] = amount * rate
	}

	// кодируем результат в json
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		fmt.Print(err.Error() + "\n\n")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("json encoding error"))
		return
	}
	// отправляем ответ с успешным статусом
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)

	fmt.Print("Ответ успешно отправлен\n\n")
}
