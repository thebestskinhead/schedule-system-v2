// update.exe —— 排班系统热更新程序
//
// 流程（对应用户要求）：
//  1. 启动后先终止 server 进程，等待 8080 端口释放；
//  2. server 停止后在 8080 端口启动进度监控页面（下载/安装进度、终止更新并恢复旧版本）；
//  3. 查询 GitHub Release，仅当存在「同大版本且版本号更高」的发布时才继续
//     （同大版本视为无数据库结构变更，可安全热更新；跨大版本需手动整包升级）；
//  4. 下载发布压缩包（可选 sha256 校验）→ 备份当前部署物 → 解压替换
//     （仅覆盖 server 可执行文件与前后端构建产物 dist/ dist-mobile/，
//     不动 configs/、migrations/、日志等运行数据）；
//  5. 释放 8080 端口，自动重启 server，健康检查确认存活后自身退出；
//  6. 任一环节失败或用户点击「终止更新」：从备份回滚并重启旧版本 server。
//
// 用法：
//
//	update.exe                完整更新流程（停服 → 进度页 → 下载安装 → 重启 → 自检退出）
//	update.exe --check        仅检查是否有可用更新（不停止服务、不安装），输出 JSON
//	update.exe --version v3.1.0   指定目标版本（与最新版不一致时仅告警）
//	update.exe --dir <path>   指定安装目录（默认为 update.exe 所在目录）
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"schedule-system-v2/backend/internal/update"
)

// ---------- 更新阶段 ----------

const (
	phaseIdle        = "idle"
	phaseStopping    = "stopping"
	phaseChecking    = "checking"
	phaseDownloading = "downloading"
	phaseBackingUp   = "backing-up"
	phaseInstalling  = "installing"
	phaseRestarting  = "restarting"
	phaseVerifying   = "verifying"
	phaseDone        = "done"
	phaseFailed      = "failed"
	phaseAborted     = "aborted"
	phaseRolledBack  = "rolled-back"
)

// doUpdate 返回码
const (
	codeSuccess   = 0 // 成功（新版本已重启并验证存活，或本就无需更新）
	codeFailed    = 1 // 失败（等待重试或自动恢复旧版本）
	codeTerminate = 2 // 已终止（abortFlow 内部已完成回滚与重启，无需再处理）
)

// State 进度页面展示的状态
type State struct {
	Phase          string   `json:"phase"`
	Message        string   `json:"message"`
	Progress       int      `json:"progress"` // 0-100
	Received       int64    `json:"received"`
	Total          int64    `json:"total"`
	CurrentVersion string   `json:"currentVersion"`
	TargetVersion  string   `json:"targetVersion"`
	CanAbort       bool     `json:"canAbort"`
	AutoAbortIn    int      `json:"autoAbortIn"` // 失败后自动恢复旧版本的倒计时（秒）
	Error          string   `json:"error"`
	Logs           []string `json:"logs"`
}

// Updater 热更新流程控制器
type Updater struct {
	mu        sync.Mutex
	state     State
	dir       string // 安装目录
	repo      string
	targetVer string
	allowPre  bool

	abortCh   chan struct{} // 用户请求中止时关闭
	abortOnce sync.Once
	retryCh   chan struct{} // 用户请求重试时发送

	backupDir  string
	httpServer *http.Server
	doneCh     chan struct{} // 流程结束（成功/失败/中止均算结束）
	doneOnce   sync.Once
	logFile    *os.File
}

