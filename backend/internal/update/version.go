package update

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Version 语义化版本号（比较遵循 semver：major.minor.patch 相同时，
// 正式版高于预发布版，如 v3.1.0 > v3.1.0-dev+abc123）
type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string // 预发布/构建后缀（不含首位 '-'/'+'），如 "dev+abc123"；为空表示正式版
	Raw        string // 原始字符串，如 "v3.0.1"
}

// versionRe 匹配 vX.Y.Z（可选 -/+ 后缀，如 -dev+abc123）
var versionRe = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:[-+]([0-9A-Za-z.+-]*))?$`)

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
	return Version{Major: major, Minor: minor, Patch: patch, Prerelease: m[4], Raw: s}, nil
}

// String 返回规范形式；带预发布后缀时一并输出（如 v3.1.0-dev+abc123）
func (v Version) String() string {
	s := fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Prerelease != "" {
		s += "-" + v.Prerelease
	}
	return s
}

// Compare 比较两个版本：v > other 返回 1，相等返回 0，小于返回 -1。
// 先比 major.minor.patch；相同时正式版（无后缀）高于预发布版，
// 两个预发布版按后缀逐段比较（数字段按数值、其余按字典序）。
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
	return comparePrerelease(v.Prerelease, other.Prerelease)
}

// comparePrerelease 比较预发布后缀：空后缀（正式版）最大
func comparePrerelease(a, b string) int {
	if a == b {
		return 0
	}
	if a == "" {
		return 1
	}
	if b == "" {
		return -1
	}
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")
	for i := 0; i < len(as) && i < len(bs); i++ {
		if c := compareIdentifier(as[i], bs[i]); c != 0 {
			return c
		}
	}
	return cmpInt(len(as), len(bs))
}

// compareIdentifier 单个后缀标识符比较：纯数字按数值，否则按字典序
func compareIdentifier(a, b string) int {
	na, ea := strconv.Atoi(a)
	nb, eb := strconv.Atoi(b)
	if ea == nil && eb == nil {
		return cmpInt(na, nb)
	}
	return strings.Compare(a, b)
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
