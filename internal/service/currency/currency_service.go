package currency

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"simple-server/internal/config"
	"simple-server/internal/model"
	"simple-server/internal/util"
)

type CurrencyService struct {
	currencyRatesBaseUrl string
	apiKey               string
}

func NewCurrencyService(config *config.Config) *CurrencyService {
	return &CurrencyService{
		currencyRatesBaseUrl: config.FreecurrencyApiUrl,
		apiKey:               config.FreecurrencyApiKey,
	}
}

// запрос курса валют через внешний api
func (s *CurrencyService) requestCurrencyRates(baseCurrency string, targetCurrencies []string) (map[string]float64, error) {
	params := url.Values{}
	params.Add("apikey", s.apiKey)
	params.Add("base_currency", baseCurrency)
	params.Add("currencies", util.SliceToCommaString(targetCurrencies))

	// добавляем параметры к url и отправляем запрос
	resp, err := http.Get(fmt.Sprintf("%s?%s", s.currencyRatesBaseUrl, params.Encode()))
	if err != nil {
		return nil, err
	}

	defer util.CloseResponseBody(resp)

	// читаем тело запроса
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Println("Получены курсы валют: " + string(raw))

	var respBody model.CurrencyRatesResponse
	// декодируем json
	if err := util.DecodeJson(raw, &respBody); err != nil {
		return nil, err
	}

	return respBody.Data, nil
}

// конвертация валют
func (s *CurrencyService) ConvertCurrency(params *model.ConvertCurrencyParams) (map[string]float64, error) {
	// запрашиваем актуальный курс
	rates, err := s.requestCurrencyRates(params.BaseCurrency, params.TargetCurrencies)
	if err != nil {
		return nil, fmt.Errorf("currency rates request error: %w", err)
	}
	// перемножаем курс на сумму для конвертации
	for currency := range rates {
		rates[currency] *= params.Amount
	}
	return rates, nil
}
