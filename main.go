package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"simple_server/handlers"
)

func main() {
	// загрузка переменных окружения 
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("Не найден .env файл")
		return
	}
	apiKey, exists := os.LookupEnv("FREECURRENCY_API_KEY")
	if exists == false {
		fmt.Println("Не найден api ключ")
		return
	}
	
	// обработчик для курсов валют
	currencyHandler := handlers.NewCurencyHandler(apiKey)

	mux := http.NewServeMux()
	// регистрация эндпоинтов
	mux.HandleFunc("GET /currency", currencyHandler.ConvertCurrency)

	http.ListenAndServe(":9000", mux)
}
