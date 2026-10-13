package db

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/go-sql-driver/mysql"
)

// 业务事务重试：针对 MySQL 1213(死锁)/1205(锁等待超时) 做指数退避重试；
// 1062(唯一键冲突)仅在标注幂等(idempotent)的场景才重试。
// 放置于 db 包便于 service/dao 统一复用，避免各调用点重复实现退避逻辑。

const (
	mysqlErrDeadlock        = 1213
	mysqlErrLockWaitTimeout = 1205
	mysqlErrDuplicateEntry  = 1062
	mysqlErrUserSignal      = 1644 // SIGNAL SQLSTATE '45000' 抛出的自定义业务错误
)

type retryConfig struct {
	maxAttempts int
	baseDelay   time.Duration
	idempotent  bool
}

// RetryOption 调整重试行为。
type RetryOption func(*retryConfig)

// WithMaxAttempts 设置最大尝试次数（含首次），默认 3。
func WithMaxAttempts(n int) RetryOption {
	return func(c *retryConfig) {
		if n > 0 {
			c.maxAttempts = n
		}
	}
}

// WithIdempotent 允许对 1062 唯一键冲突重试（仅在调用幂等时启用）。
func WithIdempotent() RetryOption {
	return func(c *retryConfig) {
		c.idempotent = true
	}
}

// WithRetry 执行 fn；遇到可重试的 MySQL 错误时按 50ms、100ms…（含抖动）指数退避重试。
// ctx 用于取消/超时。默认最多尝试 3 次。不可重试的错误会立即返回。
func WithRetry(ctx context.Context, fn func() error, opts ...RetryOption) error {
	cfg := retryConfig{maxAttempts: 3, baseDelay: 50 * time.Millisecond, idempotent: false}
	for _, o := range opts {
		o(&cfg)
	}

	var lastErr error
	for attempt := 0; attempt < cfg.maxAttempts; attempt++ {
		if attempt > 0 {
			delay := cfg.baseDelay * time.Duration(1<<(attempt-1))
			jitter := time.Duration(rand.Int63n(int64(delay / 2)))
			delay += jitter
			select {
			case <-ctx.Done():
				if lastErr != nil {
					return lastErr
				}
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}
		if !isRetryable(lastErr, cfg.idempotent) {
			return lastErr
		}
	}
	return lastErr
}

// isRetryable 判断错误是否可重试。
func isRetryable(err error, idempotent bool) bool {
	var me *mysql.MySQLError
	if !errors.As(err, &me) {
		return false
	}
	switch me.Number {
	case mysqlErrDeadlock, mysqlErrLockWaitTimeout:
		return true
	case mysqlErrDuplicateEntry:
		return idempotent
	default:
		return false
	}
}

// BusinessErrorFromSQLError 将 SP 内 SIGNAL('45000') 抛出的业务错误转换为携带原始
// 信息的普通 error，便于上层返回明确业务错误而非 500。
func BusinessErrorFromSQLError(err error) (error, bool) {
	var me *mysql.MySQLError
	if errors.As(err, &me) && me.Number == mysqlErrUserSignal {
		return errors.New(me.Message), true
	}
	return nil, false
}

// IsDuplicateEntry 判断是否为唯一键冲突(1062)。
func IsDuplicateEntry(err error) bool {
	var me *mysql.MySQLError
	return errors.As(err, &me) && me.Number == mysqlErrDuplicateEntry
}
