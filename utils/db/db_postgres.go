//go:build postgres

package db

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func initDriverPostgres(dsn string) (*sql.DB, error) {
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres open error: %w", err)
	}
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(10)
	return conn, nil
}
