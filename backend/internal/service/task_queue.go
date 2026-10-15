package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"schedule-system-v2/backend/internal/db"
	"schedule-system-v2/backend/internal/model"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending      TaskStatus = "pending"       // 排队中
	TaskStatusRunning      TaskStatus = "running"       // 执行中
	TaskStatusWaitingLogin TaskStatus = "waiting_login" // 等待用户扫码登录教务系统
	TaskStatusCompleted    TaskStatus = "completed"     // 已完成
	TaskStatusFailed       TaskStatus = "failed"        // 失败
)

// ImportType 导入类型
type ImportType string

const (
	ImportTypeExcel  ImportType = "excel"  // Excel文件导入
	ImportTypeManual ImportType = "manual" // 手动录入
	ImportTypeEdu    ImportType = "edu"    // 从教务系统下载导入
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

	// Excel导入参数
	FilePath string `json:"-"` // 临时文件路径

	// 教务导入参数
	StudentID string `json:"-"` // 教务学号，用于取缓存的登录会话

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

// SubmitEduTask 提交教务系统导入任务。
// 若已有缓存中的教务登录会话则直接入队执行；否则任务置为等待登录状态，
// 由前端引导用户扫码，扫码成功后再调用 ResumeEduTask 继续。
func (q *TaskQueue) SubmitEduTask(userID int, studentID string) (*ImportTask, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.stopped {
		return nil, fmt.Errorf("任务队列已停止")
	}

	if q.hasRunningTaskLocked(userID) {
		return nil, fmt.Errorf("已有进行中的导入任务")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	task := &ImportTask{
		ID:         generateTaskID(),
		Type:       ImportTypeEdu,
		UserID:     userID,
		StudentID:  studentID,
		Status:     TaskStatusPending,
		CreatedAt:  time.Now(),
		MaxRetries: 1,
		ctx:        ctx,
		cancel:     cancel,
	}

	if GetEduSessionCache().Get(studentID) == nil {
		// 无可用登录会话：等待用户扫码，任务暂不出队
		task.Status = TaskStatusWaitingLogin
		q.tasks[task.ID] = task
		return task, nil
	}

	return q.enqueueTaskLocked(task)
}

// ResumeEduTask 扫码登录成功后继续执行等待中的教务导入任务。
func (q *TaskQueue) ResumeEduTask(taskID string) (*ImportTask, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	task, ok := q.tasks[taskID]
	if !ok {
		return nil, fmt.Errorf("任务不存在")
	}
	if task.Type != ImportTypeEdu {
		return nil, fmt.Errorf("任务类型不匹配")
	}
	if task.Status != TaskStatusWaitingLogin {
		return nil, fmt.Errorf("任务不处于等待登录状态")
	}

	task.Status = TaskStatusPending
	task.Error = ""
	// 等待登录期间原上下文可能已过期，重建执行上下文
	if task.cancel != nil {
		task.cancel()
	}
	task.ctx, task.cancel = context.WithTimeout(context.Background(), 5*time.Minute)
	return q.enqueueTaskLocked(task)
}

// hasRunningTaskLocked 检查用户是否有进行中的任务（调用前需持有锁）
func (q *TaskQueue) hasRunningTaskLocked(userID int) bool {
	for _, task := range q.tasks {
		if task.UserID == userID && (task.Status == TaskStatusPending || task.Status == TaskStatusRunning || task.Status == TaskStatusWaitingLogin) {
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
	case ImportTypeExcel:
		result, err = q.executeExcelImport(task)
	case ImportTypeManual:
		result, err = q.executeManualImport(task)
	case ImportTypeEdu:
		result, err = q.executeEduImport(task)
	default:
		err = fmt.Errorf("未知的任务类型: %s", task.Type)
	}

	// 更新任务状态
	q.mu.Lock()
	defer q.mu.Unlock()

	// 等待登录：保留状态与上下文，待用户扫码后由 ResumeEduTask 重新入队
	if task.Status == TaskStatusWaitingLogin {
		return
	}

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

// executeEduImport 执行教务系统导入：取缓存会话下载课表，落临时文件后复用 Excel 导入流程。
func (q *TaskQueue) executeEduImport(task *ImportTask) (*ImportResult, error) {
	jar := GetEduSessionCache().Get(task.StudentID)
	if jar == nil {
		// 会话缺失或已失效：清掉缓存，转为等待登录
		GetEduSessionCache().Delete(task.StudentID)
		q.mu.Lock()
		task.Status = TaskStatusWaitingLogin
		task.Error = "教务登录已失效，请重新扫码登录"
		q.mu.Unlock()
		return nil, fmt.Errorf("教务登录已失效，请重新扫码登录")
	}

	select {
	case <-task.ctx.Done():
		return nil, fmt.Errorf("任务超时或取消")
	default:
	}

	data, filename, err := NewEduTimetableService().Download(jar)
	if err != nil {
		if errors.Is(err, ErrEduSessionInvalid) {
			GetEduSessionCache().Delete(task.StudentID)
			q.mu.Lock()
			task.Status = TaskStatusWaitingLogin
			task.Error = "教务登录已失效，请重新扫码登录"
			q.mu.Unlock()
		}
		return nil, err
	}

	// 写入临时文件，复用现有 Excel 解析与落库流程
	ext := ".xls"
	if strings.HasSuffix(strings.ToLower(filename), ".xlsx") {
		ext = ".xlsx"
	}
	tempPath := filepath.Join(os.TempDir(), fmt.Sprintf("edu_%d_%d%s", task.UserID, time.Now().UnixNano(), ext))
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return nil, fmt.Errorf("写入课表文件失败: %w", err)
	}

	task.FilePath = tempPath
	return q.doExcelImport(task)
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
		// 全量替换该用户无课表：走 sp_replace_availability（DELETE 全部 + INSERT IGNORE）
		itemsJSON, err := json.Marshal(availabilities)
		if err != nil {
			return nil, fmt.Errorf("序列化无课数据失败: %w", err)
		}
		err = db.WithRetry(context.Background(), func() error {
			_, e := db.GetDB().Exec("CALL sp_replace_availability(?, ?)", task.UserID, string(itemsJSON))
			return e
		})
		if err != nil {
			return nil, fmt.Errorf("保存无课时间失败: %w", err)
		}
	}

	return &ImportResult{
		Imported: len(availabilities),
	}, nil
}

// executeManualImport 执行手动录入
// D 项：手动录入=增量修改该用户该时段（week/weekday/period），不走 sp_replace_availability；
// 采用事务内「DELETE 该用户该时段现有行 + INSERT IGNORE 新值」，杜绝半截数据。
func (q *TaskQueue) executeManualImport(task *ImportTask) (*ImportResult, error) {
	if len(task.Availabilities) == 0 {
		return &ImportResult{Imported: 0}, nil
	}

	userID := task.UserID

	// 聚合需改动的时段 (week, weekday, period)
	type slotKey struct{ week, weekday, period int }
	slots := make(map[slotKey]struct{})
	for _, av := range task.Availabilities {
		slots[slotKey{av.Week, av.Weekday, av.Period}] = struct{}{}
	}

	// I-1 备份要求：删除前记录受影响行快照至应用日志（可恢复）
	if existing, err := NewAvailabilityService().GetMyAvailability(userID); err == nil {
		var backup []model.Availability
		for _, e := range existing {
			if _, ok := slots[slotKey{e.Week, e.Weekday, e.Period}]; ok {
				backup = append(backup, e)
			}
		}
		if len(backup) > 0 {
			fmt.Printf("[I-1 backup] 手动录入将删除用户 %d 的无课记录 %d 行: %+v\n", userID, len(backup), backup)
		}
	}

	// 事务内：删除该用户该时段现有行 + INSERT IGNORE 新值
	tx, err := db.GetDB().Beginx()
	if err != nil {
		return nil, fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	for sk := range slots {
		if _, err := tx.Exec("DELETE FROM availability WHERE user_id=? AND week=? AND weekday=? AND period=?",
			userID, sk.week, sk.weekday, sk.period); err != nil {
			return nil, fmt.Errorf("删除旧无课记录失败: %w", err)
		}
	}
	for _, av := range task.Availabilities {
		if _, err := tx.Exec("INSERT IGNORE INTO availability (user_id, week, weekday, period) VALUES (?, ?, ?, ?)",
			av.UserID, av.Week, av.Weekday, av.Period); err != nil {
			return nil, fmt.Errorf("保存无课记录失败: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	return &ImportResult{Imported: len(task.Availabilities)}, nil
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
