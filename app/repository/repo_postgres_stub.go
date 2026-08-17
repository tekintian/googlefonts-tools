//go:build !postgres

package repository

import (
	"database/sql"
	"fmt"
)

func newPostgresRepo(database *sql.DB) TaskRepository {
	fmt.Println("[Warning] PostgreSQL driver not included, rebuild with -tags postgres")
	return NewSQLiteRepo(database)
}
