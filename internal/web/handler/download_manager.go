package handler

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/qjfoidnh/BaiduPCS-Go/pcsutil"
)

const (
	MaxHistoryItems = 500
)

type WebDownloadTask struct {
	ID         string    `json:"id"`
	Path       string    `json:"path"`
	FileName   string    `json:"file_name"`
	SavePath   string    `json:"save_path"`
	Status     string    `json:"status"`
	Progress   int       `json:"progress"`
	Speed      string    `json:"speed"`
	Downloaded int64     `json:"downloaded"`
	Total      int64     `json:"total"`
	TimeLeft   string    `json:"time_left"`
	Error      string    `json:"error,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type DownloadHistory struct {
	Path      string    `json:"path"`
	SavePath  string    `json:"save_path"`
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	DoneAt    time.Time `json:"done_at"`
}

type DownloadManager struct {
	mu      sync.RWMutex
	tasks   map[string]*WebDownloadTask
	history []*DownloadHistory
}

var downloadMgr = &DownloadManager{
	tasks:   make(map[string]*WebDownloadTask),
	history: nil,
}

func GetHistoryFilePath() string {
	return filepath.Join(pcsutil.ExecutablePath(), "download_history.json")
}

func (dm *DownloadManager) CreateTask(pcspath, savepath string) *WebDownloadTask {
	taskID := fmt.Sprintf("dl_%d_%d", time.Now().UnixNano(), rand.Intn(10000))

	task := &WebDownloadTask{
		ID:        taskID,
		Path:      pcspath,
		FileName:  filepath.Base(pcspath),
		SavePath:  savepath,
		Status:    "pending",
		Progress:  0,
		Speed:     "",
		CreatedAt: time.Now(),
	}

	dm.mu.Lock()
	dm.tasks[taskID] = task
	dm.mu.Unlock()

	return task
}

func (dm *DownloadManager) GetTask(taskID string) *WebDownloadTask {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.tasks[taskID]
}

func (dm *DownloadManager) GetAllTasks() []*WebDownloadTask {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	tasks := make([]*WebDownloadTask, 0, len(dm.tasks))
	for _, t := range dm.tasks {
		tasks = append(tasks, t)
	}
	return tasks
}

func (dm *DownloadManager) RemoveTask(taskID string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	delete(dm.tasks, taskID)
}

func (dm *DownloadManager) UpdateTask(taskID string, update func(*WebDownloadTask)) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	if task, ok := dm.tasks[taskID]; ok {
		update(task)
		task.UpdatedAt = time.Now()
	}
}

func (dm *DownloadManager) LoadHistory() []*DownloadHistory {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	path := GetHistoryFilePath()
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		dm.history = make([]*DownloadHistory, 0)
		return dm.history
	}

	var history []*DownloadHistory
	if err := json.Unmarshal(data, &history); err != nil {
		dm.history = make([]*DownloadHistory, 0)
		return dm.history
	}

	dm.history = history
	return dm.history
}

func (dm *DownloadManager) SaveHistory(history []*DownloadHistory) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.history = history

	path := GetHistoryFilePath()
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return
	}

	os.WriteFile(path, data, 0644)
}

func (dm *DownloadManager) RecordHistory(task *WebDownloadTask) {
	history := dm.LoadHistory()

	record := &DownloadHistory{
		Path:      task.Path,
		SavePath:  task.SavePath,
		FileName:  task.FileName,
		FileSize:  task.Total,
		Status:    task.Status,
		CreatedAt: task.CreatedAt,
		DoneAt:    time.Now(),
	}

	history = append(history, record)

	if len(history) > MaxHistoryItems {
		history = history[len(history)-MaxHistoryItems:]
	}

	dm.SaveHistory(history)
}

func (dm *DownloadManager) ClearHistory() {
	dm.SaveHistory(make([]*DownloadHistory, 0))
}

func (dm *DownloadManager) DeleteHistory(taskID string) bool {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if dm.history == nil {
		path := GetHistoryFilePath()
		data, err := os.ReadFile(path)
		if err != nil || len(data) == 0 {
			return false
		}
		json.Unmarshal(data, &dm.history)
	}

	for i, h := range dm.history {
		if h.Path == taskID || h.FileName == taskID {
			dm.history = append(dm.history[:i], dm.history[i+1:]...)
			dm.mu.Unlock()
			dm.SaveHistory(dm.history)
			dm.mu.Lock()
			return true
		}
	}
	return false
}
