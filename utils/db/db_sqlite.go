package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func initDriverSQLite(dsn string) (*sql.DB, error) {
	if dsn == "" {
		dsn = "storage/db/googlefonts.db"
	}
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlite open error: %w", err)
	}
	conn.SetMaxOpenConns(1)
	conn.Exec("PRAGMA journal_mode=WAL")
	conn.Exec("PRAGMA synchronous=NORMAL")
	conn.Exec("PRAGMA cache_size=-64000")
	conn.Exec("PRAGMA busy_timeout=5000")
	return conn, nil
}
