package update

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Version 语义化版本号（仅比较 major.minor.patch，忽略预发布/构建后缀）
type Version struct {
	Major int
	Minor int
	Patch int
	Raw   string // 原始字符串，如 "v3.0.1"
}

// versionRe 匹配 vX.Y.Z（允许 -suffix / +build 后缀）
var versionRe = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)`)

// ParseVersion 解析版本字符串；无法解析时返回错误
func ParseVersion(s string) (Version, error) {
	s = strings.TrimSpace(s)
	m := versionRe.FindStringSubmatch(s)
	if m == nil {
		return Version{}, fmt.Errorf("无效的版本号: %q", s)
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	patch, _ := strconv.Atoi(m[3])
	return Version{Major: major, Minor: minor, Patch: patch, Raw: s}, nil
}

// String 返回 vX.Y.Z 形式
func (v Version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Compare 比较两个版本：v > other 返回 1，相等返回 0，小于返回 -1
func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		return cmpInt(v.Major, other.Major)
	}
	if v.Minor != other.Minor {
		return cmpInt(v.Minor, other.Minor)
	}
	if v.Patch != other.Patch {
		return cmpInt(v.Patch, other.Patch)
	}
	return 0
}

// SameMajor 是否与 other 同大版本（同大版本视为无数据库结构变更，可热更新）
func (v Version) SameMajor(other Version) bool {
	return v.Major == other.Major
}

func cmpInt(a, b int) int {
	switch {
	case a > b:
		return 1
	case a < b:
		return -1
	default:
		return 0
	}
}

// VersionFileName 安装目录下的版本标识文件
const VersionFileName = "version.txt"

// VersionFilePath 返回安装目录中版本文件的完整路径
func VersionFilePath(dir string) string {
	return filepath.Join(dir, VersionFileName)
}

// CurrentVersion 读取安装目录的当前版本。
// version.txt 不存在或无法解析时返回 v0.0.0（视为任意 release 均为更新）。
func CurrentVersion(dir string) Version {
	data, err := os.ReadFile(VersionFilePath(dir))
	if err != nil {
		return Version{Raw: "v0.0.0"}
	}
	v, err := ParseVersion(string(data))
	if err != nil {
		return Version{Raw: "v0.0.0"}
	}
	return v
}

// WriteVersion 将新版本号写入安装目录的 version.txt
func WriteVersion(dir string, v Version) error {
	return os.WriteFile(VersionFilePath(dir), []byte(v.String()+"\n"), 0644)
}
