package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/qjfoidnh/BaiduPCS-Go/baidupcs"
	"github.com/qjfoidnh/BaiduPCS-Go/internal/pcsconfig"
	"github.com/qjfoidnh/BaiduPCS-Go/internal/pcsfunctions/pcsdownload"
	"github.com/qjfoidnh/BaiduPCS-Go/pcsutil/converter"
	"github.com/qjfoidnh/BaiduPCS-Go/requester/downloader"
	"github.com/qjfoidnh/BaiduPCS-Go/requester/transfer"
)

type DownloadTaskRunner struct {
	Task *WebDownloadTask

	mu      sync.RWMutex
	done    chan struct{}
	cancel  chan struct{}
	paused  chan struct{}
	der     *downloader.Downloader
	stopped bool
}

func NewDownloadTaskRunner(task *WebDownloadTask) *DownloadTaskRunner {
	return &DownloadTaskRunner{
		Task:   task,
		done:   make(chan struct{}),
		cancel: make(chan struct{}),
		paused: make(chan struct{}),
	}
}

func NewDownloadTaskRunnerWithPCS(task *WebDownloadTask, _, _ *baidupcs.BaiduPCS) *DownloadTaskRunner {
	return &DownloadTaskRunner{
		Task:   task,
		done:   make(chan struct{}),
		cancel: make(chan struct{}),
		paused: make(chan struct{}),
	}
}

func (r *DownloadTaskRunner) Run() {
	r.Task.Status = "running"
	r.Task.UpdatedAt = time.Now()

	go func() {
		defer func() {
			if r.Task.Status == "running" {
				r.Task.Status = "completed"
				r.Task.Progress = 100
			}
			downloadMgr.RecordHistory(r.Task)
			close(r.done)
		}()

		err := r.runDownload()
		r.mu.Lock()
		if err != nil {
			r.Task.Status = "failed"
			r.Task.Error = err.Error()
		}
		r.Task.UpdatedAt = time.Now()
		r.mu.Unlock()
	}()
}

func (r *DownloadTaskRunner) runDownload() error {
	pcs := pcsconfig.Config.ActiveUserBaiduPCS()
	if pcs == nil {
		return fmt.Errorf("未登录或获取网盘实例失败")
	}

	fileInfo, err := pcs.FilesDirectoriesMeta(r.Task.Path)
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %s", err.Error())
	}

	r.Task.Total = fileInfo.Size

	saveDir := r.Task.SavePath
	if err := os.MkdirAll(saveDir, 0777); err != nil {
		return fmt.Errorf("创建保存目录失败: %s", err.Error())
	}

	savePath := filepath.Join(saveDir, fileInfo.Filename)
	r.Task.FileName = fileInfo.Filename
	r.Task.SavePath = savePath

	dlinks, pcsErr := pcsdownload.GetLocateDownloadLinks(pcs, r.Task.Path)
	if pcsErr != nil {
		return fmt.Errorf("获取直链失败: %s", pcsErr.Error())
	}

	cfg := &downloader.Config{
		Mode:                       transfer.RangeGenMode_BlockSize,
		CacheSize:                  pcsconfig.Config.CacheSize,
		BlockSize:                  baidupcs.InitRangeSize,
		MaxRate:                    0,
		InstanceStateStorageFormat: downloader.InstanceStateStorageFormatProto3,
		IsTest:                     false,
		TryHTTP:                    !pcsconfig.Config.EnableHTTPS,
		InstanceStatePath:          savePath + ".BaiduPCS-Go-downloading",
	}

	var writer downloader.Writer
	var file *os.File
	var writeErr error
	writer, file, writeErr = downloader.NewDownloaderWriterByFilename(savePath, os.O_CREATE|os.O_WRONLY, 0666)
	if writeErr != nil {
		return fmt.Errorf("创建文件失败: %s", writeErr.Error())
	}
	defer file.Close()

	// 遍历尝试每个 dlink
	var lastErr error
	for _, d := range dlinks {
		// 先取消之前的 der
		if r.der != nil {
			r.der.Cancel()
			r.der = nil
		}

		dlink := d.String()
		r.der = downloader.NewDownloader(dlink, writer, cfg)
		client := pcs.GetClient()
		r.der.SetClient(client)
		r.der.SetDURLCheckFunc(pcsdownload.BaiduPCSURLCheckFunc)

		r.der.OnDownloadStatusEvent(func(status transfer.DownloadStatuser, workersCallback func(downloader.RangeWorkerFunc)) {
			r.mu.Lock()
			defer r.mu.Unlock()

			if r.Task == nil {
				return
			}
			r.Task.Downloaded = status.Downloaded()
			r.Task.Total = status.TotalSize()
			if r.Task.Total > 0 {
				r.Task.Progress = int(float64(r.Task.Downloaded) / float64(r.Task.Total) * 100)
			}
			r.Task.Speed = converter.ConvertFileSize(status.SpeedsPerSecond(), 2) + "/s"
			left := status.TimeLeft()
			if left > 0 {
				r.Task.TimeLeft = left.String()
			}
			r.Task.Status = "running"
			r.Task.UpdatedAt = time.Now()
		})

		go func() {
			<-r.cancel
			if r.der != nil {
				r.der.Cancel()
			}
		}()

		go func() {
			<-r.paused
			if r.der != nil {
				r.der.Pause()
			}
		}()

		// 60秒超时检测
		success := r.tryDownloadWithTimeout()
		if success {
			return nil
		}

		// 保存错误信息
		if r.der != nil {
			r.der.Cancel()
		}

		// 等待3秒后尝试下一个
		time.Sleep(3 * time.Second)
	}

	if lastErr != nil {
		return fmt.Errorf("所有下载链接都无法下载: %s", lastErr.Error())
	}
	return fmt.Errorf("所有下载链接都无法下载")
}

