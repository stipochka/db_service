package app

import (
	"log/slog"

	"github.com/db_service/internal/app/grpcapp"
	"github.com/db_service/internal/service"
	"github.com/db_service/internal/storage"
	"github.com/jackc/pgx/v4/pgxpool"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(db *pgxpool.Pool, log *slog.Logger, port int) *App {
	pgStorage := storage.NewStorage(db)

	getterService := service.NewRecordGetterService(pgStorage)

	grpcApp := grpcapp.New(log, getterService, port)

	return &App{
		GRPCServer: grpcApp,
	}

}
