//go:build windows

package update

import (
	"path/filepath"

	"golang.org/x/sys/windows"
)

// diskFreeBytes 返回 dir 所在卷的剩余可写字节数（Windows 实现）
func diskFreeBytes(dir string) (int64, error) {
	dirPtr, err := windows.UTF16PtrFromString(filepath.Clean(dir))
	if err != nil {
		return 0, err
	}
	var freeBytesAvailable uint64
	if err := windows.GetDiskFreeSpaceEx(dirPtr, &freeBytesAvailable, nil, nil); err != nil {
		return 0, err
	}
	return int64(freeBytesAvailable), nil
}