func main() {
	var (
		checkMode = flag.Bool("check", false, "仅检查是否有可用更新（不停止服务、不安装），输出 JSON 后退出")
		repo      = flag.String("repo", update.DefaultRepo, "GitHub 仓库 owner/name")
		dirFlag   = flag.String("dir", "", "安装目录（默认为 update.exe 所在目录）")
		targetVer = flag.String("version", "", "指定目标版本（与最新版不一致时仅告警）")
		allowPre  = flag.Bool("allow-prerelease", false, "允许预发布版本")
		port      = flag.String("port", "8080", "进度页面端口")
	)
	flag.Parse()

	dir := resolveDir(*dirFlag)

	// ---- 仅检查模式：不停止服务、不动文件 ----
	if *checkMode {
		info, err := update.CheckUpdate(dir, *repo, update.PlatformAssetName(), *allowPre)
		if err != nil {
			fmt.Fprintf(os.Stderr, "检查更新失败: %v\n", err)
			os.Exit(1)
		}
		out, _ := json.MarshalIndent(info, "", "  ")
		fmt.Println(string(out))
		switch {
		case !info.HasUpdate:
			fmt.Println("当前已是最新版本。")
		case !info.SameMajor:
			fmt.Printf("存在新的大版本 %s，可能包含数据库结构变更，请手动下载整包升级。\n", info.Target)
			os.Exit(2)
		default:
			fmt.Printf("发现可热更新版本: %s -> %s\n", info.Current, info.Target)
		}
		return
	}

	u := &Updater{
		state: State{
			Phase:          phaseIdle,
			CurrentVersion: update.CurrentVersion(dir).String(),
			Logs:           []string{},
		},
		dir:       dir,
		repo:      *repo,
		targetVer: *targetVer,
		allowPre:  *allowPre,
		abortCh:   make(chan struct{}),
		retryCh:   make(chan struct{}, 1),
		doneCh:    make(chan struct{}),
	}
	u.openLog()
	defer u.closeLog()

	fmt.Printf("排班系统热更新程序\n安装目录: %s\n当前版本: %s\n\n", dir, u.state.CurrentVersion)

	// 1. 先终止 server（用户要求的流程顺序）
	u.setPhase(phaseStopping, "正在停止服务…")
	if _, err := update.StopServer(u.dir); err != nil {
		u.logf("停止服务: %v", err)
	} else {
		u.log("服务已停止")
	}

	// 2. server 停止后在 8080 启动进度监控页面
	if err := u.startPage(*port); err != nil {
		// 页面起不来不能阻断更新，退化为命令行模式
		fmt.Printf("警告: 进度页面启动失败（%v），将以命令行模式继续\n", err)
	} else {
		fmt.Printf("进度页面: http://localhost:%s （可在此监控进度或终止更新）\n", *port)
	}

	// 3. 执行更新流程（失败后可重试）
	go u.runUpdate()

	// 4. 等待流程结束
	<-u.doneCh

	// 给浏览器一点时间展示最终状态并完成自动刷新，然后退出自身进程
	time.Sleep(1500 * time.Millisecond)
	if u.httpServer != nil {
		_ = u.httpServer.Close()
	}
	fmt.Println("更新程序退出。")
}

// resolveDir 解析安装目录：优先 --dir，否则取 update.exe 所在目录
func resolveDir(dirFlag string) string {
	if dirFlag != "" {
		abs, err := filepath.Abs(dirFlag)
		if err == nil {
			return abs
		}
		return dirFlag
	}
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

// ---------- HTTP 进度服务 ----------

func (u *Updater) startPage(port string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(update.ProgressPageHTML))
	})
	mux.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(u.snapshot())
	})
	mux.HandleFunc("/api/abort", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		u.requestAbort()
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/api/retry", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		select {
		case u.retryCh <- struct{}{}:
		default:
		}
		w.WriteHeader(http.StatusOK)
	})

	u.httpServer = &http.Server{Addr: ":" + port, Handler: mux}
	errCh := make(chan error, 1)
	go func() { errCh <- u.httpServer.ListenAndServe() }()
	select {
	case err := <-errCh:
		return err
	case <-time.After(300 * time.Millisecond):
		return nil
	}
}

// stopPage 释放 8080 端口（重启 server 前必须调用，此后进度页面下线）
func (u *Updater) stopPage() {
	if u.httpServer != nil {
		_ = u.httpServer.Close()
		u.httpServer = nil
		time.Sleep(200 * time.Millisecond) // 等待监听句柄释放
	}
}

// ---------- 更新流程 ----------

