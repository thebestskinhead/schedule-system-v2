//go:build !windows

package update

import (
	"syscall"
)

// diskFreeBytes 返回 dir 所在卷的剩余可写字节数（Linux/macOS 实现）
func diskFreeBytes(dir string) (int64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(dir, &stat); err != nil {
		return 0, err
	}
	return int64(stat.Bavail) * int64(stat.Bsize), nil
}
