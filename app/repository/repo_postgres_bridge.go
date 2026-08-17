//go:build postgres

package repository

import "database/sql"

func newPostgresRepo(database *sql.DB) TaskRepository {
	return NewPostgresRepo(database)
}
