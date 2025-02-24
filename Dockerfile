# syntax=docker/dockerfile:1.4
FROM golang:1.24.0 AS builder
#export DOCKER_BUILDKIT=1

WORKDIR /app

ENV GOPRIVATE=github.com/stipochka/*

COPY go.mod go.sum ./

RUN --mount=type=ssh \
    mkdir -p /root/.ssh \
    && ssh-keyscan github.com >> /root/.ssh/known_hosts \
    && git config --global url."git@github.com:".insteadOf "https://github.com" \
    && go mod download && go mod verify
COPY . .

WORKDIR /app/cmd/db_service

RUN go build -o db_service

FROM debian:bookworm-slim 

WORKDIR /app

COPY --from=builder /app/cmd/db_service .
COPY --from=builder /app/config .
COPY --from=builder /app/schema ./schema

# Add this to your Dockerfile to install wait-for-it
RUN apt-get update && apt-get install -y wait-for-it

# Add this to your entrypoint to wait for Kafka to be ready before starting mcu_service
#CMD ["./db_service --config /app/config.yaml"]
CMD ["wait-for-it", "kafka:9092", "--", "./db_service", "--config", "/app/config.yaml"]

