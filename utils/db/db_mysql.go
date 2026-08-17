//go:build mysql

package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func initDriverMySQL(dsn string) (*sql.DB, error) {
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("mysql open error: %w", err)
	}
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(10)
	return conn, nil
}
