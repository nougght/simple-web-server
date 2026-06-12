package config

import (
	"fmt"
	"log"
	"os"
	"simple-server/internal/model"
	"strconv"

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
	TaskWorkersCount   int
	TaskBufferSize     int
}

// получение переменной из окружения
func getValue(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		panic("env variable not found - " + key)
	}
	return value
}

// получение необязательной переменной типа int из окружения со значением по умолчанию
func getOptionalInt(key string, def int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return def
	}
	intVal, err := strconv.Atoi(value)
	if err != nil {
		panic(fmt.Sprintf("invalid int env variable %s: %s", key, err))
	}
	return intVal
}

// загрузка конфига
func LoadConfig() *Config {

	// попытка загрузки .env файла
	if err := godotenv.Load(); err != nil {
		log.Println(err.Error())
	}

	noteStorageType := getValue("NOTE_STORAGE_TYPE")

	pgConfig := &PostgresConfig{}
	if noteStorageType != model.StorageTypePostgres && noteStorageType != model.StorageTypeMemory {
		panic(fmt.Sprintf("unknown NOTE_STORAGE_TYPE value: %s", noteStorageType))
	}

	// if noteStorageType == model.StorageTypePostgres {
	// }

	pgConfig.Host = getValue("POSTGRES_HOST")
	pgConfig.Port = getValue("POSTGRES_PORT")
	pgConfig.User = getValue("POSTGRES_USER")
	pgConfig.Password = getValue("POSTGRES_PASSWORD")
	pgConfig.DBName = getValue("POSTGRES_DB")
	pgConfig.SSLMode = getValue("POSTGRES_SSLMODE")

	apiUrl := getValue("FREECURRENCY_API_URL")
	apiKey := getValue("FREECURRENCY_API_KEY")

	workersCount := getOptionalInt("TASK_WORKERS_COUNT", model.DefaultTaskWorkersCount)
	if workersCount < 1 {
		panic("TASK_WORKERS_COUNT can't be negative")
	}

	bufferSize := getOptionalInt("TASK_BUFFER_SIZE", model.DefaultTaskBufferSize)
	if bufferSize < 1 {
		panic("TASK_BUFFER_SIZE can't be negative")
	}
	return &Config{Postgres: pgConfig, NoteStorageType: noteStorageType, FreecurrencyApiUrl: apiUrl, FreecurrencyApiKey: apiKey,
		TaskWorkersCount: workersCount,
		TaskBufferSize:   bufferSize}
}
