package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"schedule-system-v2/backend/internal/config"
	"schedule-system-v2/backend/internal/update"
	"schedule-system-v2/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// UpdateHandler 热更新相关接口。
// 说明：真正的更新动作由独立进程 update.exe 分两个阶段完成（先 download 下载校验、
// 再 install 停机安装），server.exe 正在运行时无法替换自身；这里只负责
// 「查询版本 / 检查更新 / 拉起 update.exe / 托管下载任务 / 镜像配置」。
type UpdateHandler struct{}

func NewUpdateHandler() *UpdateHandler {
	return &UpdateHandler{}
}

// installDir 返回安装目录（server.exe 所在目录，即部署根目录）
func installDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

// installDirFn 安装目录解析函数（测试可覆盖为临时目录）
var installDirFn = installDir

// GetVersion 获取当前版本与运行环境（公开接口，前端登录页也可展示版本号）
func (h *UpdateHandler) GetVersion(c *gin.Context) {
	dir := installDirFn()
	utils.Success(c, gin.H{
		"version":        update.CurrentVersion(dir).String(),
		"goVersion":      runtime.Version(),
		"platform":       runtime.GOOS + "/" + runtime.GOARCH,
		"updaterPresent": update.UpdaterPresent(dir),
	})
}

// checkCache 检查结果缓存，避免频繁请求 GitHub API（未认证限流 60 次/小时）
var (
	checkCacheMu   sync.Mutex
	checkCacheInfo *update.UpdateInfo
	checkCacheErr  error
	checkCacheAt   time.Time
)

const checkCacheTTL = 5 * time.Minute

// cachedCheckUpdate 带 5 分钟缓存的更新检查
func cachedCheckUpdate() (*update.UpdateInfo, error) {
	checkCacheMu.Lock()
	defer checkCacheMu.Unlock()
	if checkCacheInfo != nil && time.Since(checkCacheAt) < checkCacheTTL {
		return checkCacheInfo, checkCacheErr
	}
	info, err := update.CheckUpdate(installDirFn(), update.DefaultRepo, update.PlatformAssetName(), false)
	checkCacheInfo, checkCacheErr, checkCacheAt = info, err, time.Now()
	return info, err
}

// CheckUpdate 检查是否有可热更新的版本（系统管理员）
func (h *UpdateHandler) CheckUpdate(c *gin.Context) {
	info, err := cachedCheckUpdate()
	if err != nil {
		utils.Error(c, 500, "检查更新失败: "+err.Error())
		return
	}
	utils.Success(c, gin.H{
		"current":      info.Current.String(),
		"target":       info.Target.String(),
		"hasUpdate":    info.HasUpdate,
		"sameMajor":    info.SameMajor,
		"crossMajor":   info.CrossMajor,
		"assetName":    info.AssetName,
		"assetSize":    info.AssetSize,
		"downloadUrl":  info.DownloadURL,
		"releaseNotes": info.ReleaseNotes,
	})
}

// ---------- 下载任务托管（内存只存子进程句柄，状态以文件为准） ----------

// downloadTask 当前下载任务：server 重启后内存态丢失，
// 就绪状态由 updates/tmp/download-state.json 重建，故此处只存进程句柄。
type downloadTask struct {
	mu      sync.Mutex
	cmd     *exec.Cmd
	done    chan struct{} // cmd.Wait() 返回后关闭
	running bool
}

var download downloadTask

// StartDownload 启动下载阶段（系统管理员）。
// 以子进程方式拉起 update.exe download：服务不停止、8080 不换手，
// 下载与校验完成后进入就绪态，可再执行安装。
func (h *UpdateHandler) StartDownload(c *gin.Context) {
	dir := installDirFn()
	updater := filepath.Join(dir, update.UpdaterExeName())
	if _, err := os.Stat(updater); err != nil {
		utils.Error(c, 400, "未找到更新程序 update.exe，请确认使用完整部署包安装")
		return
	}

	// 更新前先确认存在可用的同大版本更新，避免无谓下载
	info, err := cachedCheckUpdate()
	if err != nil {
		utils.Error(c, 500, "检查更新失败: "+err.Error())
		return
	}
	if !info.HasUpdate {
		utils.Error(c, 400, "当前已是最新版本")
		return
	}
	if !info.SameMajor {
		utils.Error(c, 400, "目标版本 "+info.Target.String()+" 与当前版本跨大版本，可能包含数据库结构变更，请手动下载整包升级")
		return
	}

	// 单航班：已有下载任务在进行时拒绝重复启动
	download.mu.Lock()
	if download.running {
		download.mu.Unlock()
		utils.Error(c, 409, "已有下载任务正在进行，请等待完成或取消后重试")
		return
	}
	download.mu.Unlock()

	args := []string{"download", "--dir", dir}
	if mirrors := enabledMirrors(); len(mirrors) > 0 {
		args = append(args, "--mirrors", strings.Join(mirrors, ","))
	}
	cmd := exec.Command(updater, args...)
	cmd.Dir = dir
	if err := cmd.Start(); err != nil {
		utils.Error(c, 500, "启动更新程序失败: "+err.Error())
		return
	}

	done := make(chan struct{})
	download.mu.Lock()
	download.cmd, download.done, download.running = cmd, done, true
	download.mu.Unlock()

	// 后台等待退出：进程退出后若状态文件仍无终态，标记为失败
	go func() {
		_ = cmd.Wait()
		close(done)
		download.mu.Lock()
		if download.cmd == cmd {
			download.running = false
		}
		download.mu.Unlock()
		if st := update.ReadDownloadState(dir); st == nil || !st.TerminalPhase() {
			msg := "下载进程意外退出"
			if fileExists(update.DownloadCancelFlagPath(dir)) {
				msg = "下载已取消"
			}
			now := nowStamp()
			_ = update.WriteDownloadState(dir, &update.DownloadState{
				Phase: update.DownloadPhaseFailed, Message: "下载失败", Error: msg,
				Current: stateField(st, func(s *update.DownloadState) string { return s.Current }),
				Target:  stateField(st, func(s *update.DownloadState) string { return s.Target }),
				PID:     os.Getpid(), UpdatedAt: now, FinishedAt: now,
			})
		}
		// 取消标志寿命仅限于本次取消
		_ = os.Remove(update.DownloadCancelFlagPath(dir))
	}()

	utils.Success(c, gin.H{
		"message":    "下载任务已启动，可在本页查看进度",
		"target":     info.Target.String(),
		"updaterPid": cmd.Process.Pid,
	})
}

