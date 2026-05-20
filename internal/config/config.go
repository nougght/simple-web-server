package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func (c *PostgresConfig) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

type Config struct {
	Postgres           *PostgresConfig
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
