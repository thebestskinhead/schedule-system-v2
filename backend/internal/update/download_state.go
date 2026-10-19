package update

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// 下载阶段状态（updates/tmp/download-state.json）的 phase 取值。
// 该文件是下载阶段的唯一事实来源：update.exe 写入，主程序读取展示进度与就绪状态；
// 主程序重启后凭文件即可恢复状态（内存中只保存子进程句柄）。
const (
	DownloadPhaseChecking    = "checking"    // 检查 Release
	DownloadPhaseDownloading = "downloading" // 下载安装包（含镜像切换）
	DownloadPhaseVerifying   = "verifying"   // 校验安装包
	DownloadPhaseReady       = "ready"       // 已就绪，可执行安装
	DownloadPhaseFailed      = "failed"      // 失败或已取消
)

// DownloadStaleAfter 心跳停滞判定时长：状态文件 UpdatedAt 超过该时长仍未刷新、
// 且 phase 非终态时，视为下载进程已意外死亡（主程序据此判失败，防止进度假转圈）。
const DownloadStaleAfter = 30 * time.Second

// stampLayout 状态文件时间戳格式（与更新日志、结果文件保持一致）
const stampLayout = "2006-01-02 15:04:05"

// DownloadState 下载阶段的状态文件内容
type DownloadState struct {
	Phase          string `json:"phase"`          // 见 DownloadPhase* 常量
	Message        string `json:"message"`        // 当前阶段的可读描述
	Received       int64  `json:"received"`       // 已下载字节数
	Total          int64  `json:"total"`          // 总字节数（Release 声明值）
	Current        string `json:"current"`        // 当前版本
	Target         string `json:"target"`         // 目标版本
	AssetName      string `json:"assetName"`      // 安装包文件名
	AssetSize      int64  `json:"assetSize"`      // 安装包大小（字节）
	Checksum       string `json:"checksum"`       // 期望的 sha256（无校验文件时为空）
	Mirror         string `json:"mirror"`         // 本次下载实际使用的镜像（直连为空）
	LastGoodMirror string `json:"lastGoodMirror"` // 上次下载成功的镜像，下次优先使用
	Error          string `json:"error"`          // 失败原因
	PID            int    `json:"pid"`            // 下载进程 PID（识别无主下载）
	StartedAt      string `json:"startedAt"`
	UpdatedAt      string `json:"updatedAt"`  // 心跳刷新时间（判活依据）
	FinishedAt     string `json:"finishedAt"` // 进入终态的时间
}

// TerminalPhase 是否已进入终态（就绪/失败）
func (s *DownloadState) TerminalPhase() bool {
	return s.Phase == DownloadPhaseReady || s.Phase == DownloadPhaseFailed
}

// DownloadTmpDir 下载临时目录（安装包与状态文件都放在这里）
func DownloadTmpDir(dir string) string {
	return filepath.Join(dir, "updates", "tmp")
}

// DownloadStatePath 状态文件路径
func DownloadStatePath(dir string) string {
	return filepath.Join(DownloadTmpDir(dir), "download-state.json")
}

// DownloadCancelFlagPath 取消标志文件路径（主程序写入，下载中的 update.exe 轮询）
func DownloadCancelFlagPath(dir string) string {
	return filepath.Join(DownloadTmpDir(dir), "download-cancel")
}

// WriteDownloadState 原子写入状态文件（先写 .tmp 再 rename，防止读到半截 JSON）
func WriteDownloadState(dir string, s *DownloadState) error {
	path := DownloadStatePath(dir)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// ReadDownloadState 读取状态文件；不存在或解析失败时返回 nil
func ReadDownloadState(dir string) *DownloadState {
	data, err := os.ReadFile(DownloadStatePath(dir))
	if err != nil {
		return nil
	}
	var s DownloadState
	if err := json.Unmarshal(data, &s); err != nil {
		return nil
	}
	return &s
}

// DownloadStateStale 状态文件是否已停滞（进程意外死亡时主程序据此判失败）
func DownloadStateStale(s *DownloadState, now time.Time) bool {
	if s == nil || s.TerminalPhase() {
		return false
	}
	t, err := time.Parse(stampLayout, s.UpdatedAt)
	if err != nil {
		return true // 时间戳缺失或不可解析，按停滞处理
	}
	return now.Sub(t) > DownloadStaleAfter
}