func (r *DownloadTaskRunner) tryDownloadWithTimeout() bool {
	if r.der == nil {
		return false
	}

	done := make(chan error, 1)
	der := r.der // 保存当前 der 的引用

	go func() {
		err := der.Execute()
		select {
		case done <- err:
		default:
		}
	}()

	for i := 0; i < 60; i++ {
		// 检查是否已取消
		select {
		case <-r.cancel:
			der.Cancel()
			return false
		default:
		}

		time.Sleep(1 * time.Second)

		r.mu.RLock()
		downloaded := r.Task.Downloaded
		r.mu.RUnlock()

		if downloaded > 0 {
			return true // 有进度
		}

		select {
		case err := <-done:
			if err == nil {
				return true // 下载完成
			}
			return false
		default:
		}
	}

	// 超时，取消下载
	der.Cancel()
	return false // 60秒内没有进度
}

func (r *DownloadTaskRunner) Pause() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Task.Status == "running" && !r.stopped {
		r.Task.Status = "paused"
		r.Task.UpdatedAt = time.Now()
		select {
		case r.paused <- struct{}{}:
		default:
		}
	}
}

func (r *DownloadTaskRunner) Resume() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Task.Status == "paused" && !r.stopped {
		r.Task.Status = "running"
		r.Task.UpdatedAt = time.Now()
		if r.der != nil {
			r.der.Resume()
		}
	}
}

func (r *DownloadTaskRunner) Cancel() {
	r.mu.Lock()
	if r.stopped {
		r.mu.Unlock()
		return
	}
	r.stopped = true
	r.Task.Status = "cancelled"
	r.Task.UpdatedAt = time.Now()
	r.mu.Unlock()

	select {
	case r.cancel <- struct{}{}:
	default:
	}
}

func (r *DownloadTaskRunner) GetProgress() (downloaded, total int64, progress int, speed string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Task.Downloaded, r.Task.Total, r.Task.Progress, r.Task.Speed
}

var downloadRunners = &struct {
	sync.RWMutex
	runners map[string]*DownloadTaskRunner
}{runners: make(map[string]*DownloadTaskRunner)}

func RegisterDownloadRunner(taskID string, runner *DownloadTaskRunner) {
	downloadRunners.Lock()
	downloadRunners.runners[taskID] = runner
	downloadRunners.Unlock()
}

func GetDownloadRunner(taskID string) *DownloadTaskRunner {
	downloadRunners.RLock()
	defer downloadRunners.RUnlock()
	return downloadRunners.runners[taskID]
}

func RemoveDownloadRunner(taskID string) {
	downloadRunners.Lock()
	delete(downloadRunners.runners, taskID)
	downloadRunners.Unlock()
}
