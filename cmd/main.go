package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"simple-server/internal/handler"
	"simple-server/internal/model"
	"simple-server/internal/service"
)

func main() {
	config, err := model.LoadConfig()
	if err != nil {
		log.Println("Ошибка при загрузке конфигурации")
		panic(err)
	}

	services := service.GetServices(config)
	mux, _ := handler.GetHandlers(services)

	// перехват сигналов завершения работы
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	server := &http.Server{
		Addr:    ":8085",
		Handler: mux,
	}

	errChan := make(chan error)
	go func() {
		<-rootCtx.Done()
		log.Println("Остановка сервара, ожидание завершения текущих запросов")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		err := server.Shutdown(shutdownCtx)
		errChan <- err
	}()

	log.Println("Сервер запущен")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Panic(err)
	}

	// ожидание завершения shutdown
	err = <-errChan
	if err != nil {
		log.Println(err)
	}
	log.Println("Сервер остановлен")

}
