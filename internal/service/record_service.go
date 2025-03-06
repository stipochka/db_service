package service

import (
	"context"

	"github.com/db_service/internal/models"
)

type RecordCreator interface {
	CreateRecord(ctx context.Context, key []byte, value []byte) (int, error)
}

//go:generate mockery --name=RecordGetter --output=../servicemoks --outpkg=servicemocks
type RecordGetter interface {
	GetRecordByID(ctx context.Context, id int) (models.Record, error)
	GetAllRecords(ctx context.Context) ([]models.Record, error)
}
