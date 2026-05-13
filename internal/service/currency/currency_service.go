package currency

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"simple-server/internal/model"
	"simple-server/internal/util"
)

type CurrencyService struct {
	currencyRatesBaseUrl string
	apiKey               string
}

func NewCurrencyService(apiKey string) *CurrencyService {
	return &CurrencyService{
		currencyRatesBaseUrl: "https://api.freecurrencyapi.com/v1/latest",
		apiKey:               apiKey,
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

	defer util.CloseBody(resp)

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

// конвертация валют
func (s *CurrencyService) ConvertCurrency(params *model.ConvertCurrencyParams) (map[string]float64, error) {
	// запрашиваем актуальный курс
	rates, err := s.requestCurrencyRates(params.BaseCurrency, params.TargetCurrencies)
	if err != nil {
		fmt.Print(err.Error() + "\n\n")
		return nil, fmt.Errorf("currency rates request error: %e", err)
	}
	// перемножаем курс на сумму для конвертации
	for currency := range rates {
		rates[currency] *= params.Amount
	}
	return rates, nil
}
