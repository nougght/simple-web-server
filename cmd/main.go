package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"simple-server/internal/model"
	"simple-server/internal/service"
	"simple-server/internal/handler"
)

func main() {
	config, err := model.LoadConfig()
	if err != nil {
		log.Println("ошибка при загрузке конфигурации")
		panic(err)
	}

	services := service.GetServices(config)
	mux, _ := handler.GetHandlers(services)

	// перехват сигналов завершения работы
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverCtx, stopServer := context.WithCancel(context.Background())
	server := &http.Server{
		Addr: ":8085",
		BaseContext: func(_ net.Listener) context.Context {
			return serverCtx
		},
		Handler: mux,
	}

	go func() {
		log.Println("Сервер запущен")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	// ожидание сигнала завершения работы
	<-rootCtx.Done()
	// выключаем перехват сигналов
	stop()

	log.Println("Остановка сервара, ожидание завершения текущих запросов")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	// отключаем неактивные соединения и ждем завершения запросов
	err = server.Shutdown(shutdownCtx)

	// если запросы не завершились, отменяем контекст сервера
	stopServer()
	if err != nil {
		log.Println("Ошибка при остановке сервера, ожидаем еще 3 секунды")
		time.Sleep(time.Second * 3)
	}
	log.Println("Сервер остановлен")
}
