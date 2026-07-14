FROM golang:1.24.6-alpine AS builder

WORKDIR /src

COPY go.mod ./
COPY go.sum ./
COPY cmd ./cmd
COPY internal ./internal

RUN go build -o /app/api ./cmd/api && \
    go build -o /app/worker ./cmd/worker

FROM alpine:3.20

RUN apk add --no-cache ca-certificates wget

WORKDIR /app

COPY --from=builder /app/api /app/api
COPY --from=builder /app/worker /app/worker

EXPOSE 8080
