# Stage 1: сборка
FROM golang:1.23 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd

# Stage 2: финальный distroless
FROM gcr.io/distroless/static:nonroot

WORKDIR /app
COPY --from=builder /app/app .

# /uploads будет монтироваться volume, доступ к нему даст хост
VOLUME ["/uploads"]

USER nonroot
ENTRYPOINT ["/app/app"]
