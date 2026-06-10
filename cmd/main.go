package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"simple-server/internal/config"
	"simple-server/internal/handler"
	"simple-server/internal/service"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Panic("Ошибка при загрузке конфигурации")
	}

	// перехват сигналов завершения работы
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	wg := &sync.WaitGroup{}
	services, err := service.GetServices(config, httpClient, wg)
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
		errChan <- err
	}()

	log.Println("Сервер запущен")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Panic(err)
	}

	// ожидание завершения(сохранения) задач
	wg.Wait()

	// ожидание завершения shutdown
	err = <-errChan
	if err != nil {
		log.Println(err)
	}

	log.Println("Сервер остановлен")

}
