package currency

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"simple-server/internal/config"
	"simple-server/internal/model"
	"simple-server/internal/util"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CurrencyService struct {
	currencyRatesBaseUrl string
	apiKey               string
	taskService          model.TaskService
}

func NewCurrencyService(config *config.Config, taskService model.TaskService) *CurrencyService {
	return &CurrencyService{
		currencyRatesBaseUrl: config.FreecurrencyApiUrl,
		apiKey:               config.FreecurrencyApiKey,
		taskService:          taskService,
	}
}

// запрос курса валют через внешний api
func (s *CurrencyService) requestCurrencyRates(ctx context.Context, baseCurrency string, targetCurrencies []string) (map[string]float64, error) {
	params := url.Values{}
	params.Add("apikey", s.apiKey)
	params.Add("base_currency", baseCurrency)
	params.Add("currencies", util.SliceToCommaString(targetCurrencies))

	// добавляем параметры к url
	url := fmt.Sprintf("%s?%s", s.currencyRatesBaseUrl, params.Encode())

	// создаем запрос с контекстом
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request with context: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	defer util.CloseResponseBody(resp)
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	log.Println("Получены курсы валют: " + string(raw))

	var respBody model.CurrencyRatesResponse
	if err := util.DecodeJson(raw, &respBody); err != nil {
		return nil, err
	}
	return respBody.Data, nil
}

// конвертация с известными курсами
func (s *CurrencyService) convertCurrencyWithRates(amount float64, rates map[string]float64) map[string]float64 {
	amountDecimal := decimal.NewFromFloat(amount)
	result := make(map[string]float64)
	for currency, rate := range rates {
		result[currency] = amountDecimal.Mul(decimal.NewFromFloat(rate)).InexactFloat64()
	}
	return result
}

// конвертация валют
func (s *CurrencyService) ConvertCurrency(ctx context.Context, params *model.ConvertCurrencyParams) (map[string]float64, error) {
	// запрашиваем актуальный курс
	rates, err := s.requestCurrencyRates(ctx, params.BaseCurrency, params.TargetCurrencies)
	if err != nil {
		return nil, fmt.Errorf("currency rates request error: %w", err)
	}
	return s.convertCurrencyWithRates(params.Amount, rates), nil
}

func (s *CurrencyService) ConvertAndSaveAsync(ctx context.Context, params *model.ConvertCurrencyParams) (uuid.UUID, error) {
	id, err := s.taskService.ExecuteAndSaveAsync(ctx,
		func(taskCtx context.Context) (any, error) {
			return s.ConvertCurrency(taskCtx, params)
		})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to run async task: %w", err)
	}
	return id, nil
}
