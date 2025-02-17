package service

import (
	"context"

	"github.com/db_service/internal/models"
)

type RecordCreator interface {
	CreateRecord(ctx context.Context, key []byte, value []byte) (int, error)
}

type RecordGetter interface {
	GetRecord(ctx context.Context, id int) (models.Record, error)
	GetRecords(ctx context.Context) ([]models.Record, error)
}
