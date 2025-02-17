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

	rows, err := conn.Query(context.Background(), "CREATE TABLE mcu_data (id INT NOT NULL, data VARCHAR(1000) DEFAULT NULL);")
	rows.Close()
	return conn, err
}
