FROM golang:1.26-alpine AS build

WORKDIR /build
# копируем содержимое директории
COPY . .
# запускаем сборку 
RUN go build -o app ./cmd/main.go


# легковесный образ для запуска (20-30мб)
FROM alpine
WORKDIR /app
# копируем исполняемый файл
COPY --from=build /build/app ./app

CMD ["./app"]