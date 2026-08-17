//go:build mysql

package repository

import "database/sql"

func newMySQLRepo(database *sql.DB) TaskRepository {
	return NewMySQLRepo(database)
}