// GetDownloadProgress 查询下载/就绪状态（系统管理员）。
// 以状态文件为准：进程已退出且文件无终态、或心跳停滞超过 30 秒时判为失败。
func (h *UpdateHandler) GetDownloadProgress(c *gin.Context) {
	dir := installDirFn()
	st := update.ReadDownloadState(dir)

	download.mu.Lock()
	cmd, done, running := download.cmd, download.done, download.running
	download.mu.Unlock()

	if st == nil {
		utils.Success(c, gin.H{"phase": "idle", "running": false})
		return
	}

	// 心跳停滞 / 进程已退出但无终态 → 下载进程已意外死亡：杀掉僵尸进程并判失败
	if !st.TerminalPhase() && (!running || update.DownloadStateStale(st, time.Now())) {
		if running && cmd != nil && cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		waitDone(done)
		download.mu.Lock()
		if download.cmd == cmd {
			download.running = false
		}
		download.mu.Unlock()
		now := nowStamp()
		st.Phase, st.Message, st.Error = update.DownloadPhaseFailed, "下载失败", "下载进程无响应，已终止"
		if st.FinishedAt == "" {
			st.FinishedAt = now
		}
		_ = update.WriteDownloadState(dir, st)
		running = false
	}

	data := gin.H{
		"phase":          st.Phase,
		"message":        st.Message,
		"error":          st.Error,
		"received":       st.Received,
		"total":          st.Total,
		"current":        st.Current,
		"target":         st.Target,
		"assetName":      st.AssetName,
		"assetSize":      st.AssetSize,
		"mirror":         st.Mirror,
		"lastGoodMirror": st.LastGoodMirror,
		"startedAt":      st.StartedAt,
		"updatedAt":      st.UpdatedAt,
		"finishedAt":     st.FinishedAt,
		"running":        running && !st.TerminalPhase(),
	}
	if st.Total > 0 {
		data["progress"] = int(st.Received * 100 / st.Total)
	} else {
		data["progress"] = 0
	}
	utils.Success(c, data)
}

// CancelDownload 取消进行中的下载（系统管理员）。
// 下载阶段未触碰任何部署物，取消零风险：写取消标志 → 杀进程 → 标记失败。
func (h *UpdateHandler) CancelDownload(c *gin.Context) {
	dir := installDirFn()

	download.mu.Lock()
	cmd, done, running := download.cmd, download.done, download.running
	if running {
		download.running = false
	}
	download.mu.Unlock()

	if !running || cmd == nil || cmd.Process == nil {
		utils.Error(c, 400, "当前没有进行中的下载任务")
		return
	}

	// 取消标志：update.exe 轮询到后自行清理临时文件并退出
	_ = os.MkdirAll(update.DownloadTmpDir(dir), 0755)
	_ = os.WriteFile(update.DownloadCancelFlagPath(dir), []byte("cancel"), 0644)
	_ = cmd.Process.Kill()
	waitDone(done)

	// 兜底：进程退出后若状态仍无终态（被强杀来不及自报），标记为已取消
	if st := update.ReadDownloadState(dir); st == nil || !st.TerminalPhase() {
		now := nowStamp()
		_ = update.WriteDownloadState(dir, &update.DownloadState{
			Phase: update.DownloadPhaseFailed, Message: "下载失败", Error: "下载已取消",
			Current: stateField(st, func(s *update.DownloadState) string { return s.Current }),
			Target:  stateField(st, func(s *update.DownloadState) string { return s.Target }),
			UpdatedAt: now, FinishedAt: now,
		})
	}
	_ = os.Remove(update.DownloadCancelFlagPath(dir))
	utils.Success(c, gin.H{"message": "下载已取消"})
}

