package service

import (
	"fmt"
	"sync"
	"time"
)

// MailJob 一封待发送的邮件（内容已渲染完成，worker 只负责投递与重试）
type MailJob struct {
	To      string
	Subject string
	Body    string
}

// MailQueue 邮件发送队列：提交即返回，实际投递由后台 worker 异步完成。
// 队列满时降级为同步发送，保证邮件不丢失。
type MailQueue struct {
	queue    chan *MailJob
	workers  int
	maxRetry int
	sender   func(job *MailJob) error
	stopCh   chan struct{}
	wg       sync.WaitGroup
	once     sync.Once
}

var (
	mailQueueInstance *MailQueue
	mailQueueOnce     sync.Once
)

// GetMailQueue 获取邮件队列单例（4 个并发 worker，队列容量 500，失败重试 2 次）
func GetMailQueue() *MailQueue {
	mailQueueOnce.Do(func() {
		mailQueueInstance = newMailQueue(4, 500, 2, defaultMailSender)
	})
	return mailQueueInstance
}

// newMailQueue 创建邮件队列并启动 worker
func newMailQueue(workers, capacity, maxRetry int, sender func(*MailJob) error) *MailQueue {
	q := &MailQueue{
		queue:    make(chan *MailJob, capacity),
		workers:  workers,
		maxRetry: maxRetry,
		sender:   sender,
		stopCh:   make(chan struct{}),
	}
	for i := 0; i < workers; i++ {
		q.wg.Add(1)
		go q.worker()
	}
	return q
}

// Enqueue 提交邮件任务（非阻塞）。队列已满时同步发送兜底，避免丢信。
func (q *MailQueue) Enqueue(job *MailJob) {
	if job == nil || job.To == "" {
		return
	}
	select {
	case q.queue <- job:
	default:
		// 队列满：同步投递兜底，仍保证送达
		if err := q.deliver(job); err != nil {
			fmt.Printf("[WARN] 邮件队列已满且同步发送失败 to=%s err=%v\n", job.To, err)
		}
	}
}

// worker 后台消费协程
func (q *MailQueue) worker() {
	defer q.wg.Done()
	for {
		select {
		case <-q.stopCh:
			return
		case job := <-q.queue:
			q.deliver(job)
		}
	}
}

// deliver 投递单封邮件，失败按指数退避重试
func (q *MailQueue) deliver(job *MailJob) error {
	var lastErr error
	for attempt := 0; attempt <= q.maxRetry; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
		if err := q.sender(job); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	fmt.Printf("[WARN] 邮件发送失败（重试 %d 次后放弃）to=%s subject=%s err=%v\n",
		q.maxRetry, job.To, job.Subject, lastErr)
	return lastErr
}

// Stop 停止队列并等待在途任务结束
func (q *MailQueue) Stop() {
	q.once.Do(func() {
		close(q.stopCh)
	})
	q.wg.Wait()
}

// defaultMailSender 默认投递实现：使用当前生效的 SMTP 配置发送
func defaultMailSender(job *MailJob) error {
	smtpService := NewSMTPService()
	cfg, err := smtpService.GetConfig()
	if err != nil {
		return fmt.Errorf("邮件服务未配置: %v", err)
	}
	return smtpService.sendEmail(cfg, job.To, job.Subject, job.Body)
}
