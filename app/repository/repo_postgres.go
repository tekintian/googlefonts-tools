package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/tekintian/googlefonts-tools/app/model"
	"github.com/tekintian/googlefonts-tools/utils/db"
)

type postgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(database *sql.DB) TaskRepository {
	return &postgresRepo{db: database}
}

func (r *postgresRepo) Init() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			sign          VARCHAR(32) PRIMARY KEY,
			url           TEXT NOT NULL,
			font_name     VARCHAR(128) DEFAULT '',
			status        VARCHAR(16) DEFAULT 'pending',
			progress      INTEGER DEFAULT 0,
			total_files   INTEGER DEFAULT 0,
			done_files    INTEGER DEFAULT 0,
			zip_path      VARCHAR(512) DEFAULT '',
			zip_size      BIGINT DEFAULT 0,
			error_msg     TEXT DEFAULT '',
			notify_config TEXT DEFAULT '',
			created_at    TIMESTAMP NOT NULL,
			updated_at    TIMESTAMP NOT NULL,
			completed_at  TIMESTAMP,
			download_count INTEGER DEFAULT 0
		)
	`)
	if err != nil {
		return fmt.Errorf("postgres create table error: %w", err)
	}
	r.db.Exec(`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)`)
	r.db.Exec(`CREATE INDEX IF NOT EXISTS idx_tasks_font_name ON tasks(font_name)`)
	r.db.Exec(`CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at)`)
	return nil
}

func (r *postgresRepo) pq(query string) string {
	return db.AdaptPlaceholders(query)
}

func (r *postgresRepo) GetBySign(sign string) (*model.Task, error) {
	q := r.pq(buildSelectQuery() + ` WHERE sign = ?`)
	return scanOne(r.db.QueryRow(q, sign))
}

func (r *postgresRepo) GetByStatus(status string, limit int) ([]*model.Task, error) {
	q := r.pq(buildSelectQuery() + ` WHERE status = ? ORDER BY created_at DESC LIMIT ?`)
	rows, err := r.db.Query(q, status, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *postgresRepo) List(offset, limit int) ([]*model.Task, error) {
	q := r.pq(buildSelectQuery() + ` ORDER BY created_at DESC LIMIT ? OFFSET ?`)
	rows, err := r.db.Query(q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *postgresRepo) Count() (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM tasks`).Scan(&count)
	return count, err
}

func (r *postgresRepo) Create(task *model.Task) error {
	q := r.pq(`INSERT INTO tasks (sign, url, font_name, status, progress,
		total_files, done_files, zip_path, zip_size, error_msg,
		notify_config, created_at, updated_at, completed_at, download_count)
		VALUES (` + buildPlaceholders(15) + `)`)
	_, err := r.db.Exec(q,
		task.Sign, task.URL, task.FontName, task.Status, task.Progress,
		task.TotalFiles, task.DoneFiles, task.ZipPath, task.ZipSize, task.ErrorMsg,
		task.NotifyConfig, task.CreatedAt, task.UpdatedAt, task.CompletedAt, task.DownloadCount,
	)
	return err
}

func (r *postgresRepo) Update(task *model.Task) error {
	q := r.pq(`UPDATE tasks SET url=?, font_name=?, status=?, progress=?,
		total_files=?, done_files=?, zip_path=?, zip_size=?, error_msg=?,
		notify_config=?, updated_at=?, completed_at=?, download_count=?
		WHERE sign=?`)
	_, err := r.db.Exec(q,
		task.URL, task.FontName, task.Status, task.Progress,
		task.TotalFiles, task.DoneFiles, task.ZipPath, task.ZipSize, task.ErrorMsg,
		task.NotifyConfig, task.UpdatedAt, task.CompletedAt, task.DownloadCount,
		task.Sign,
	)
	return err
}

func (r *postgresRepo) UpdateProgress(sign string, status model.TaskStatus, progress, doneFiles, totalFiles int) error {
	q := r.pq(`UPDATE tasks SET status=?, progress=?, done_files=?, total_files=?, updated_at=? WHERE sign=?`)
	_, err := r.db.Exec(q, status, progress, doneFiles, totalFiles, time.Now(), sign)
	return err
}

func (r *postgresRepo) UpdateSuccess(sign string, zipPath string, zipSize int64) error {
	q := r.pq(`UPDATE tasks SET status=?, progress=?, zip_path=?, zip_size=?, updated_at=?, completed_at=? WHERE sign=?`)
	_, err := r.db.Exec(q, model.StatusSuccess, 100, zipPath, zipSize, time.Now(), time.Now(), sign)
	return err
}

func (r *postgresRepo) UpdateFailed(sign string, errMsg string) error {
	q := r.pq(`UPDATE tasks SET status=?, error_msg=?, updated_at=?, completed_at=? WHERE sign=?`)
	_, err := r.db.Exec(q, model.StatusFailed, errMsg, time.Now(), time.Now(), sign)
	return err
}

func (r *postgresRepo) IncrementDownloadCount(sign string) error {
	q := r.pq(`UPDATE tasks SET download_count = download_count + 1, updated_at = ? WHERE sign = ?`)
	_, err := r.db.Exec(q, time.Now(), sign)
	return err
}

func (r *postgresRepo) DeleteBefore(t time.Time) (int64, error) {
	q := r.pq(`DELETE FROM tasks WHERE created_at < ? AND status = ?`)
	result, err := r.db.Exec(q, t, model.StatusSuccess)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
