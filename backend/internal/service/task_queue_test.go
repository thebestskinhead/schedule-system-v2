package service

import (
	"testing"
	"time"
)

// 无教务会话时提交教务导入，任务应进入等待登录状态且不占用执行槽位。
func TestSubmitEduTaskWithoutSessionWaitsLogin(t *testing.T) {
	q := NewTaskQueue(1, 10)
	defer q.Stop()

	task, err := q.SubmitEduTask(9001, "9001-no-session")
	if err != nil {
		t.Fatalf("提交失败: %v", err)
	}
	if task.Status != TaskStatusWaitingLogin {
		t.Fatalf("状态 = %s, 期望 waiting_login", task.Status)
	}
	if q.isExpired(task, time.Now()) {
		t.Fatal("刚创建的任务不应判定为过期")
	}
}

// 同一用户重复提交不再被拒绝，旧任务应被作废让路（原实现在此返回 429，导致死锁）。
func TestResubmitSupersedesPreviousTask(t *testing.T) {
	q := NewTaskQueue(1, 10)
	defer q.Stop()

	first, err := q.SubmitEduTask(9002, "9002-no-session")
	if err != nil {
		t.Fatalf("首次提交失败: %v", err)
	}

	second, err := q.SubmitEduTask(9002, "9002-no-session")
	if err != nil {
		t.Fatalf("重复提交被拒绝（死锁未解）: %v", err)
	}
	if first.ID == second.ID {
		t.Fatal("应创建新任务")
	}

	old, ok := q.GetTask(first.ID)
	if !ok {
		t.Fatal("旧任务不应被删除")
	}
	if old.Status != TaskStatusCancelled {
		t.Fatalf("旧任务状态 = %s, 期望 cancelled", old.Status)
	}
	if second.Status != TaskStatusWaitingLogin {
		t.Fatalf("新任务状态 = %s, 期望 waiting_login", second.Status)
	}
}

// 取消任务后应释放槽位，且可立即重新提交。
func TestCancelTaskReleasesSlot(t *testing.T) {
	q := NewTaskQueue(1, 10)
	defer q.Stop()

	task, err := q.SubmitEduTask(9003, "9003-no-session")
	if err != nil {
		t.Fatalf("提交失败: %v", err)
	}

	if err := q.CancelTask(task.ID, 9003); err != nil {
		t.Fatalf("取消失败: %v", err)
	}
	got, _ := q.GetTask(task.ID)
	if got.Status != TaskStatusCancelled {
		t.Fatalf("取消后状态 = %s, 期望 cancelled", got.Status)
	}

	if _, err := q.SubmitEduTask(9003, "9003-no-session"); err != nil {
		t.Fatalf("取消后无法重新提交: %v", err)
	}

	// 越权取消应被拒绝
	if err := q.CancelTask(task.ID, 99999); err == nil {
		t.Fatal("取消他人任务应返回错误")
	}
}

// 等待登录超过 TTL 应自动判失败并释放槽位。
func TestSweepExpiresWaitingLogin(t *testing.T) {
	q := NewTaskQueue(1, 10)
	defer q.Stop()

	task, err := q.SubmitEduTask(9004, "9004-no-session")
	if err != nil {
		t.Fatalf("提交失败: %v", err)
	}

	// 回拨创建时间，模拟长时间无人扫码
	q.mu.Lock()
	task.CreatedAt = time.Now().Add(-waitingLoginTTL - time.Minute)
	q.mu.Unlock()

	q.SweepExpiredTasks()

	got, _ := q.GetTask(task.ID)
	if got.Status != TaskStatusFailed {
		t.Fatalf("超时后状态 = %s, 期望 failed", got.Status)
	}
	if got.Error == "" {
		t.Fatal("超时任务应记录原因")
	}

	// 释放后应能重新提交
	if _, err := q.SubmitEduTask(9004, "9004-no-session"); err != nil {
		t.Fatalf("超时释放后无法重新提交: %v", err)
	}
}

// 执行中超时的任务应被清理为失败，避免永久占用槽位。
func TestSweepExpiresStaleRunning(t *testing.T) {
	q := NewTaskQueue(1, 10)
	defer q.Stop()

	started := time.Now().Add(-taskExecTimeout - time.Minute)
	task := &ImportTask{
		ID:        "task_stale_running",
		Type:      ImportTypeExcel,
		UserID:    9006,
		Status:    TaskStatusRunning,
		CreatedAt: started,
		StartedAt: &started,
	}
	q.mu.Lock()
	q.tasks[task.ID] = task
	q.mu.Unlock()

	q.SweepExpiredTasks()

	got, _ := q.GetTask(task.ID)
	if got.Status != TaskStatusFailed {
		t.Fatalf("停滞任务状态 = %s, 期望 failed", got.Status)
	}
}

// 执行路径 panic 时任务须落到失败终态，而不是永久停留在 running。
func TestProcessTaskRecoversFromPanic(t *testing.T) {
	q := NewTaskQueue(1, 10)
	defer q.Stop()

	// ctx 为 nil 会让执行路径在 ctx.Done() 处触发 panic，用于验证 recover 兜底
	task := &ImportTask{
		ID:     "task_panic",
		Type:   ImportTypeExcel,
		UserID: 9007,
		Status: TaskStatusPending,
	}
	q.mu.Lock()
	q.tasks[task.ID] = task
	q.mu.Unlock()

	q.processTask(task)

	got, ok := q.GetTask(task.ID)
	if !ok {
		t.Fatal("任务不应丢失")
	}
	if got.Status != TaskStatusFailed {
		t.Fatalf("panic 后状态 = %s, 期望 failed（不得停留在 running）", got.Status)
	}
	if got.Error == "" {
		t.Fatal("panic 后应记录错误信息")
	}
	if got.FinishedAt == nil {
		t.Fatal("panic 后应记录结束时间")
	}
}

// 已取消的任务被 worker 取出时不应再执行。
func TestCancelledTaskSkippedOnDequeue(t *testing.T) {
	q := NewTaskQueue(1, 10)
	defer q.Stop()

	task := &ImportTask{
		ID:        "task_cancelled_skip",
		Type:      ImportTypeExcel,
		UserID:    9008,
		Status:    TaskStatusCancelled,
		CreatedAt: time.Now(),
	}
	q.mu.Lock()
	q.tasks[task.ID] = task
	q.mu.Unlock()

	q.processTask(task)

	got, _ := q.GetTask(task.ID)
	if got.Status != TaskStatusCancelled {
		t.Fatalf("状态 = %s, 期望保持 cancelled", got.Status)
	}
	if got.StartedAt != nil {
		t.Fatal("已取消的任务不应被标记为已开始执行")
	}
}

// 停止队列不得阻塞：原实现持锁等待 wg.Wait()，而 worker 退出路径也要取该锁。
func TestStopDoesNotDeadlock(t *testing.T) {
	q := NewTaskQueue(2, 10)

	done := make(chan struct{})
	go func() {
		q.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Stop 阻塞超过 3 秒，疑似持锁等待导致死锁")
	}
}
