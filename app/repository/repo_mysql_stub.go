//go:build !mysql

package repository

import (
	"database/sql"
	"fmt"
)

func newMySQLRepo(database *sql.DB) TaskRepository {
	fmt.Println("[Warning] MySQL driver not included, rebuild with -tags mysql")
	return NewSQLiteRepo(database)
}
