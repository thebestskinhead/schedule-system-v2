package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"schedule-system-v2/backend/internal/update"
	"schedule-system-v2/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// UpdateHandler 热更新相关接口。
// 说明：真正的更新动作由独立进程 update.exe 完成（server.exe 正在运行时无法替换自身），
// 这里只负责「查询版本 / 检查更新 / 拉起 update.exe」。
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

// StartUpdate 启动热更新流程（系统管理员）。
// update.exe 会先停止本服务，因此这里先把响应返回前端，再由 update.exe 接管 8080 展示进度。
func (h *UpdateHandler) StartUpdate(c *gin.Context) {
	dir := installDirFn()
	updater := filepath.Join(dir, update.UpdaterExeName())
	if _, err := os.Stat(updater); err != nil {
		utils.Error(c, 400, "未找到更新程序 update.exe，请确认使用完整部署包安装")
		return
	}

	// 更新前先确认存在可用的同大版本更新，避免无谓停服
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

	cmd := exec.Command(updater)
	cmd.Dir = dir
	if err := cmd.Start(); err != nil {
		utils.Error(c, 500, "启动更新程序失败: "+err.Error())
		return
	}
	_ = cmd.Process.Release() // 分离进程：本服务被停止后 update.exe 继续运行

	utils.Success(c, gin.H{
		"message":    "更新程序已启动，服务即将停止。页面将自动跳转到更新进度页（http://localhost:8080）",
		"target":     info.Target.String(),
		"updaterPid": cmd.Process.Pid,
	})
}

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
