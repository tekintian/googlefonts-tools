package model

import "time"

type TaskStatus string

const (
	StatusPending TaskStatus = "pending"
	StatusRunning TaskStatus = "running"
	StatusSuccess TaskStatus = "success"
	StatusFailed  TaskStatus = "failed"
)

type Task struct {
	Sign          string     `json:"sign"`
	URL           string     `json:"url"`
	FontName      string     `json:"font_name"`
	Status        TaskStatus `json:"status"`
	Progress      int        `json:"progress"`
	TotalFiles    int        `json:"total_files"`
	DoneFiles     int        `json:"done_files"`
	ZipPath       string     `json:"zip_path"`
	ZipSize       int64      `json:"zip_size"`
	ErrorMsg      string     `json:"error_msg,omitempty"`
	NotifyConfig  string     `json:"notify_config,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	DownloadCount int        `json:"download_count"`
}

type TaskProgress struct {
	Sign       string     `json:"sign"`
	Status     TaskStatus `json:"status"`
	Progress   int        `json:"progress"`
	DoneFiles  int        `json:"done_files"`
	TotalFiles int        `json:"total_files"`
	ErrorMsg   string     `json:"error_msg,omitempty"`
}