// runUpdate 执行更新流程；失败后等待用户重试，超时则自动恢复旧版本
func (u *Updater) runUpdate() {
	defer u.finish()

	for attempt := 1; ; attempt++ {
		if attempt > 1 {
			u.logf("==== 第 %d 次尝试 ====", attempt)
			u.backupDir = ""
			u.mu.Lock()
			u.state.Received, u.state.Total = 0, 0
			u.mu.Unlock()
		}

		code := u.doUpdate(attempt)
		if code == codeTerminate {
			return // 中止流程内部已完成回滚 + 重启 + 结果落盘
		}
		if code == codeSuccess {
			return
		}

		// 失败：保持进度页面（8080 仍由 updater 持有），等待重试或自动恢复
		u.mu.Lock()
		u.state.AutoAbortIn = 60
		u.mu.Unlock()
		u.log("更新失败，60 秒后自动恢复旧版本（可点击「重试更新」立即重试）")

		select {
		case <-u.retryCh:
			u.mu.Lock()
			u.state.Phase = phaseIdle
			u.state.AutoAbortIn = 0
			u.state.Error = ""
			u.mu.Unlock()
			continue
		case <-time.After(60 * time.Second):
			u.log("超过 60 秒未重试，自动恢复旧版本")
		case <-u.abortCh:
			u.log("收到终止更新请求（失败状态）")
		}
		u.restartOldServer("已恢复更新前的版本")
		u.writeResult(phaseRolledBack, "更新失败，已自动恢复旧版本", nil)
		return
	}
}

