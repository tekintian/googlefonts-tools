//go:build !postgres

package db

import (
	"database/sql"
	"fmt"
)

func initDriverPostgres(dsn string) (*sql.DB, error) {
	return nil, fmt.Errorf("postgres driver not included: rebuild with -tags postgres")
}
