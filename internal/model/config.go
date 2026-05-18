package model

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	FreecurrencyApiUrl string
	FreecurrencyApiKey string
}

// загрузка конфига
func LoadConfig() (*Config, error) {

	// попытка загрузки .env файла
	if err := godotenv.Load(); err != nil {
		log.Println(err.Error())
	}

	apiUrl, exists := os.LookupEnv("FREECURRENCY_API_URL")
	if !exists {
		return nil, fmt.Errorf("не найден api url")
	}
	apiKey, exists := os.LookupEnv("FREECURRENCY_API_KEY")
	if !exists {
		return nil, fmt.Errorf("не найден api ключ")
	}

	return &Config{FreecurrencyApiUrl: apiUrl, FreecurrencyApiKey: apiKey}, nil
}