// doUpdate 执行一次更新。
// 返回 codeSuccess：新版本已重启并验证存活（或本就无需更新）；
// 返回 codeFailed：失败，server 未重启（由 runUpdate 决定重试或恢复）；
// 返回 codeTerminate：用户中止，abortFlow 内部已完成回滚 + 重启。
func (u *Updater) doUpdate(attempt int) int {
	u.setPhase(phaseStopping, "正在停止服务…")
	if _, err := update.StopServer(u.dir); err != nil {
		u.logf("停止服务: %v", err)
	}

	// ---- 检查更新 ----
	u.setPhase(phaseChecking, "正在检查 GitHub Release…")
	info, err := update.CheckUpdate(u.dir, u.repo, update.PlatformAssetName(), u.allowPre)
	if err != nil {
		if u.abortRequested() {
			u.abortFlow()
			return codeTerminate
		}
		return u.fail(fmt.Sprintf("检查更新失败: %v", err))
	}
	u.setVersion(info.Current.String())
	if info.HasUpdate {
		u.logf("当前版本 %s，最新同大版本候选 %s", info.Current, info.Target)
	} else {
		u.logf("当前版本 %s，未发现可用的同大版本更新", info.Current)
	}

	if !info.HasUpdate {
		u.setPhase(phaseDone, fmt.Sprintf("当前已是最新版本（%s），无需更新", info.Current))
		u.log("没有可用的热更新版本，服务即将恢复")
		u.restartOldServer("服务已恢复运行")
		return codeSuccess
	}
	if !info.SameMajor {
		u.fail(fmt.Sprintf("发现新的大版本 %s（当前 %s）。大版本可能包含数据库结构变更，热更新已禁止，请手动下载整包升级。",
			info.Target, info.Current))
		return codeFailed
	}
	if u.targetVer != "" && info.Target.String() != u.targetVer {
		u.logf("提示: 指定版本 %s 与最新版 %s 不一致，按最新版更新", u.targetVer, info.Target)
	}
	u.setTargetVersion(info.Target.String())
	u.logf("发现可更新版本: %s -> %s（%s，%.2f MB）",
		info.Current, info.Target, info.AssetName, float64(info.AssetSize)/1048576)

	// ---- 下载 ----
	u.setPhase(phaseDownloading, "正在下载更新包…")
	pkgPath := filepath.Join(u.dir, "updates", "tmp", info.AssetName)
	if err := os.MkdirAll(filepath.Dir(pkgPath), 0755); err != nil {
		return u.fail(fmt.Sprintf("创建临时目录失败: %v", err))
	}
	err = update.DownloadFileCancelable(info.DownloadURL, pkgPath,
		func(received, total int64) {
			u.mu.Lock()
			u.state.Received, u.state.Total = received, total
			if total > 0 {
				u.state.Progress = int(received * 100 / total)
				u.state.Message = fmt.Sprintf("正在下载更新包… %d%%", u.state.Progress)
			}
			u.mu.Unlock()
		},
		func() bool { return u.abortRequested() })
	if err != nil {
		os.Remove(pkgPath)
		if u.abortRequested() {
			u.abortFlow()
			return codeTerminate
		}
		return u.fail(fmt.Sprintf("下载更新包失败: %v", err))
	}
	u.logf("下载完成: %s", pkgPath)

	// ---- 校验（Release 附带 .sha256 时） ----
	if info.ChecksumURL != "" {
		expect, err := update.DownloadChecksum(info.ChecksumURL)
		if err != nil {
			u.logf("警告: 获取校验文件失败（%v），跳过校验", err)
		} else if err := update.VerifyChecksum(pkgPath, expect); err != nil {
			os.Remove(pkgPath)
			return u.fail(fmt.Sprintf("安装包校验失败: %v", err))
		} else {
			u.log("sha256 校验通过")
		}
	}

	// ---- 备份 ----
	u.setPhase(phaseBackingUp, "正在备份当前版本…")
	backupDir, err := update.Backup(u.dir, info.Current.String()+"-"+update.NowLabel())
	if err != nil {
		return u.fail(fmt.Sprintf("备份当前版本失败: %v", err))
	}
	u.backupDir = backupDir
	u.logf("已备份到 %s", backupDir)

	// ---- 解压并替换 ----
	u.setPhase(phaseInstalling, "正在应用新版本…")
	stagingDir := filepath.Join(u.dir, "updates", "staging")
	_ = os.RemoveAll(stagingDir)
	if _, err := update.ExtractPackage(pkgPath, stagingDir); err != nil {
		if u.abortRequested() {
			u.abortFlow()
			return codeTerminate
		}
		return u.fail(fmt.Sprintf("解压更新包失败: %v", err))
	}
	if err := update.ApplyStaging(u.dir, stagingDir); err != nil {
		u.logf("应用新版本失败: %v", err)
		return u.failWithRollback(fmt.Sprintf("应用新版本失败: %v", err))
	}
	if err := update.WriteVersion(u.dir, info.Target); err != nil {
		u.logf("警告: 写入版本号失败: %v", err)
	}
	u.logf("已应用新版本 %s", info.Target)

	// ---- 重启 server ----
	u.setPhase(phaseRestarting, "更新完成，正在重启服务…")
	u.log("正在重启服务，页面即将自动刷新…")
	u.stopPage() // 必须先释放 8080，server 才能绑定

	if err := update.StartServer(u.dir); err != nil {
		return u.failAfterRestart(fmt.Sprintf("启动新版本失败: %v", err), info)
	}

	// ---- 健康检查 ----
	u.setPhase(phaseVerifying, "正在验证服务…")
	if err := update.WaitServerAlive(60 * time.Second); err != nil {
		return u.failAfterRestart(fmt.Sprintf("新版本健康检查失败: %v", err), info)
	}

	// ---- 成功 ----
	u.setPhase(phaseDone, fmt.Sprintf("更新成功！已升级到 %s，服务恢复正常", info.Target))
	u.logf("更新成功: %s -> %s，服务已验证存活", info.Current, info.Target)
	u.writeResult(phaseDone, fmt.Sprintf("更新成功，已升级到 %s", info.Target), info)
	_ = os.RemoveAll(filepath.Join(u.dir, "updates", "staging"))
	return codeSuccess
}

// failWithRollback 记录失败并立即回滚文件（不重启 server，页面保持可用可重试）
func (u *Updater) failWithRollback(msg string) int {
	u.log("正在回滚到更新前的版本…")
	if u.backupDir != "" {
		if err := update.Rollback(u.dir, u.backupDir); err != nil {
			u.logf("回滚失败: %v", err)
			u.fail(msg + "（且回滚失败，请手动处理）")
			return codeFailed
		}
		u.log("已回滚到更新前的版本")
	}
	return u.fail(msg)
}

