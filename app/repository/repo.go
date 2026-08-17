package repository

import (
	"time"

	"github.com/tekintian/googlefonts-tools/app/model"
)

type TaskRepository interface {
	Init() error
	GetBySign(sign string) (*model.Task, error)
	GetByStatus(status string, limit int) ([]*model.Task, error)
	List(offset, limit int) ([]*model.Task, error)
	Count() (int, error)
	Create(task *model.Task) error
	Update(task *model.Task) error
	UpdateProgress(sign string, status model.TaskStatus, progress, doneFiles, totalFiles int) error
	UpdateSuccess(sign string, zipPath string, zipSize int64) error
	UpdateFailed(sign string, errMsg string) error
	IncrementDownloadCount(sign string) error
	DeleteBefore(t time.Time) (int64, error)
}
