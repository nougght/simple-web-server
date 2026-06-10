package config

import (
	"fmt"
	"log"
	"os"
	"simple-server/internal/model"

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
	NoteStorageType    string
	FreecurrencyApiUrl string
	FreecurrencyApiKey string
}

// загрузка конфига
func LoadConfig() (*Config, error) {

	// попытка загрузки .env файла
	if err := godotenv.Load(); err != nil {
		log.Println(err.Error())
	}

	noteStorageType, exists := os.LookupEnv("NOTE_STORAGE_TYPE")
	if !exists {
		return nil, fmt.Errorf("не найден тип хранилища")
	}

	pgConfig := &PostgresConfig{}
	if noteStorageType != model.StorageTypePostgres && noteStorageType != model.StorageTypeMemory {
		return nil, fmt.Errorf("неизвестный тип хранилища: %s", noteStorageType)
	}

	// if noteStorageType == model.StorageTypePostgres {
	// }

	pgConfig.Host, exists = os.LookupEnv("POSTGRES_HOST")
	if !exists {
		return nil, fmt.Errorf("не найден хост postgres")
	}
	pgConfig.Port, exists = os.LookupEnv("POSTGRES_PORT")
	if !exists {
		return nil, fmt.Errorf("не найден порт postgres")
	}
	pgConfig.User, exists = os.LookupEnv("POSTGRES_USER")
	if !exists {
		return nil, fmt.Errorf("не найден пользователь postgres")
	}
	pgConfig.Password, exists = os.LookupEnv("POSTGRES_PASSWORD")
	if !exists {
		return nil, fmt.Errorf("не найден пароль postgres")
	}
	pgConfig.DBName, exists = os.LookupEnv("POSTGRES_DB")
	if !exists {
		return nil, fmt.Errorf("не найдено имя базы данных postgres")
	}
	pgConfig.SSLMode, exists = os.LookupEnv("POSTGRES_SSLMODE")
	if !exists {
		return nil, fmt.Errorf("не найден SSLMode postgres")
	}

	apiUrl, exists := os.LookupEnv("FREECURRENCY_API_URL")
	if !exists {
		return nil, fmt.Errorf("не найден api url")
	}
	apiKey, exists := os.LookupEnv("FREECURRENCY_API_KEY")
	if !exists {
		return nil, fmt.Errorf("не найден api ключ")
	}

	return &Config{Postgres: pgConfig, NoteStorageType: noteStorageType, FreecurrencyApiUrl: apiUrl, FreecurrencyApiKey: apiKey}, nil
}
