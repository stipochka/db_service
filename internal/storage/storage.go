package storage

import (
	"github.com/db_service/internal/models"
	"github.com/jackc/pgx/v4/pgxpool"
)

type RecordAdder interface {
}

type Record interface {
	CreateRecord(deviceData models.Record) (int, error)
	GetRecordByID(id int) (models.Record, error)
	GetAllRecords() ([]models.Record, error)
}

type Storage struct {
	Record
}

func NewStorage(db *pgxpool.Pool) *Storage {
	return &Storage{
		Record: NewPostgresRecord(db),
	}
}
