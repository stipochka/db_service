package storage

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

func NewPostgresConn(dbUrl string) (*pgxpool.Pool, error) {
	conn, err := pgxpool.Connect(context.Background(), dbUrl)

	if err != nil {
		return nil, err
	}

	err = conn.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	_, err = conn.Exec(context.Background(), `
		CREATE TABLE records (
			id SERIAL PRIMARY KEY,
			mcu_id INT NOT NULL,
			status VARCHAR(10) NOT NULL CHECK (status IN ('ON', 'OFF')),
			polling_period INT NOT NULL,
			temperature INT NOT NULL, -- Храним в целых числах, реальное значение = temperature / 100
			lamp_status VARCHAR(10) NOT NULL CHECK (lamp_status IN ('ON', 'OFF')),
			voltage INT NOT NULL, -- Храним в целых числах, реальное значение = voltage / 100
			created_at TIMESTAMP DEFAULT NOW(), -- Автоматическое время создания
			UNIQUE (mcu_id, created_at) -- Гарантирует уникальные записи по времени для каждого MCU
		);
		CREATE INDEX idx_records_mcu_created ON records (mcu_id, created_at DESC);

		CREATE TABLE thresholds (
			id SERIAL PRIMARY KEY,
			record_id INT REFERENCES records(id) ON DELETE CASCADE,
			type VARCHAR(20) NOT NULL CHECK (type IN ('temperature', 'humidity', 'voltage')), -- Тип порога
			value INT NOT NULL, -- Храним в целых числах, реальное значение = value / 100
			polling_period INT NOT NULL
		);
	`)
	return conn, err
}
