# ============================================================
# Этап 1 — Сборка Go приложения
# ============================================================
FROM golang:1.22-alpine AS builder

# Установить зависимости для сборки
RUN apk add --no-cache git

WORKDIR /app

# Сначала копируем только go.mod — чтобы кэшировать зависимости
COPY go.mod .
RUN go mod download

# Копируем остальной код
COPY . .

# Собрать бинарник
# CGO_ENABLED=0 — статическая сборка без внешних библиотек
RUN CGO_ENABLED=0 GOOS=linux go build -o openwatermap .

# ============================================================
# Этап 2 — Финальный образ (минимальный)
# ============================================================
FROM alpine:3.19

# Установить wget для healthcheck
RUN apk add --no-cache wget

WORKDIR /app

# Скопировать только бинарник из builder этапа
COPY --from=builder /app/openwatermap .

# Создать директорию для базы данных
RUN mkdir -p /data

# Открыть порт
EXPOSE 8080

# Запустить сервер
CMD ["./openwatermap"]
