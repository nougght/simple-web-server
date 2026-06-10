package currency

import (
	"context"
	"net/http"
	"net/http/httptest"
	"simple-server/internal/config"
	"simple-server/internal/model"
	"simple-server/internal/util"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertCurrencyWithRates(t *testing.T) {
	service := NewCurrencyService(&config.Config{}, nil)

	tests := []struct {
		name     string
		amount   float64
		rates    map[string]float64
		expected map[string]float64
	}{
		{"RUB to other", 1000, map[string]float64{"USD": 0.012, "EUR": 0.011},
			map[string]float64{"USD": 12.0, "EUR": 11.0}},
		{"USD to other", 100, map[string]float64{"RUB": 83.3312, "EUR": 0.915},
			map[string]float64{"RUB": 8333.12, "EUR": 91.5}},
		{"large amount", 1000000, map[string]float64{"USD": 0.012, "EUR": 0.011},
			map[string]float64{"USD": 12000.0, "EUR": 11000.0}},
		{"small amount", 0.001, map[string]float64{"USD": 0.012, "EUR": 0.011},
			map[string]float64{"USD": 0.000012, "EUR": 0.000011}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result := service.convertCurrencyWithRates(test.amount, test.rates)
			require.Equal(t, test.expected, result)
		})
	}
}

func TestRequestCurrencyRates(t *testing.T) {
	rates := map[string]float64{"USD": 0.012, "EUR": 0.011}

	tests := []struct {
		name             string
		baseCurrency     string
		targetCurrencies []string
		context          context.Context
		serverHandler    http.HandlerFunc
		expectedResult   map[string]float64
		isErrorExpected  bool
	}{
		{"valid response", "RUB", []string{"USD", "EUR"}, context.Background(),
			func(w http.ResponseWriter, r *http.Request) {
				jsonResponse, err := util.EncodeJson(model.CurrencyRatesResponse{Data: rates})
				if err != nil {
					t.Fatalf("failed to encode response: %s", err)
				}
				if _, err := w.Write(jsonResponse); err != nil {
					t.Fatalf("failed to write response: %s", err)
				}
			}, rates, false},
		{"error response", "RUB", []string{"USD", "EUR"}, context.Background(),
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}, nil, true},
		{"invalid resposne body", "RUB", []string{"USD", "EUR"}, context.Background(),
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				if _, err := w.Write([]byte("54613")); err != nil {
					t.Fatalf("failed to write response: %s", err)
				}
			}, nil, true},
		{"context canceled", "RUB", []string{"USD", "EUR"}, func() context.Context {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			return ctx
		}(), func(w http.ResponseWriter, r *http.Request) {
			jsonResponse, err := util.EncodeJson(model.CurrencyRatesResponse{Data: rates})
			if err != nil {
				t.Fatalf("failed to encode response: %s", err)
			}
			if _, err := w.Write(jsonResponse); err != nil {
				t.Fatalf("failed to write response: %s", err)
			}
		}, nil, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(test.serverHandler)
			defer server.Close()
			service := NewCurrencyService(&config.Config{FreecurrencyApiUrl: server.URL, FreecurrencyApiKey: "test-api-key"}, nil)
			result, err := service.requestCurrencyRates(test.context, test.baseCurrency, test.targetCurrencies)
			if test.isErrorExpected {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, test.expectedResult, result)
			}
		})
	}
}
