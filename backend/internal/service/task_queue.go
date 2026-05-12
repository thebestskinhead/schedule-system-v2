package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"schedule-system-v2/backend/internal/model"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"   // 排队中
	TaskStatusRunning   TaskStatus = "running"   // 执行中
	TaskStatusCompleted TaskStatus = "completed" // 已完成
	TaskStatusFailed    TaskStatus = "failed"    // 失败
)

// ImportType 导入类型
type ImportType string

const (
	ImportTypeCookie ImportType = "cookie" // Cookie爬虫导入
	ImportTypeExcel  ImportType = "excel"  // Excel文件导入
	ImportTypeManual ImportType = "manual" // 手动录入
)

// ImportTask 通用导入任务
type ImportTask struct {
	ID         string        `json:"id"`
	Type       ImportType    `json:"type"`
	UserID     int           `json:"user_id"`
	Status     TaskStatus    `json:"status"`
	Result     *ImportResult `json:"result,omitempty"`
	Error      string        `json:"error,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
	StartedAt  *time.Time    `json:"started_at,omitempty"`
	FinishedAt *time.Time    `json:"finished_at,omitempty"`
	RetryCount int           `json:"retry_count"`
	MaxRetries int           `json:"max_retries"`

	// Cookie导入参数
	Cookies  string `json:"-"`
	Semester string `json:"semester,omitempty"`

	// Excel导入参数
	FilePath string `json:"-"` // 临时文件路径

	// 手动录入参数
	Availabilities []model.Availability `json:"-"`

	ctx    context.Context
	cancel context.CancelFunc
}

// ImportResult 导入结果
type ImportResult struct {
	WeeksParsed    int `json:"weeks_parsed,omitempty"`
	TotalCells     int `json:"total_cells,omitempty"`
	AvailableCells int `json:"available_cells,omitempty"`
	Imported       int `json:"imported"`
}

// CookieImportTask 保留用于兼容（实际使用ImportTask）
type CookieImportTask = ImportTask

// CookieImportResult 保留用于兼容
type CookieImportResult = ImportResult

// TaskQueue 任务队列
type TaskQueue struct {
	tasks        map[string]*ImportTask
	queue        chan *ImportTask
	maxWorkers   int
	maxQueueSize int
	mu           sync.RWMutex
	wg           sync.WaitGroup
	stopCh       chan struct{}
	stopped      bool
}

var (
	queueInstance *TaskQueue
	queueOnce     sync.Once
)

// GetTaskQueue 获取任务队列单例
func GetTaskQueue() *TaskQueue {
	queueOnce.Do(func() {
		queueInstance = NewTaskQueue(5, 200) // 5个并发工作者，队列长度200
	})
	return queueInstance
}

// NewTaskQueue 创建任务队列
func NewTaskQueue(maxWorkers, maxQueueSize int) *TaskQueue {
	q := &TaskQueue{
		tasks:        make(map[string]*ImportTask),
		queue:        make(chan *ImportTask, maxQueueSize),
		maxWorkers:   maxWorkers,
		maxQueueSize: maxQueueSize,
		stopCh:       make(chan struct{}),
	}

	// 启动工作者
	for i := 0; i < maxWorkers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}

	return q
}

// SubmitTask 提交Cookie导入任务（兼容旧接口）
func (q *TaskQueue) SubmitTask(userID int, cookies, semester string) (*ImportTask, error) {
	return q.SubmitCookieTask(userID, cookies, semester)
}

// SubmitCookieTask 提交Cookie导入任务
func (q *TaskQueue) SubmitCookieTask(userID int, cookies, semester string) (*ImportTask, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.stopped {
		return nil, fmt.Errorf("任务队列已停止")
	}

	// 检查用户是否已有进行中的任务
	if q.hasRunningTaskLocked(userID) {
		return nil, fmt.Errorf("已有进行中的导入任务")
	}

	// 创建任务
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	task := &ImportTask{
		ID:         generateTaskID(),
		Type:       ImportTypeCookie,
		UserID:     userID,
		Status:     TaskStatusPending,
		Cookies:    cookies,
		Semester:   semester,
		CreatedAt:  time.Now(),
		MaxRetries: 3,
		ctx:        ctx,
		cancel:     cancel,
	}

	return q.enqueueTaskLocked(task)
}

// SubmitExcelTask 提交Excel导入任务
func (q *TaskQueue) SubmitExcelTask(userID int, filePath string) (*ImportTask, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.stopped {
		return nil, fmt.Errorf("任务队列已停止")
	}

	// 检查用户是否已有进行中的任务
	if q.hasRunningTaskLocked(userID) {
		return nil, fmt.Errorf("已有进行中的导入任务")
	}

	// 创建任务
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	task := &ImportTask{
		ID:         generateTaskID(),
		Type:       ImportTypeExcel,
		UserID:     userID,
		Status:     TaskStatusPending,
		FilePath:   filePath,
		CreatedAt:  time.Now(),
		MaxRetries: 2,
		ctx:        ctx,
		cancel:     cancel,
	}

	return q.enqueueTaskLocked(task)
}

// SubmitManualTask 提交手动录入任务
func (q *TaskQueue) SubmitManualTask(userID int, weekday, period int, weeks []int) (*ImportTask, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.stopped {
		return nil, fmt.Errorf("任务队列已停止")
	}

	// 检查用户是否已有进行中的任务
	if q.hasRunningTaskLocked(userID) {
		return nil, fmt.Errorf("已有进行中的导入任务")
	}

	// 构建 availabilities
	var availabilities []model.Availability
	for _, week := range weeks {
		availabilities = append(availabilities, model.Availability{
			UserID:  userID,
			Week:    week,
			Weekday: weekday,
			Period:  period,
		})
	}

	// 创建任务
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	task := &ImportTask{
		ID:             generateTaskID(),
		Type:           ImportTypeManual,
		UserID:         userID,
		Status:         TaskStatusPending,
		Availabilities: availabilities,
		CreatedAt:      time.Now(),
		MaxRetries:     1,
		ctx:            ctx,
		cancel:         cancel,
	}

	return q.enqueueTaskLocked(task)
}

// hasRunningTaskLocked 检查用户是否有进行中的任务（调用前需持有锁）
func (q *TaskQueue) hasRunningTaskLocked(userID int) bool {
	for _, task := range q.tasks {
		if task.UserID == userID && (task.Status == TaskStatusPending || task.Status == TaskStatusRunning) {
			return true
		}
	}
	return false
}

// enqueueTaskLocked 将任务加入队列（调用前需持有锁）
func (q *TaskQueue) enqueueTaskLocked(task *ImportTask) (*ImportTask, error) {
	q.tasks[task.ID] = task

	select {
	case q.queue <- task:
		return task, nil
	default:
		delete(q.tasks, task.ID)
		if task.cancel != nil {
			task.cancel()
		}
		return nil, fmt.Errorf("任务队列已满，请稍后重试")
	}
}

// GetTask 获取任务
func (q *TaskQueue) GetTask(taskID string) (*ImportTask, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	task, ok := q.tasks[taskID]
	return task, ok
}

// GetUserTask 获取用户的最新任务
func (q *TaskQueue) GetUserTask(userID int) (*ImportTask, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var latestTask *ImportTask
	for _, task := range q.tasks {
		if task.UserID == userID {
			if latestTask == nil || task.CreatedAt.After(latestTask.CreatedAt) {
				latestTask = task
			}
		}
	}

	return latestTask, latestTask != nil
}

// GetUserTasks 获取用户的所有任务
func (q *TaskQueue) GetUserTasks(userID int) []*ImportTask {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var tasks []*ImportTask
	for _, task := range q.tasks {
		if task.UserID == userID {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// worker 工作者协程
func (q *TaskQueue) worker(id int) {
	defer q.wg.Done()

	for {
		select {
		case <-q.stopCh:
			return
		case task := <-q.queue:
			q.processTask(task)
		}
	}
}

// processTask 处理任务
func (q *TaskQueue) processTask(task *ImportTask) {
	// 更新状态为执行中
	q.mu.Lock()
	task.Status = TaskStatusRunning
	now := time.Now()
	task.StartedAt = &now
	q.mu.Unlock()

	// 根据任务类型执行
	var result *ImportResult
	var err error

	switch task.Type {
	case ImportTypeCookie:
		result, err = q.executeCookieImport(task)
	case ImportTypeExcel:
		result, err = q.executeExcelImport(task)
	case ImportTypeManual:
		result, err = q.executeManualImport(task)
	default:
		err = fmt.Errorf("未知的任务类型: %s", task.Type)
	}

	// 更新任务状态
	q.mu.Lock()
	defer q.mu.Unlock()

	finishTime := time.Now()
	task.FinishedAt = &finishTime

	if err != nil {
		task.Status = TaskStatusFailed
		task.Error = err.Error()
	} else {
		task.Status = TaskStatusCompleted
		task.Result = result
	}

	// 取消上下文
	if task.cancel != nil {
		task.cancel()
	}
}

// executeCookieImport 执行Cookie导入
func (q *TaskQueue) executeCookieImport(task *ImportTask) (*ImportResult, error) {
	var lastErr error

	for attempt := 0; attempt <= task.MaxRetries; attempt++ {
		if attempt > 0 {
			task.RetryCount = attempt
			backoff := time.Duration(attempt) * 5 * time.Second
			time.Sleep(backoff)
		}

		select {
		case <-task.ctx.Done():
			return nil, fmt.Errorf("任务超时或取消")
		default:
		}

		result, err := q.doCookieImport(task)
		if err == nil {
			return result, nil
		}

		lastErr = err

		if attempt < task.MaxRetries {
			q.mu.Lock()
			task.Status = TaskStatusRunning
			q.mu.Unlock()
		}
	}

	return nil, fmt.Errorf("重试%d次后失败: %w", task.MaxRetries, lastErr)
}

// doCookieImport 实际执行Cookie导入
func (q *TaskQueue) doCookieImport(task *ImportTask) (*ImportResult, error) {
	crawler := &EduCrawler{
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		cookies: task.Cookies,
	}

	scheduleMap, err := crawler.CrawlSchedule(task.Semester)
	if err != nil {
		return nil, err
	}

	availabilities := ConvertToAvailability(task.UserID, scheduleMap)

	if len(availabilities) > 0 {
		service := NewAvailabilityService()
		if err := service.DeleteByUserID(task.UserID); err != nil {
			return nil, fmt.Errorf("删除旧记录失败: %w", err)
		}
		if err := service.CreateBatch(task.UserID, availabilities); err != nil {
			return nil, fmt.Errorf("保存无课时间失败: %w", err)
		}
	}

	totalCells := len(scheduleMap) * 5 * 4

	return &ImportResult{
		WeeksParsed:    len(scheduleMap),
		TotalCells:     totalCells,
		AvailableCells: len(availabilities),
		Imported:       len(availabilities),
	}, nil
}

// executeExcelImport 执行Excel导入
func (q *TaskQueue) executeExcelImport(task *ImportTask) (*ImportResult, error) {
	var lastErr error

	for attempt := 0; attempt <= task.MaxRetries; attempt++ {
		if attempt > 0 {
			task.RetryCount = attempt
			backoff := time.Duration(attempt) * 2 * time.Second
			time.Sleep(backoff)
		}

		select {
		case <-task.ctx.Done():
			return nil, fmt.Errorf("任务超时或取消")
		default:
		}

		result, err := q.doExcelImport(task)
		if err == nil {
			return result, nil
		}

		lastErr = err

		if attempt < task.MaxRetries {
			q.mu.Lock()
			task.Status = TaskStatusRunning
			q.mu.Unlock()
		}
	}

	return nil, fmt.Errorf("重试%d次后失败: %w", task.MaxRetries, lastErr)
}

// doExcelImport 实际执行Excel导入
func (q *TaskQueue) doExcelImport(task *ImportTask) (*ImportResult, error) {
	defer os.Remove(task.FilePath) // 完成后删除临时文件

	parser := NewXLSParser()
	availabilities, err := parser.ParseXLS(task.FilePath, task.UserID)
	if err != nil {
		return nil, fmt.Errorf("解析Excel失败: %w", err)
	}

	if len(availabilities) > 0 {
		service := NewAvailabilityService()
		if err := service.DeleteByUserID(task.UserID); err != nil {
			return nil, fmt.Errorf("删除旧记录失败: %w", err)
		}
		if err := service.CreateBatch(task.UserID, availabilities); err != nil {
			return nil, fmt.Errorf("保存无课时间失败: %w", err)
		}
	}

	return &ImportResult{
		Imported: len(availabilities),
	}, nil
}

// executeManualImport 执行手动录入
func (q *TaskQueue) executeManualImport(task *ImportTask) (*ImportResult, error) {
	if len(task.Availabilities) == 0 {
		return &ImportResult{Imported: 0}, nil
	}

	service := NewAvailabilityService()

	// 删除该时段的所有记录
	for _, av := range task.Availabilities {
		service.availabilityDAO.DeleteByUserIDAndTime(task.UserID, av.Week, av.Weekday, av.Period)
	}

	// 批量插入新记录
	if err := service.CreateBatch(task.UserID, task.Availabilities); err != nil {
		return nil, fmt.Errorf("保存无课时间失败: %w", err)
	}

	return &ImportResult{
		Imported: len(task.Availabilities),
	}, nil
}

// Stop 停止队列
func (q *TaskQueue) Stop() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.stopped {
		return
	}

	q.stopped = true
	close(q.stopCh)
	q.wg.Wait()
}

// generateTaskID 生成任务ID
func generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}

// CleanupOldTasks 清理旧任务（保留最近24小时的）
func (q *TaskQueue) CleanupOldTasks() {
	q.mu.Lock()
	defer q.mu.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour)
	for id, task := range q.tasks {
		if task.CreatedAt.Before(cutoff) {
			if task.cancel != nil {
				task.cancel()
			}
			// 清理临时文件
			if task.Type == ImportTypeExcel && task.FilePath != "" {
				os.Remove(task.FilePath)
			}
			delete(q.tasks, id)
		}
	}
}
