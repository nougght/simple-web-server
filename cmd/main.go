package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"simple-server/internal/config"
	"simple-server/internal/handler"
	"simple-server/internal/service"
)

func main() {
	config := config.LoadConfig()

	// перехват сигналов завершения работы
	rootCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	services, err := service.GetServices(config, httpClient, rootCtx)
	if err != nil {
		log.Panicf("Ошибка при инициализации сервисов: %s", err.Error())
	}
	mux, _ := handler.GetHandlers(services, rootCtx)

	server := &http.Server{
		Addr:    ":8085",
		Handler: mux,
	}

	errChan := make(chan error)
	go func() {
		<-rootCtx.Done()
		log.Println("Остановка сервера, ожидание завершения текущих запросов")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		err := server.Shutdown(shutdownCtx)
		services.TaskService().Stop()
		errChan <- err
	}()

	log.Println("Сервер запущен")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Println(err)
		cancel()
	}

	// ожидание завершения shutdown
	err = <-errChan
	if err != nil {
		log.Println(err)
	}

	log.Println("Сервер остановлен")

}
