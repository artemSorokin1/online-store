# Базовый образ для сборки
FROM golang:1.21-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы модуля Go
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем все файлы проекта
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd

FROM postgres:15-alpine

COPY --from=builder /app/main /app/main
COPY --from=builder /app/scripts/init-db.sh /docker-entrypoint-initdb.d/

RUN chmod +x /docker-entrypoint-initdb.d/init-db.sh

EXPOSE 8084

CMD ["./app/main"]