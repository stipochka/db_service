package service

import (
	"context"

	"github.com/db_service/internal/models"
	"github.com/db_service/internal/storage"
)

//GetRecord(ctx context.Context, id int) (models.Record, error)
//GetRecords(ctx context.Context) ([]models.Record, error)

type RecordGetterService struct {
	db *storage.Storage
}

func NewRecordGetterService(db *storage.Storage) *RecordGetterService {
	return &RecordGetterService{
		db: db,
	}
}

func (r *RecordGetterService) GetRecord(ctx context.Context, id int) (models.Record, error) {
	return r.db.GetRecordByID(ctx, id)
}

func (r *RecordGetterService) GetRecords(ctx context.Context) ([]models.Record, error) {
	return r.db.GetAllRecords(ctx)
}
