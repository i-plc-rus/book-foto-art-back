# Stage 1: сборка
FROM golang:1.23 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/app .

# /uploads будет монтироваться volume, доступ к нему даст хост
VOLUME ["/uploads"]

USER root
ENTRYPOINT ["/app/app"]
