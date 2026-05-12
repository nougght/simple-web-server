package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// проверка валидации параметров конвертации валют
func TestConvertParams(t *testing.T) {
	test := []struct {
		name          string
		params        ConvertCurrencyParams
		errorExpected bool
	}{
		{"valid params", ConvertCurrencyParams{1000, "RUS", []string{"EUR", "USD"}}, false},
		{"negative amount", ConvertCurrencyParams{-1000, "RUS", []string{"EUR", "USD"}}, true},
		{"invalid base currency", ConvertCurrencyParams{1000, "RU", []string{"EUR", "USD"}}, true},
		{"invalid target currency", ConvertCurrencyParams{1000, "RUS", []string{"EURO", "USD"}}, true},
	}

	for _, test := range test {
		t.Run(test.name, func(t *testing.T) {
			err := test.params.Validate()
			if test.errorExpected {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
			}
		})
	}
}
