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
	"sync/atomic"
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
	TaskStatusCancelled    TaskStatus = "cancelled"     // 已取消
)

// 任务生命周期参数。执行类超时用于兜住卡死，避免任务与并发槽位被永久占用。
const (
	taskExecTimeout   = 2 * time.Minute  // 排队与执行的最长耗时，超时判失败
	waitingLoginTTL   = 2 * time.Minute  // 等待扫码登录的最长时长
	taskRetention     = 24 * time.Hour   // 终态任务的保留时长
	taskSweepInterval = 20 * time.Second // 过期任务扫描间隔
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
		queueInstance.StartCleanup()
	})
	return queueInstance
}

// StartCleanup 启动后台清理协程，周期回收超时任务与过期终态任务。
func (q *TaskQueue) StartCleanup() {
	go func() {
		ticker := time.NewTicker(taskSweepInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				q.SweepExpiredTasks()
			case <-q.stopCh:
				return
			}
		}
	}()
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

	// 同用户再次提交时作废其未完成任务，避免旧任务占位导致无法重试
	q.supersedeActiveTasksLocked(userID)

	// 创建任务
	ctx, cancel := context.WithTimeout(context.Background(), taskExecTimeout)
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

	// 同用户再次提交时作废其未完成任务，避免旧任务占位导致无法重试
	q.supersedeActiveTasksLocked(userID)

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
	ctx, cancel := context.WithTimeout(context.Background(), taskExecTimeout)
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

	// 同用户再次提交时作废其未完成任务，避免旧任务占位导致无法重试
	q.supersedeActiveTasksLocked(userID)

	ctx, cancel := context.WithTimeout(context.Background(), taskExecTimeout)
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
	task.ctx, task.cancel = context.WithTimeout(context.Background(), taskExecTimeout)
	return q.enqueueTaskLocked(task)
}

// statusActive 表示任务处于未完成状态，占用并发槽位。
func (t *ImportTask) statusActive() bool {
	return t.Status == TaskStatusPending || t.Status == TaskStatusRunning || t.Status == TaskStatusWaitingLogin
}

// supersedeActiveTasksLocked 作废指定用户的全部未完成任务，为新提交让路（调用前需持有锁）。
// 用户无需等待旧任务结束即可重试，避免旧任务占位造成无法再次提交。
func (q *TaskQueue) supersedeActiveTasksLocked(userID int) {
	for _, task := range q.tasks {
		if task.UserID != userID || !task.statusActive() {
			continue
		}
		task.Status = TaskStatusCancelled
		task.Error = "已被新的导入任务替代"
		now := time.Now()
		task.FinishedAt = &now
		if task.cancel != nil {
			task.cancel()
		}
		if task.Type == ImportTypeExcel && task.FilePath != "" {
			os.Remove(task.FilePath)
			task.FilePath = ""
		}
	}
}

// CancelTask 主动取消指定用户的未完成任务并释放其占用的资源。
func (q *TaskQueue) CancelTask(taskID string, userID int) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	task, ok := q.tasks[taskID]
	if !ok {
		return fmt.Errorf("任务不存在")
	}
	if task.UserID != userID {
		return fmt.Errorf("无权操作此任务")
	}
	if !task.statusActive() {
		return fmt.Errorf("任务已结束，无需取消")
	}

	task.Status = TaskStatusCancelled
	task.Error = "任务已取消"
	now := time.Now()
	task.FinishedAt = &now
	if task.cancel != nil {
		task.cancel()
	}
	if task.Type == ImportTypeExcel && task.FilePath != "" {
		os.Remove(task.FilePath)
		task.FilePath = ""
	}
	return nil
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
	// 已被取消的任务（重复提交或主动取消）不再执行
	q.mu.Lock()
	if task.Status == TaskStatusCancelled {
		q.mu.Unlock()
		return
	}
	task.Status = TaskStatusRunning
	now := time.Now()
	task.StartedAt = &now
	q.mu.Unlock()

	// 执行过程 panic 时任务必须落到失败终态，否则会永久停留在执行中，
	// 并使并发槽位无法释放
	defer func() {
		if r := recover(); r != nil {
			q.finishTask(task, nil, fmt.Errorf("任务执行异常: %v", r))
		}
	}()

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

	q.mu.RLock()
	waiting := task.Status == TaskStatusWaitingLogin
	q.mu.RUnlock()

	// 等待登录：保留状态与上下文，待用户扫码后由 ResumeEduTask 重新入队
	if waiting {
		return
	}

	q.finishTask(task, result, err)
}

// finishTask 写入任务终态。已被取消或已进入终态的任务不会被覆盖。
func (q *TaskQueue) finishTask(task *ImportTask, result *ImportResult, err error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	switch task.Status {
	case TaskStatusCancelled, TaskStatusCompleted, TaskStatusFailed:
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

	data, filename, err := NewEduTimetableService().Download(task.ctx, jar)
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
		err = db.WithRetry(task.ctx, func() error {
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
	if q.stopped {
		q.mu.Unlock()
		return
	}
	q.stopped = true
	close(q.stopCh)
	q.mu.Unlock()

	// 必须先释放锁再等待：worker 收尾时也要获取该锁
	q.wg.Wait()
}

var taskIDSeq uint64

// generateTaskID 生成任务ID。时间戳受系统时钟粒度限制（Windows 约 1ms），
// 连续提交可能落在同一刻度上，故追加自增序号保证唯一。
func generateTaskID() string {
	seq := atomic.AddUint64(&taskIDSeq, 1)
	return fmt.Sprintf("task_%d_%d", time.Now().UnixNano(), seq)
}

// SweepExpiredTasks 回收超时任务与过期终态任务。
// 超时任务置为失败并释放其占用的槽位；终态任务保留一段时间后移除。
func (q *TaskQueue) SweepExpiredTasks() {
	now := time.Now()

	q.mu.Lock()
	defer q.mu.Unlock()

	for id, task := range q.tasks {
		switch task.Status {
		case TaskStatusPending, TaskStatusRunning, TaskStatusWaitingLogin:
			reason := q.expirationReason(task, now)
			if reason == "" {
				continue
			}
			task.Status = TaskStatusFailed
			task.Error = reason
			finished := now
			task.FinishedAt = &finished
			if task.cancel != nil {
				task.cancel()
			}
			if task.Type == ImportTypeExcel && task.FilePath != "" {
				os.Remove(task.FilePath)
				task.FilePath = ""
			}
		default:
			if !task.CreatedAt.Add(taskRetention).After(now) {
				if task.Type == ImportTypeExcel && task.FilePath != "" {
					os.Remove(task.FilePath)
				}
				delete(q.tasks, id)
			}
		}
	}
}

// isExpired 判断任务是否已超出其当前状态允许的最长时长。
func (q *TaskQueue) isExpired(task *ImportTask, now time.Time) bool {
	return q.expirationReason(task, now) != ""
}

// expirationReason 返回任务超时的原因，未超时返回空串。
func (q *TaskQueue) expirationReason(task *ImportTask, now time.Time) string {
	switch task.Status {
	case TaskStatusWaitingLogin:
		if now.Sub(task.CreatedAt) > waitingLoginTTL {
			return "等待扫码登录超时，请重新发起导入"
		}
	case TaskStatusRunning:
		start := task.CreatedAt
		if task.StartedAt != nil {
			start = *task.StartedAt
		}
		if now.Sub(start) > taskExecTimeout {
			return "导入执行超时，请重试"
		}
	case TaskStatusPending:
		if now.Sub(task.CreatedAt) > taskExecTimeout {
			return "任务排队超时，请重试"
		}
	}
	return ""
}
