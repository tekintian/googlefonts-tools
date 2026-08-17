package service

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/tekintian/googlefonts-tools/app/model"
	"github.com/tekintian/googlefonts-tools/app/repository"
	"github.com/tekintian/googlefonts-tools/utils"
)

type TaskManager struct {
	repo        repository.TaskRepository
	engine      *DownloadEngine
	notifier    *NotifyDispatcher
	subscribers map[string][]chan model.TaskProgress
	mu          sync.RWMutex
	taskCh      chan *model.Task
	workers     int
}

var DefaultTaskManager *TaskManager

func NewTaskManager(repo repository.TaskRepository, engine *DownloadEngine, notifier *NotifyDispatcher, workers int) *TaskManager {
	if workers <= 0 {
		workers = 3
	}
	tm := &TaskManager{
		repo:        repo,
		engine:      engine,
		notifier:    notifier,
		subscribers: make(map[string][]chan model.TaskProgress),
		taskCh:      make(chan *model.Task, 100),
		workers:     workers,
	}
	DefaultTaskManager = tm
	return tm
}

func (tm *TaskManager) Start() {
	for i := 0; i < tm.workers; i++ {
		go tm.worker(i)
	}
	fmt.Printf("[TaskManager] started %d workers\n", tm.workers)
}

func (tm *TaskManager) worker(id int) {
	for task := range tm.taskCh {
		tm.processTask(task)
	}
}

func (tm *TaskManager) processTask(task *model.Task) {
	fmt.Printf("[Worker] processing task: sign=%s font=%s\n", task.Sign, task.FontName)

	now := time.Now()
	task.Status = model.StatusRunning
	task.UpdatedAt = now
	tm.repo.UpdateProgress(task.Sign, model.StatusRunning, 0, 0, 0)

	onProgress := func(p model.TaskProgress) {
		tm.repo.UpdateProgress(p.Sign, p.Status, p.Progress, p.DoneFiles, p.TotalFiles)
		tm.publish(p)
	}

	err := tm.engine.Download(task, onProgress)
	if err != nil {
		fmt.Printf("[Worker] task failed: sign=%s error=%v\n", task.Sign, err)
		tm.repo.UpdateFailed(task.Sign, err.Error())
		tm.publish(model.TaskProgress{
			Sign:     task.Sign,
			Status:   model.StatusFailed,
			ErrorMsg: err.Error(),
		})
		if tm.notifier != nil {
			updatedTask, _ := tm.repo.GetBySign(task.Sign)
			if updatedTask != nil {
				go tm.notifier.Dispatch(updatedTask)
			}
		}
		return
	}

	tm.repo.UpdateSuccess(task.Sign, task.ZipPath, task.ZipSize)
	tm.publish(model.TaskProgress{
		Sign:       task.Sign,
		Status:     model.StatusSuccess,
		Progress:   100,
		DoneFiles:  task.TotalFiles,
		TotalFiles: task.TotalFiles,
	})

	fmt.Printf("[Worker] task completed: sign=%s font=%s zip=%s\n", task.Sign, task.FontName, task.ZipPath)

	if tm.notifier != nil {
		updatedTask, _ := tm.repo.GetBySign(task.Sign)
		if updatedTask != nil {
			go tm.notifier.Dispatch(updatedTask)
		}
	}
}

func (tm *TaskManager) Submit(url string, notifyConfig string) (*model.Task, error) {
	sign := utils.Md5(url)

	existing, err := tm.repo.GetBySign(sign)
	if err != nil {
		return nil, fmt.Errorf("query task error: %w", err)
	}

	if existing != nil {
		switch existing.Status {
		case model.StatusSuccess:
			if _, err := os.Stat(existing.ZipPath); err == nil {
				fmt.Printf("[TaskManager] cache hit: sign=%s\n", sign)
				return existing, nil
			}
			fmt.Printf("[TaskManager] cache expired, re-downloading: sign=%s\n", sign)
			now := time.Now()
			existing.Status = model.StatusPending
			existing.Progress = 0
			existing.UpdatedAt = now
			existing.ZipPath = ""
			existing.ZipSize = 0
			existing.ErrorMsg = ""
			existing.CompletedAt = nil
			tm.repo.Update(existing)

		case model.StatusRunning, model.StatusPending:
			return existing, nil

		case model.StatusFailed:
			now := time.Now()
			existing.Status = model.StatusPending
			existing.Progress = 0
			existing.UpdatedAt = now
			existing.ErrorMsg = ""
			existing.CompletedAt = nil
			tm.repo.Update(existing)
		}

		tm.taskCh <- existing
		return existing, nil
	}

	now := time.Now()
	task := &model.Task{
		Sign:         sign,
		URL:          url,
		Status:       model.StatusPending,
		NotifyConfig: notifyConfig,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := tm.repo.Create(task); err != nil {
		return nil, fmt.Errorf("create task error: %w", err)
	}

	tm.taskCh <- task
	fmt.Printf("[TaskManager] new task submitted: sign=%s url=%s\n", sign, url)
	return task, nil
}

func (tm *TaskManager) SubmitSync(url string, notifyConfig string) (*model.Task, error) {
	sign := utils.Md5(url)

	existing, err := tm.repo.GetBySign(sign)
	if err != nil {
		return nil, fmt.Errorf("query task error: %w", err)
	}

	if existing != nil && existing.Status == model.StatusSuccess {
		if _, err := os.Stat(existing.ZipPath); err == nil {
			return existing, nil
		}
	}

	now := time.Now()
	task := &model.Task{
		Sign:         sign,
		URL:          url,
		Status:       model.StatusPending,
		NotifyConfig: notifyConfig,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if existing != nil {
		task = existing
		task.Status = model.StatusPending
		task.Progress = 0
		task.UpdatedAt = now
		task.ErrorMsg = ""
		task.CompletedAt = nil
		tm.repo.Update(task)
	} else {
		if err := tm.repo.Create(task); err != nil {
			return nil, fmt.Errorf("create task error: %w", err)
		}
	}

	tm.processTask(task)

	result, _ := tm.repo.GetBySign(sign)
	if result == nil {
		return task, nil
	}
	return result, nil
}

func (tm *TaskManager) Subscribe(sign string) chan model.TaskProgress {
	ch := make(chan model.TaskProgress, 10)
	tm.mu.Lock()
	tm.subscribers[sign] = append(tm.subscribers[sign], ch)
	tm.mu.Unlock()
	return ch
}

func (tm *TaskManager) Unsubscribe(sign string, ch chan model.TaskProgress) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	subs := tm.subscribers[sign]
	for i, s := range subs {
		if s == ch {
			tm.subscribers[sign] = append(subs[:i], subs[i+1:]...)
			close(s)
			break
		}
	}
}

func (tm *TaskManager) publish(progress model.TaskProgress) {
	tm.mu.RLock()
	subs := tm.subscribers[progress.Sign]
	channels := make([]chan model.TaskProgress, len(subs))
	copy(channels, subs)
	tm.mu.RUnlock()

	for _, ch := range channels {
		select {
		case ch <- progress:
		default:
		}
	}
}

func (tm *TaskManager) GetTask(sign string) (*model.Task, error) {
	return tm.repo.GetBySign(sign)
}

func (tm *TaskManager) ListTasks(offset, limit int) ([]*model.Task, error) {
	return tm.repo.List(offset, limit)
}

func (tm *TaskManager) IncrementDownloadCount(sign string) error {
	return tm.repo.IncrementDownloadCount(sign)
}
