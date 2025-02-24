package storage

import (
	"context"

	"github.com/db_service/internal/models"
	"github.com/jackc/pgx/v4/pgxpool"
)

type RecordAdder interface {
}

type Record interface {
	CreateRecord(ctx context.Context, deviceData models.Record) (int, error)
	GetRecordByID(ctx context.Context, id int) (models.Record, error)
	GetAllRecords(ctx context.Context) ([]models.Record, error)
}

type Storage struct {
	Record
}

func NewStorage(db *pgxpool.Pool) *Storage {
	return &Storage{
		Record: NewPostgresRecord(db),
	}
}