// failAfterRestart 重启/验证阶段失败：此时页面已释放，只能回滚并尽力恢复服务，
// 结果写入 updates/last-result.json 供管理端查看。
func (u *Updater) failAfterRestart(msg string, info *update.UpdateInfo) int {
	u.logf("%s，正在回滚", msg)
	if u.backupDir != "" {
		if err := update.Rollback(u.dir, u.backupDir); err != nil {
			u.logf("回滚失败: %v", err)
			u.writeResult(phaseFailed, msg+"（回滚失败，请手动处理）", info)
			u.finish()
			os.Exit(3)
		}
	}
	if err := update.StartServer(u.dir); err != nil {
		u.logf("恢复旧版本失败: %v", err)
		u.writeResult(phaseFailed, msg+"（旧版本恢复失败，请手动启动 server.exe）", info)
		u.finish()
		os.Exit(3)
	}
	if err := update.WaitServerAlive(60 * time.Second); err != nil {
		u.writeResult(phaseFailed, msg+"（旧版本未就绪，请检查 server.err.log）", info)
		u.finish()
		os.Exit(3)
	}
	u.log("已回滚并恢复旧版本服务")
	u.writeResult(phaseFailed, msg+"（已自动回滚并恢复旧版本）", info)
	u.finish()
	os.Exit(3)
	return codeFailed // 不会执行到这里
}

// restartOldServer 恢复/启动旧版本 server（未发生替换时就是原版本）。
// 这是唯一负责「释放 8080 → 启动 server → 验证」的终点路径。
func (u *Updater) restartOldServer(msg string) {
	u.setPhase(phaseRestarting, msg+"，正在重启服务…")
	u.stopPage()

	if update.ServerRunning() {
		u.log("服务已在运行，等待就绪")
	} else if err := update.StartServer(u.dir); err != nil {
		u.fail(msg + "，但服务重启失败，请手动启动 server.exe")
		u.writeResult(phaseFailed, msg+"，服务重启失败", nil)
		return
	}
	if err := update.WaitServerAlive(60 * time.Second); err != nil {
		u.fail(msg + "，但服务未在预期时间内就绪，请检查 server.err.log")
		u.writeResult(phaseFailed, msg+"，服务未就绪", nil)
		return
	}
	u.setPhase(phaseDone, msg)
}

// abortFlow 用户中止：清理下载残留，若有备份则回滚，最后重启旧版本
func (u *Updater) abortFlow() {
	u.log("用户请求终止更新")
	u.setPhase(phaseAborted, "正在终止更新…")
	_ = os.RemoveAll(filepath.Join(u.dir, "updates", "tmp"))
	_ = os.RemoveAll(filepath.Join(u.dir, "updates", "staging"))
	if u.backupDir != "" {
		u.log("正在从备份恢复旧版本…")
		if err := update.Rollback(u.dir, u.backupDir); err != nil {
			u.logf("回滚失败: %v", err)
			u.restartOldServer("终止更新失败：无法恢复旧版本，请手动处理")
			u.writeResult(phaseFailed, "终止更新失败：无法恢复旧版本", nil)
			return
		}
		u.log("已回滚到更新前的版本")
	}
	u.restartOldServer("更新已终止，已恢复更新前的版本")
	u.writeResult(phaseRolledBack, "用户终止更新，已恢复旧版本", nil)
}

// ---------- 中止/重试 ----------

func (u *Updater) abortRequested() bool {
	select {
	case <-u.abortCh:
		return true
	default:
		return false
	}
}

// requestAbort 页面「终止更新并恢复旧版本」按钮
func (u *Updater) requestAbort() {
	u.mu.Lock()
	canAbort := u.state.CanAbort
	u.mu.Unlock()
	if !canAbort {
		return // 重启中/验证中/已完成 已无法中止
	}
	u.abortOnce.Do(func() {
		u.log("收到终止更新请求")
		close(u.abortCh)
	})
}

