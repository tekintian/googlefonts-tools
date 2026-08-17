package db

import (
	"database/sql"
	"fmt"
	"time"
)

var DB *sql.DB

type DriverType string

const (
	DriverSQLite   DriverType = "sqlite"
	DriverMySQL    DriverType = "mysql"
	DriverPostgres DriverType = "postgres"
)

var CurrentDriver DriverType

func Init(driverName, dsn string) error {
	var err error

	CurrentDriver = DriverType(driverName)

	switch CurrentDriver {
	case DriverSQLite:
		DB, err = initDriverSQLite(dsn)
	case DriverMySQL:
		DB, err = initDriverMySQL(dsn)
	case DriverPostgres:
		DB, err = initDriverPostgres(dsn)
	default:
		return fmt.Errorf("unsupported database driver: %s", driverName)
	}

	if err != nil {
		return err
	}

	DB.SetConnMaxLifetime(5 * time.Minute)

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("database ping error: %w", err)
	}

	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

func AdaptPlaceholders(query string) string {
	if CurrentDriver != DriverPostgres {
		return query
	}
	var buf []byte
	n := 1
	for i := 0; i < len(query); i++ {
		if query[i] == '?' {
			buf = append(buf, []byte(fmt.Sprintf("$%d", n))...)
			n++
		} else {
			buf = append(buf, query[i])
		}
	}
	return string(buf)
}
