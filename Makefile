
.PHONY: all lint build run test

# проверка, тестирование, сборка и запуск в контейнере
all: lint test docker-build docker-up

all-windows: lint test build-windows run-windows

# сборка под windows
build-windows:
	go build -o bin/server.exe .\cmd

run-windows: 
	.\bin\server.exe

# запуск линтеров
lint:
	docker run -t --rm -v .:/app -w /app golangci/golangci-lint:v2.12.2 \
	golangci-lint run --no-config -E govet,staticcheck

# запуск тестов
test:
	go test -v --cover ./...

# сборка образа
docker-build:
	docker build --no-cache -t simple-server -f docker/Dockerfile .

# запуск контейнера
docker-up:
	docker-compose -p simple-server -f docker/docker-compose.yml up -d 
