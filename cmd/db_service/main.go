package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/db_service/internal/config"
	"github.com/db_service/internal/consumer"
	"github.com/db_service/internal/service"
	"github.com/db_service/internal/storage"
)

func main() {
	//TODO: init config

	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("started db_service")

	conn, err := storage.NewPostgresConn(cfg.DbUrl)
	if err != nil {
		log.Error("failed to connect to db", slog.String("error", err.Error()))
		return
	}

	st := storage.NewStorage(conn)

	kafkaService := service.NewRecordSaverService(st)
	messageConsumer, err := consumer.NewKafkaConsumer(
		cfg.Producer.Address,
		cfg.Producer.GroupID,
		cfg.Producer.Topic,
		kafkaService,
		*log,
	)

	if err != nil {
		log.Error("failed to create consumer", slog.String("error", err.Error()))
		return
	}

	log.Info("Created kafka consumer")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		messageConsumer.Consumer.Consume(ctx)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	cancel()

	//TODO: create gRPC-server
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case "local":
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case "dev":
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	case "prod":
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