// StartInstall 启动安装阶段（系统管理员）。
// 要求下载已就绪；update.exe install 以分离进程启动（随后本服务会被它停止）。
func (h *UpdateHandler) StartInstall(c *gin.Context) {
	dir := installDirFn()
	updater := filepath.Join(dir, update.UpdaterExeName())
	if _, err := os.Stat(updater); err != nil {
		utils.Error(c, 400, "未找到更新程序 update.exe，请确认使用完整部署包安装")
		return
	}

	st := update.ReadDownloadState(dir)
	if st == nil || st.Phase != update.DownloadPhaseReady {
		utils.Error(c, 409, "更新包尚未就绪，请先执行下载并等待校验完成")
		return
	}
	pkgPath := filepath.Join(update.DownloadTmpDir(dir), st.AssetName)
	if st.AssetName == "" || !fileExists(pkgPath) {
		utils.Error(c, 409, "更新包文件缺失，请重新执行下载")
		return
	}

	cmd := exec.Command(updater, "install", "--dir", dir)
	cmd.Dir = dir
	if err := cmd.Start(); err != nil {
		utils.Error(c, 500, "启动更新程序失败: "+err.Error())
		return
	}
	_ = cmd.Process.Release() // 分离进程：本服务将被 update.exe 停止，它必须继续运行

	utils.Success(c, gin.H{
		"message":    "安装程序已启动，服务即将停止。页面将自动跳转到更新进度页（http://localhost:8080）",
		"target":     st.Target,
		"updaterPid": cmd.Process.Pid,
	})
}

// ---------- 镜像加速配置 ----------

// GetMirrors 获取镜像加速地址列表（系统管理员）
func (h *UpdateHandler) GetMirrors(c *gin.Context) {
	utils.Success(c, gin.H{"mirrors": enabledMirrors()})
}

// SaveMirrors 保存镜像加速地址列表（系统管理员）。
// 列表有序，下载时按顺序尝试；置空表示全部直连。
func (h *UpdateHandler) SaveMirrors(c *gin.Context) {
	var req struct {
		Mirrors []string `json:"mirrors"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}
	mirrors, err := normalizeMirrors(req.Mirrors)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	cfg, err := config.LoadConfig(config.ConfigFilePath)
	if err != nil {
		utils.Error(c, 500, "加载配置失败")
		return
	}
	cfg.Update.Mirrors = mirrors
	if err := config.SaveConfig(cfg); err != nil {
		utils.Error(c, 500, "保存配置失败")
		return
	}
	utils.Success(c, gin.H{"mirrors": mirrors})
}

// enabledMirrors 当前启用的镜像列表（来自 config.yaml 的 update.mirrors）
func enabledMirrors() []string {
	return config.GetConfig().Update.Mirrors
}

// normalizeMirrors 清洗镜像列表：去空白、去重、校验 http/https，保序
func normalizeMirrors(in []string) ([]string, error) {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, m := range in {
		m = strings.TrimSpace(m)
		if m == "" {
			continue
		}
		if !strings.HasPrefix(m, "http://") && !strings.HasPrefix(m, "https://") {
			return nil, fmt.Errorf("镜像地址必须以 http:// 或 https:// 开头: %s", m)
		}
		if strings.ContainsAny(m, " \t\r\n") {
			return nil, fmt.Errorf("镜像地址不能包含空白字符: %s", m)
		}
		if seen[m] {
			continue
		}
		seen[m] = true
		out = append(out, m)
	}
	return out, nil
}

// ---------- 结果与日志 ----------

// GetUpdateStatus 查看最近一次更新的结果（系统管理员）
func (h *UpdateHandler) GetUpdateStatus(c *gin.Context) {
	data, err := os.ReadFile(filepath.Join(installDirFn(), "updates", "last-result.json"))
	if err != nil {
		utils.Success(c, gin.H{"hasResult": false})
		return
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		utils.Error(c, 500, "解析更新结果失败: "+err.Error())
		return
	}
	result["hasResult"] = true
	utils.Success(c, result)
}

// UpdateLogTail 读取更新日志末尾（系统管理员，便于排查失败原因）
func (h *UpdateHandler) UpdateLogTail(c *gin.Context) {
	data, err := os.ReadFile(filepath.Join(installDirFn(), "updates", "update.log"))
	if err != nil {
		utils.Success(c, gin.H{"log": ""})
		return
	}
	log := string(data)
	if len(log) > 8*1024 {
		log = log[len(log)-8*1024:]
	}
	utils.Success(c, gin.H{"log": log})
}

// HealthCheck 探活（供 update.exe 健康检查使用，公开接口）
func (h *UpdateHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "ok"})
}

// ---------- 内部工具 ----------

func nowStamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// waitDone 等待子进程退出（带超时，防止卡死请求）
func waitDone(done chan struct{}) {
	if done == nil {
		return
	}
	select {
	case <-done:
	case <-time.After(10 * time.Second):
	}
}

// stateField 从状态文件安全取字段（st 为 nil 时返回空串）
func stateField(st *update.DownloadState, get func(*update.DownloadState) string) string {
	if st == nil {
		return ""
	}
	return get(st)
}
