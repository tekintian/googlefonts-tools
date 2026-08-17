package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/tekintian/googlefonts-tools/app/model"
	"github.com/tekintian/googlefonts-tools/utils/db"
)

type sqliteRepo struct {
	db *sql.DB
}

func NewSQLiteRepo(database *sql.DB) TaskRepository {
	return &sqliteRepo{db: database}
}

func (r *sqliteRepo) Init() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			sign          TEXT    PRIMARY KEY,
			url           TEXT    NOT NULL,
			font_name     TEXT    DEFAULT '',
			status        TEXT    DEFAULT 'pending',
			progress      INTEGER DEFAULT 0,
			total_files   INTEGER DEFAULT 0,
			done_files    INTEGER DEFAULT 0,
			zip_path      TEXT    DEFAULT '',
			zip_size      INTEGER DEFAULT 0,
			error_msg     TEXT    DEFAULT '',
			notify_config TEXT    DEFAULT '',
			created_at    DATETIME NOT NULL,
			updated_at    DATETIME NOT NULL,
			completed_at  DATETIME,
			download_count INTEGER DEFAULT 0
		)
	`)
	if err != nil {
		return fmt.Errorf("create table error: %w", err)
	}
	r.db.Exec(`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)`)
	r.db.Exec(`CREATE INDEX IF NOT EXISTS idx_tasks_font_name ON tasks(font_name)`)
	r.db.Exec(`CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at)`)
	return nil
}

func (r *sqliteRepo) GetBySign(sign string) (*model.Task, error) {
	q := buildSelectQuery() + ` WHERE sign = ?`
	row := r.db.QueryRow(q, sign)
	return scanOne(row)
}

func (r *sqliteRepo) GetByStatus(status string, limit int) ([]*model.Task, error) {
	q := buildSelectQuery() + ` WHERE status = ? ORDER BY created_at DESC LIMIT ?`
	rows, err := r.db.Query(q, status, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *sqliteRepo) List(offset, limit int) ([]*model.Task, error) {
	q := buildSelectQuery() + ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := r.db.Query(q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *sqliteRepo) Count() (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM tasks`).Scan(&count)
	return count, err
}

func (r *sqliteRepo) Create(task *model.Task) error {
	q := `INSERT INTO tasks (sign, url, font_name, status, progress,
		total_files, done_files, zip_path, zip_size, error_msg,
		notify_config, created_at, updated_at, completed_at, download_count)
		VALUES (` + buildPlaceholders(15) + `)`
	_, err := r.db.Exec(q,
		task.Sign, task.URL, task.FontName, task.Status, task.Progress,
		task.TotalFiles, task.DoneFiles, task.ZipPath, task.ZipSize, task.ErrorMsg,
		task.NotifyConfig, task.CreatedAt, task.UpdatedAt, task.CompletedAt, task.DownloadCount,
	)
	return err
}

func (r *sqliteRepo) Update(task *model.Task) error {
	q := `UPDATE tasks SET url=?, font_name=?, status=?, progress=?,
		total_files=?, done_files=?, zip_path=?, zip_size=?, error_msg=?,
		notify_config=?, updated_at=?, completed_at=?, download_count=?
		WHERE sign=?`
	_, err := r.db.Exec(q,
		task.URL, task.FontName, task.Status, task.Progress,
		task.TotalFiles, task.DoneFiles, task.ZipPath, task.ZipSize, task.ErrorMsg,
		task.NotifyConfig, task.UpdatedAt, task.CompletedAt, task.DownloadCount,
		task.Sign,
	)
	return err
}

func (r *sqliteRepo) UpdateProgress(sign string, status model.TaskStatus, progress, doneFiles, totalFiles int) error {
	q := `UPDATE tasks SET status=?, progress=?, done_files=?, total_files=?, updated_at=? WHERE sign=?`
	_, err := r.db.Exec(q, status, progress, doneFiles, totalFiles, time.Now(), sign)
	return err
}

func (r *sqliteRepo) UpdateSuccess(sign string, zipPath string, zipSize int64) error {
	q := `UPDATE tasks SET status=?, progress=?, zip_path=?, zip_size=?, updated_at=?, completed_at=? WHERE sign=?`
	_, err := r.db.Exec(q, model.StatusSuccess, 100, zipPath, zipSize, time.Now(), time.Now(), sign)
	return err
}

func (r *sqliteRepo) UpdateFailed(sign string, errMsg string) error {
	q := `UPDATE tasks SET status=?, error_msg=?, updated_at=?, completed_at=? WHERE sign=?`
	_, err := r.db.Exec(q, model.StatusFailed, errMsg, time.Now(), time.Now(), sign)
	return err
}

func (r *sqliteRepo) IncrementDownloadCount(sign string) error {
	q := `UPDATE tasks SET download_count = download_count + 1, updated_at = ? WHERE sign = ?`
	_, err := r.db.Exec(q, time.Now(), sign)
	return err
}

func (r *sqliteRepo) DeleteBefore(t time.Time) (int64, error) {
	q := `DELETE FROM tasks WHERE created_at < ? AND status = ?`
	result, err := r.db.Exec(q, t, model.StatusSuccess)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func buildSelectQuery() string {
	return `SELECT sign, url, font_name, status, progress, total_files, done_files,
		zip_path, zip_size, error_msg, notify_config,
		created_at, updated_at, completed_at, download_count
		FROM tasks`
}

func buildPlaceholders(n int) string {
	if db.CurrentDriver == db.DriverPostgres {
		parts := make([]string, n)
		for i := 0; i < n; i++ {
			parts[i] = fmt.Sprintf("$%d", i+1)
		}
		return strings.Join(parts, ", ")
	}
	return strings.Repeat("?, ", n-1) + "?"
}

func NewRepo(database *sql.DB) TaskRepository {
	switch db.CurrentDriver {
	case db.DriverMySQL:
		return NewMySQLRepo(database)
	case db.DriverPostgres:
		return NewPostgresRepo(database)
	default:
		return NewSQLiteRepo(database)
	}
}

func scanOne(row *sql.Row) (*model.Task, error) {
	var t model.Task
	var completedAt sql.NullTime
	err := row.Scan(
		&t.Sign, &t.URL, &t.FontName, &t.Status, &t.Progress,
		&t.TotalFiles, &t.DoneFiles, &t.ZipPath, &t.ZipSize,
		&t.ErrorMsg, &t.NotifyConfig,
		&t.CreatedAt, &t.UpdatedAt, &completedAt, &t.DownloadCount,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if completedAt.Valid {
		t.CompletedAt = &completedAt.Time
	}
	return &t, nil
}

func scanRows(rows *sql.Rows) ([]*model.Task, error) {
	var tasks []*model.Task
	for rows.Next() {
		var t model.Task
		var completedAt sql.NullTime
		err := rows.Scan(
			&t.Sign, &t.URL, &t.FontName, &t.Status, &t.Progress,
			&t.TotalFiles, &t.DoneFiles, &t.ZipPath, &t.ZipSize,
			&t.ErrorMsg, &t.NotifyConfig,
			&t.CreatedAt, &t.UpdatedAt, &completedAt, &t.DownloadCount,
		)
		if err != nil {
			return nil, err
		}
		if completedAt.Valid {
			t.CompletedAt = &completedAt.Time
		}
		tasks = append(tasks, &t)
	}
	return tasks, nil
}
