package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/db_service/internal/models"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	McuTable        = "records"
	ThresholdsTable = "thresholds"
)

type PostgresRecord struct {
	db *pgxpool.Pool
}

func NewPostgresRecord(db *pgxpool.Pool) *PostgresRecord {
	return &PostgresRecord{
		db: db,
	}
}

func (p *PostgresRecord) CreateRecord(ctx context.Context, deviceData models.Record) (int, error) {
	const op = "storage.CreateRecord"

	tx, err := p.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("%s: failed to start transaction: %w", op, err)
	}
	defer tx.Rollback(ctx) // Гарантированный откат в случае ошибки

	// Вставка основной записи
	query := fmt.Sprintf(`INSERT INTO %s (mcu_id, status, polling_period, temperature, lamp_status, voltage, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`, McuTable)

	var recordID int
	err = tx.QueryRow(ctx, query,
		deviceData.ID, deviceData.Status, deviceData.PollingPeriod,
		deviceData.Temperature, deviceData.LampStatus, deviceData.Voltage, time.Now(),
	).Scan(&recordID)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to insert record: %w", op, err)
	}

	// Вставка пороговых значений
	queryThresholds := fmt.Sprintf(`INSERT INTO %s (record_id, type, value, polling_period)
		VALUES ($1, $2, $3, $4)`, ThresholdsTable)

	insertThresholds := func(sensorType string, thresholds []models.SensorData) error {
		for _, threshold := range thresholds {
			_, err := tx.Exec(ctx, queryThresholds, recordID, sensorType, threshold.Value, threshold.PollingPeriod)
			if err != nil {
				return fmt.Errorf("%s: failed to insert threshold %s: %w", op, sensorType, err)
			}
		}
		return nil
	}

	// Вставляем данные о порогах
	if err := insertThresholds("temperature", deviceData.Thresholds.Temperature); err != nil {
		return 0, err
	}
	if err := insertThresholds("humidity", deviceData.Thresholds.Humidity); err != nil {
		return 0, err
	}
	if err := insertThresholds("voltage", deviceData.Thresholds.Voltage); err != nil {
		return 0, err
	}

	// Фиксируем транзакцию
	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to commit transaction: %w", op, err)
	}

	return recordID, nil
}

func (p *PostgresRecord) GetRecordByID(ctx context.Context, id int) (models.Record, error) {
	query := `
	SELECT 
		r.id,
		r.mcu_id,
		r.status,
		r.polling_period,
		r.temperature,
		r.lamp_status,
		r.voltage,
		r.created_at,
		COALESCE(json_agg(json_build_object(
			'type', t.type,
			'value', t.value, 
			'polling_period', t.polling_period
		)) FILTER (WHERE t.id IS NOT NULL), '[]'::json) AS thresholds
	FROM records r
	LEFT JOIN thresholds t ON r.id = t.record_id
	WHERE r.mcu_id = $1
	GROUP BY r.id
	ORDER BY r.created_at DESC
	LIMIT 1;`

	var mcuRes models.Record
	var timeStamp time.Time
	var thresholdsJSON []byte

	err := p.db.QueryRow(ctx, query, id).Scan(
		&mcuRes.ID, &mcuRes.Status, &mcuRes.PollingPeriod,
		&mcuRes.Temperature, &mcuRes.LampStatus, &mcuRes.Voltage,
		&timeStamp, &thresholdsJSON,
	)

	if err != nil {
		return models.Record{}, err // Возвращаем ошибку, а не nil
	}

	// Преобразуем timestamp в Go-формат
	mcuRes.Thresholds = models.Thresholds{}
	if len(thresholdsJSON) > 0 {
		err = json.Unmarshal(thresholdsJSON, &mcuRes.Thresholds)
		if err != nil {
			return models.Record{}, fmt.Errorf("failed to unmarshal thresholds: %w", err)
		}
	}

	return mcuRes, nil

}

func (p *PostgresRecord) GetAllRecords(ctx context.Context) ([]models.Record, error) {
	const op = "storage.GetAllRecords"
	query := `
		SELECT DISTINCT ON (r.mcu_id) 
			r.id,
			r.mcu_id,
			r.status,
			r.polling_period,
			r.temperature AS temperature,
			r.lamp_status,
			r.voltage AS voltage,
			r.created_at,
			COALESCE(json_agg(json_build_object(
				'type', t.type,
				'value', t.value, 
				'polling_period', t.polling_period
			)) FILTER (WHERE t.id IS NOT NULL), '[]'::json) AS thresholds
		FROM records r
		LEFT JOIN thresholds t ON r.id = t.record_id
		GROUP BY r.id
		ORDER BY r.mcu_id, r.created_at DESC;
	`

	rows, err := p.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	mcuRecords := make([]models.Record, 0)

	for rows.Next() {
		var record models.Record
		var thresholdsJSON []byte
		var timeStamp time.Time

		err := rows.Scan(
			&record.ID, &record.ID, &record.Status, &record.PollingPeriod,
			&record.Temperature, &record.LampStatus, &record.Voltage,
			&timeStamp, &thresholdsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		record.Thresholds = models.Thresholds{}

		// Временная структура для парсинга JSON
		var rawThresholds []struct {
			Type          string `json:"type"`
			Value         int    `json:"value"`
			PollingPeriod int    `json:"polling_period"`
		}

		if err := json.Unmarshal(thresholdsJSON, &rawThresholds); err != nil {
			return nil, fmt.Errorf("%s: ошибка парсинга thresholds: %w", op, err)
		}

		// Рапределяем пороговые значения по категориям (Temperature, Humidity, Voltage)
		for _, threshold := range rawThresholds {
			sensorData := models.SensorData{
				Value:         threshold.Value,
				PollingPeriod: threshold.PollingPeriod,
			}

			switch threshold.Type {
			case "temperature":
				record.Thresholds.Temperature = append(record.Thresholds.Temperature, sensorData)
			case "humidity":
				record.Thresholds.Humidity = append(record.Thresholds.Humidity, sensorData)
			case "voltage":
				record.Thresholds.Voltage = append(record.Thresholds.Voltage, sensorData)
			}
		}

		mcuRecords = append(mcuRecords, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return mcuRecords, nil
}
