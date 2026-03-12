package service

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"   // 排队中
	TaskStatusRunning   TaskStatus = "running"   // 执行中
	TaskStatusCompleted TaskStatus = "completed" // 已完成
	TaskStatusFailed    TaskStatus = "failed"    // 失败
)

// CookieImportTask Cookie导入任务
type CookieImportTask struct {
	ID         string                 `json:"id"`
	UserID     int                    `json:"user_id"`
	Status     TaskStatus             `json:"status"`
	Cookies    string                 `json:"-"` // 敏感信息不序列化
	Semester   string                 `json:"semester"`
	Result     *CookieImportResult    `json:"result,omitempty"`
	Error      string                 `json:"error,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	StartedAt  *time.Time             `json:"started_at,omitempty"`
	FinishedAt *time.Time             `json:"finished_at,omitempty"`
	RetryCount int                    `json:"retry_count"`
	MaxRetries int                    `json:"max_retries"`
	ctx        context.Context        `json:"-"`
	cancel     context.CancelFunc     `json:"-"`
}

// CookieImportResult 导入结果
type CookieImportResult struct {
	WeeksParsed    int `json:"weeks_parsed"`
	TotalCells     int `json:"total_cells"`
	AvailableCells int `json:"available_cells"`
	Imported       int `json:"imported"`
}

// TaskQueue 任务队列
type TaskQueue struct {
	tasks       map[string]*CookieImportTask
	queue       chan *CookieImportTask
	maxWorkers  int
	maxQueueSize int
	mu          sync.RWMutex
	wg          sync.WaitGroup
	stopCh      chan struct{}
	stopped     bool
}

var (
	queueInstance *TaskQueue
	queueOnce     sync.Once
)

// GetTaskQueue 获取任务队列单例
func GetTaskQueue() *TaskQueue {
	queueOnce.Do(func() {
		queueInstance = NewTaskQueue(3, 100) // 3个并发工作者，队列长度100
	})
	return queueInstance
}

// NewTaskQueue 创建任务队列
func NewTaskQueue(maxWorkers, maxQueueSize int) *TaskQueue {
	q := &TaskQueue{
		tasks:        make(map[string]*CookieImportTask),
		queue:        make(chan *CookieImportTask, maxQueueSize),
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

// SubmitTask 提交任务
func (q *TaskQueue) SubmitTask(userID int, cookies, semester string) (*CookieImportTask, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.stopped {
		return nil, fmt.Errorf("任务队列已停止")
	}

	// 检查用户是否已有进行中的任务
	for _, task := range q.tasks {
		if task.UserID == userID && (task.Status == TaskStatusPending || task.Status == TaskStatusRunning) {
			return nil, fmt.Errorf("已有进行中的导入任务")
		}
	}

	// 创建任务
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute) // 10分钟超时
	task := &CookieImportTask{
		ID:         generateTaskID(),
		UserID:     userID,
		Status:     TaskStatusPending,
		Cookies:    cookies,
		Semester:   semester,
		CreatedAt:  time.Now(),
		MaxRetries: 3, // 最大重试3次
		ctx:        ctx,
		cancel:     cancel,
	}

	q.tasks[task.ID] = task

	// 发送到队列
	select {
	case q.queue <- task:
		return task, nil
	default:
		delete(q.tasks, task.ID)
		cancel()
		return nil, fmt.Errorf("任务队列已满，请稍后重试")
	}
}

// GetTask 获取任务
func (q *TaskQueue) GetTask(taskID string) (*CookieImportTask, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	task, ok := q.tasks[taskID]
	return task, ok
}

// GetUserTask 获取用户的最新任务
func (q *TaskQueue) GetUserTask(userID int) (*CookieImportTask, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var latestTask *CookieImportTask
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
func (q *TaskQueue) GetUserTasks(userID int) []*CookieImportTask {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var tasks []*CookieImportTask
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
func (q *TaskQueue) processTask(task *CookieImportTask) {
	// 更新状态为执行中
	q.mu.Lock()
	task.Status = TaskStatusRunning
	now := time.Now()
	task.StartedAt = &now
	q.mu.Unlock()

	// 执行任务（带重试）
	result, err := q.executeWithRetry(task)

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

// executeWithRetry 带重试的执行
func (q *TaskQueue) executeWithRetry(task *CookieImportTask) (*CookieImportResult, error) {
	var lastErr error

	for attempt := 0; attempt <= task.MaxRetries; attempt++ {
		if attempt > 0 {
			task.RetryCount = attempt
			// 指数退避重试
			backoff := time.Duration(attempt) * 5 * time.Second
			time.Sleep(backoff)
		}

		// 检查上下文是否取消
		select {
		case <-task.ctx.Done():
			return nil, fmt.Errorf("任务超时或取消")
		default:
		}

		result, err := q.doExecute(task)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// 检查是否应该重试
		if attempt < task.MaxRetries {
			// 更新状态为执行中（重试）
			q.mu.Lock()
			task.Status = TaskStatusRunning
			q.mu.Unlock()
		}
	}

	return nil, fmt.Errorf("重试%d次后失败: %w", task.MaxRetries, lastErr)
}

// doExecute 实际执行
func (q *TaskQueue) doExecute(task *CookieImportTask) (*CookieImportResult, error) {
	// 使用更长的超时设置创建爬虫
	crawler := &EduCrawler{
		client: &http.Client{
			Timeout: 60 * time.Second, // 单次请求60秒超时
		},
		cookies: task.Cookies,
	}

	// 爬取课表
	scheduleMap, err := crawler.CrawlSchedule(task.Semester)
	if err != nil {
		return nil, err
	}

	// 转换为无课时间
	availabilities := ConvertToAvailability(task.UserID, scheduleMap)

	// 保存到数据库
	if len(availabilities) > 0 {
		service := NewAvailabilityService()
		// 删除旧记录
		if err := service.DeleteByUserID(task.UserID); err != nil {
			return nil, fmt.Errorf("删除旧记录失败: %w", err)
		}
		// 批量插入
		if err := service.CreateBatch(task.UserID, availabilities); err != nil {
			return nil, fmt.Errorf("保存无课时间失败: %w", err)
		}
	}

	totalCells := len(scheduleMap) * 5 * 4

	return &CookieImportResult{
		WeeksParsed:    len(scheduleMap),
		TotalCells:     totalCells,
		AvailableCells: len(availabilities),
		Imported:       len(availabilities),
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
			delete(q.tasks, id)
		}
	}
}
