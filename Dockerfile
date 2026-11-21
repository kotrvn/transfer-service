# Этап 1: Сборка приложения
FROM golang:1.27-alpine AS builder

WORKDIR /app

# Копируем файлы зависимостей и скачиваем их
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь остальной код и собираем бинарник
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./cmd/api/main.go

# Этап 2: Финальный легкий образ
FROM alpine:latest

WORKDIR /app

# Устанавливаем goose для миграций
RUN apk add --no-cache curl \
    && curl -fsSL https://raw.githubusercontent.com/pressly/goose/master/install.sh | sh

# Копируем скомпилированный бинарник
COPY --from=builder /app/api .
# Копируем файлы миграций
COPY migrations ./migrations

EXPOSE 8080

# Теперь при запуске мы сначала применяем миграции, а потом запускаем сервер!
CMD goose -dir migrations postgres "$DATABASE_URL" up && ./api