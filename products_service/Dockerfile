# Базовый образ для сборки
FROM golang:1.24 as builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы модуля Go
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod tidy

# Копируем все файлы проекта
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o products_service cmd/main.go


# Используем минималистичный образ для запуска
FROM gcr.io/distroless/base-debian11
WORKDIR /app
COPY --from=builder /app/products_service .

# Указываем команду для запуска
CMD ["./products_service"]