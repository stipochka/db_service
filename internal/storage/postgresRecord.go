package storage

import (
	"context"
	"fmt"

	"github.com/db_service/internal/models"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	McuTable = "mcu_data"
)

type PostgresRecord struct {
	db *pgxpool.Pool
}

func NewPostgresRecord(db *pgxpool.Pool) *PostgresRecord {
	return &PostgresRecord{
		db: db,
	}
}

func (p *PostgresRecord) CreateRecord(deviceData models.Record) (int, error) {
	query := fmt.Sprintf("INSERT INTO %s (id, data) VALUES ($1, $2) RETURNING id", McuTable)

	row, err := p.db.Query(context.Background(), query, deviceData.ID, deviceData.Data)
	if err != nil {
		return 0, err
	}
	defer row.Close()

	return deviceData.ID, nil

}

func (p *PostgresRecord) GetRecordByID(id int) (models.Record, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id=$1", McuTable)

	var mcuRes models.Record

	err := p.db.QueryRow(context.Background(), query, id).Scan(&mcuRes.ID, &mcuRes.Data)
	if err != nil {
		return models.Record{}, nil
	}

	return mcuRes, nil
}

func (p *PostgresRecord) GetAllRecords() ([]models.Record, error) {

	mcuRows := make([]models.Record, 0)

	query := fmt.Sprintf("SELECT * FROM %s;", McuTable)

	rows, err := p.db.Query(context.Background(), query)
	if err != nil {
		return []models.Record{}, err
	}

	for rows.Next() {
		var row models.Record
		err = rows.Scan(&row.ID, &row.Data)
		if err != nil {
			return mcuRows, err
		}

		mcuRows = append(mcuRows, row)
	}

	return mcuRows, nil
}
