FROM golang:1.23.6 AS builder


WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

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