// ---------- 状态管理 ----------

func (u *Updater) setPhase(phase, msg string) {
	u.mu.Lock()
	u.state.Phase = phase
	u.state.Message = msg
	u.state.CanAbort = phase == phaseChecking || phase == phaseDownloading ||
		phase == phaseBackingUp || phase == phaseInstalling || phase == phaseFailed
	u.state.Error = ""
	if phase != phaseDownloading {
		u.state.Progress = phaseProgress(phase)
	}
	u.mu.Unlock()
	u.log(msg)
}

func phaseProgress(phase string) int {
	switch phase {
	case phaseStopping:
		return 5
	case phaseChecking:
		return 12
	case phaseBackingUp:
		return 70
	case phaseInstalling:
		return 85
	case phaseRestarting:
		return 95
	case phaseVerifying:
		return 98
	case phaseDone:
		return 100
	default:
		return 0
	}
}

func (u *Updater) setVersion(v string) {
	u.mu.Lock()
	u.state.CurrentVersion = v
	u.mu.Unlock()
}

func (u *Updater) setTargetVersion(v string) {
	u.mu.Lock()
	u.state.TargetVersion = v
	u.mu.Unlock()
}

// fail 记录失败状态（不重启 server，页面保持可用，等待重试或自动恢复）
func (u *Updater) fail(msg string) int {
	u.mu.Lock()
	u.state.Phase = phaseFailed
	u.state.Message = "更新失败"
	u.state.Error = msg
	u.state.Progress = 0
	u.mu.Unlock()
	u.log("失败: " + msg)
	return codeFailed
}

func (u *Updater) finish() { u.doneOnce.Do(func() { close(u.doneCh) }) }

func (u *Updater) snapshot() State {
	u.mu.Lock()
	defer u.mu.Unlock()
	s := u.state
	s.Logs = append([]string{}, u.state.Logs...)
	if len(s.Logs) > 200 {
		s.Logs = s.Logs[len(s.Logs)-200:]
	}
	return s
}

// ---------- 日志与结果 ----------

func (u *Updater) openLog() {
	logDir := filepath.Join(u.dir, "updates")
	_ = os.MkdirAll(logDir, 0755)
	f, err := os.OpenFile(filepath.Join(logDir, "update.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err == nil {
		u.logFile = f
	}
}

func (u *Updater) closeLog() {
	if u.logFile != nil {
		_ = u.logFile.Close()
	}
}

func (u *Updater) log(msg string) {
	line := time.Now().Format("2006-01-02 15:04:05") + " " + msg
	fmt.Println(line)
	u.mu.Lock()
	u.state.Logs = append(u.state.Logs, line)
	if len(u.state.Logs) > 500 {
		u.state.Logs = u.state.Logs[len(u.state.Logs)-500:]
	}
	u.mu.Unlock()
	if u.logFile != nil {
		_, _ = u.logFile.WriteString(line + "\n")
	}
}

func (u *Updater) logf(format string, args ...interface{}) {
	u.log(fmt.Sprintf(format, args...))
}

// LastResult 最近一次更新的结果（写入 updates/last-result.json，供管理端展示）
type LastResult struct {
	Phase   string `json:"phase"` // done / failed / rolled-back
	Message string `json:"message"`
	From    string `json:"from"`
	To      string `json:"to"`
	At      string `json:"at"`
}

// writeResult 落盘最近一次更新结果
func (u *Updater) writeResult(phase, message string, info *update.UpdateInfo) {
	r := LastResult{Phase: phase, Message: message, At: time.Now().Format("2006-01-02 15:04:05")}
	if info != nil {
		r.From, r.To = info.Current.String(), info.Target.String()
	} else {
		u.mu.Lock()
		r.From, r.To = u.state.CurrentVersion, u.state.TargetVersion
		u.mu.Unlock()
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Join(u.dir, "updates"), 0755)
	_ = os.WriteFile(filepath.Join(u.dir, "updates", "last-result.json"), data, 0644)
}
