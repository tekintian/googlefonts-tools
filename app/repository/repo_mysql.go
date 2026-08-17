package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/tekintian/googlefonts-tools/app/model"
)

type mysqlRepo struct {
	db *sql.DB
}

func NewMySQLRepo(database *sql.DB) TaskRepository {
	return &mysqlRepo{db: database}
}

func (r *mysqlRepo) Init() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			sign          VARCHAR(32) PRIMARY KEY,
			url           TEXT NOT NULL,
			font_name     VARCHAR(128) DEFAULT '',
			status        VARCHAR(16) DEFAULT 'pending',
			progress      INT DEFAULT 0,
			total_files   INT DEFAULT 0,
			done_files    INT DEFAULT 0,
			zip_path      VARCHAR(512) DEFAULT '',
			zip_size      BIGINT DEFAULT 0,
			error_msg     TEXT DEFAULT '',
			notify_config TEXT DEFAULT '',
			created_at    DATETIME NOT NULL,
			updated_at    DATETIME NOT NULL,
			completed_at  DATETIME NULL,
			download_count INT DEFAULT 0,
			INDEX idx_tasks_status (status),
			INDEX idx_tasks_font_name (font_name),
			INDEX idx_tasks_created_at (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`)
	if err != nil {
		return fmt.Errorf("mysql create table error: %w", err)
	}
	return nil
}

func (r *mysqlRepo) GetBySign(sign string) (*model.Task, error) {
	q := buildSelectQuery() + ` WHERE sign = ?`
	return scanOne(r.db.QueryRow(q, sign))
}

func (r *mysqlRepo) GetByStatus(status string, limit int) ([]*model.Task, error) {
	q := buildSelectQuery() + ` WHERE status = ? ORDER BY created_at DESC LIMIT ?`
	rows, err := r.db.Query(q, status, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *mysqlRepo) List(offset, limit int) ([]*model.Task, error) {
	q := buildSelectQuery() + ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := r.db.Query(q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *mysqlRepo) Count() (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM tasks`).Scan(&count)
	return count, err
}

func (r *mysqlRepo) Create(task *model.Task) error {
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

func (r *mysqlRepo) Update(task *model.Task) error {
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

func (r *mysqlRepo) UpdateProgress(sign string, status model.TaskStatus, progress, doneFiles, totalFiles int) error {
	q := `UPDATE tasks SET status=?, progress=?, done_files=?, total_files=?, updated_at=? WHERE sign=?`
	_, err := r.db.Exec(q, status, progress, doneFiles, totalFiles, time.Now(), sign)
	return err
}

func (r *mysqlRepo) UpdateSuccess(sign string, zipPath string, zipSize int64) error {
	q := `UPDATE tasks SET status=?, progress=?, zip_path=?, zip_size=?, updated_at=?, completed_at=? WHERE sign=?`
	_, err := r.db.Exec(q, model.StatusSuccess, 100, zipPath, zipSize, time.Now(), time.Now(), sign)
	return err
}

func (r *mysqlRepo) UpdateFailed(sign string, errMsg string) error {
	q := `UPDATE tasks SET status=?, error_msg=?, updated_at=?, completed_at=? WHERE sign=?`
	_, err := r.db.Exec(q, model.StatusFailed, errMsg, time.Now(), time.Now(), sign)
	return err
}

func (r *mysqlRepo) IncrementDownloadCount(sign string) error {
	q := `UPDATE tasks SET download_count = download_count + 1, updated_at = ? WHERE sign = ?`
	_, err := r.db.Exec(q, time.Now(), sign)
	return err
}

func (r *mysqlRepo) DeleteBefore(t time.Time) (int64, error) {
	q := `DELETE FROM tasks WHERE created_at < ? AND status = ?`
	result, err := r.db.Exec(q, t, model.StatusSuccess)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
