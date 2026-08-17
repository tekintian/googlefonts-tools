//go:build !mysql

package db

import (
	"database/sql"
	"fmt"
)

func initDriverMySQL(dsn string) (*sql.DB, error) {
	return nil, fmt.Errorf("mysql driver not included: rebuild with -tags mysql")
}
