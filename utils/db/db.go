package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
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
		if dsn == "" {
			dsn = "storage/db/googlefonts.db"
		}
		DB, err = sql.Open("sqlite", dsn)
		if err != nil {
			return fmt.Errorf("sqlite open error: %w", err)
		}
		DB.SetMaxOpenConns(1)
		DB.Exec("PRAGMA journal_mode=WAL")
		DB.Exec("PRAGMA synchronous=NORMAL")
		DB.Exec("PRAGMA cache_size=-64000")
		DB.Exec("PRAGMA busy_timeout=5000")

	case DriverMySQL:
		DB, err = sql.Open("mysql", dsn)
		if err != nil {
			return fmt.Errorf("mysql open error: %w", err)
		}
		DB.SetMaxOpenConns(25)
		DB.SetMaxIdleConns(10)

	case DriverPostgres:
		DB, err = sql.Open("pgx", dsn)
		if err != nil {
			return fmt.Errorf("postgres open error: %w", err)
		}
		DB.SetMaxOpenConns(25)
		DB.SetMaxIdleConns(10)

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
