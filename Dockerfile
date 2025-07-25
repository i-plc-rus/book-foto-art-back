# Стадия сборки
FROM golang:1.23 AS builder

WORKDIR /app
COPY . .

# Сборка бинарника с отключённым CGO
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd

# Финальный образ — минимальный
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /app/app /app/app

USER root
RUN mkdir -p /app/uploads && chown -R nonroot:nonroot /app/uploads
# Запускаем как не-root пользователь
USER nonroot:nonroot
ENTRYPOINT ["/app/app"]
