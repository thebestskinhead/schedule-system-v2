package db

import (
	"context"
	"errors"
	"testing"

	"github.com/go-sql-driver/mysql"
)

func TestIsRetryable(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		idempotent bool
		want       bool
	}{
		{"死锁 1213 可重试", &mysql.MySQLError{Number: 1213, Message: "Deadlock"}, false, true},
		{"锁等待超时 1205 可重试", &mysql.MySQLError{Number: 1205, Message: "Lock wait timeout"}, false, true},
		{"唯一键冲突 1062 非幂等不重试", &mysql.MySQLError{Number: 1062, Message: "Duplicate"}, false, false},
		{"唯一键冲突 1062 幂等可重试", &mysql.MySQLError{Number: 1062, Message: "Duplicate"}, true, true},
		{"业务 SIGNAL 1644 不重试", &mysql.MySQLError{Number: 1644, Message: "业务错误"}, false, false},
		{"普通 error 不重试", errors.New("boom"), false, false},
		{"nil 不重试", nil, false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isRetryable(c.err, c.idempotent); got != c.want {
				t.Fatalf("isRetryable(%v, %v) = %v, want %v", c.err, c.idempotent, got, c.want)
			}
		})
	}
}

func TestBusinessErrorFromSQLError(t *testing.T) {
	biz, ok := BusinessErrorFromSQLError(&mysql.MySQLError{Number: 1644, Message: "该周已有已确认或完成的排班，禁止重排"})
	if !ok {
		t.Fatal("期望识别 1644 为业务错误")
	}
	if biz == nil || biz.Error() != "该周已有已确认或完成的排班，禁止重排" {
		t.Fatalf("业务错误消息未保留原文: %v", biz)
	}

	if _, ok := BusinessErrorFromSQLError(&mysql.MySQLError{Number: 1062, Message: "duplicate"}); ok {
		t.Fatal("1062 不应被识别为业务错误")
	}
	if _, ok := BusinessErrorFromSQLError(errors.New("plain")); ok {
		t.Fatal("普通 error 不应被识别为业务错误")
	}
}

func TestIsDuplicateEntry(t *testing.T) {
	if !IsDuplicateEntry(&mysql.MySQLError{Number: 1062, Message: "dup"}) {
		t.Fatal("1062 应识别为唯一键冲突")
	}
	if IsDuplicateEntry(&mysql.MySQLError{Number: 1644, Message: "sig"}) {
		t.Fatal("1644 不应识别为唯一键冲突")
	}
}

func TestWithRetryStopsOnNonRetryable(t *testing.T) {
	calls := 0
	err := WithRetry(context.Background(), func() error {
		calls++
		return &mysql.MySQLError{Number: 1062, Message: "dup"}
	})
	if err == nil {
		t.Fatal("期望返回错误")
	}
	if calls != 1 {
		t.Fatalf("非可重试错误应只尝试 1 次，实际 %d", calls)
	}
}

func TestWithRetryRetriesDeadlock(t *testing.T) {
	calls := 0
	err := WithRetry(context.Background(), func() error {
		calls++
		if calls < 3 {
			return &mysql.MySQLError{Number: 1213, Message: "Deadlock"}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("最终应成功，实际 %v", err)
	}
	if calls != 3 {
		t.Fatalf("期望尝试 3 次，实际 %d", calls)
	}
}
