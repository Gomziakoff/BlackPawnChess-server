# Используем i386 образ для сборки
#FROM --platform=linux/386 i386/golang:1.25.5-alpine AS builder
#WORKDIR /app
#COPY go.mod go.sum ./
#RUN go mod download
#COPY . .
#RUN GOOS=linux GOARCH=386 go build -o main ./cmd/server/main.go

# Финальный образ
FROM --platform=linux/386 i386/alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]

# Этап 1: Сборка бинарного файла
FROM golang:1.25.5-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код и собираем
COPY . .
RUN go build -o main ./cmd/server/main.go

# Этап 2: Минимальный образ для запуска
FROM alpine:latest
WORKDIR /root/

# Копируем только исполняемый файл из первого этапа
COPY --from=builder /app/main .
# Если у вас есть папка с конфигами или .env, скопируйте их (необязательно, если через compose)
# COPY --from=builder /app/.env . 

EXPOSE 8080

# Запуск приложения
CMD ["./main"]