package service

import (
	"context"
	"encoding/json"

	"github.com/db_service/internal/models"
	"github.com/db_service/internal/storage"
)

type RecordSaverService struct {
	db *storage.Storage
}

func NewRecordSaverService(db *storage.Storage) *RecordSaverService {
	return &RecordSaverService{
		db: db,
	}
}

func (r *RecordSaverService) CreateRecord(ctx context.Context, key []byte, value []byte) (int, error) {
	var record models.Record

	err := json.Unmarshal(value, &record)
	if err != nil {
		return 0, err
	}

	return r.db.Record.CreateRecord(ctx, record)
}
