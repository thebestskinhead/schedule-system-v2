// update.exe —— 排班系统热更新程序
//
// 更新分两个阶段，总生命周期由主程序（系统管理页）管理：
//
//	update.exe download [--dir <path>] [--repo owner/name] [--mirrors u1,u2]
//	    阶段一「下载」：检查 Release → 下载安装包（镜像优先、直连兜底）→ sha256 校验。
//	    全程不停止服务、不占用 8080；进度与就绪状态写入 updates/tmp/download-state.json
//	    （原子写 + 2 秒心跳）。退出码 0 = 已就绪；非 0 = 失败。
//
//	update.exe install [--dir <path>]
//	    阶段二「安装」：读取已就绪的下载状态 → 预检（包在、磁盘、备份目录）→
//	    停止服务 → 在 8080 展示安装进度（安装不可终止，失败可重试或自动恢复）→
//	    校验安装包 → 备份 → 解压替换 → 写版本号 → 重启服务 → 健康检查 →
//	    清理临时文件后退出。安装全程不访问外网。
//
//	update.exe --check [--dir <path>] [--repo owner/name] [--version vX.Y.Z]
//	    仅检查是否有可用更新（不停止服务、不安装），输出 JSON 后退出
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"schedule-system-v2/backend/internal/update"
)

// ---------- 安装阶段（进度页面展示） ----------

const (
	phaseVerifyPkg  = "verifying-pkg" // 校验安装包
	phaseStopping   = "stopping"      // 停止服务
	phaseBackingUp  = "backing-up"    // 备份当前版本
	phaseInstalling = "installing"    // 应用新版本
	phaseRestarting = "restarting"    // 重启服务
	phaseVerifying  = "verifying"     // 验证服务
	phaseDone       = "done"          // 完成
	phaseFailed     = "failed"        // 失败（等待重试或自动恢复）
	phaseRolledBack = "rolled-back"   // 已恢复更新前的版本
)

// State 安装进度页面展示的状态
type State struct {
	Phase          string   `json:"phase"`
	Message        string   `json:"message"`
	Progress       int      `json:"progress"` // 0-100
	Received       int64    `json:"received"`
	Total          int64    `json:"total"`
	CurrentVersion string   `json:"currentVersion"`
	TargetVersion  string   `json:"targetVersion"`
	AutoAbortIn    int      `json:"autoAbortIn"` // 失败后自动恢复旧版本的倒计时（秒）
	Error          string   `json:"error"`
	Logs           []string `json:"logs"`
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
	}
	switch args[0] {
	case "download":
		runDownload(args[1:])
	case "install":
		runInstall(args[1:])
	default:
		runCheck(args)
	}
}

func usage() {
	fmt.Println(`用法：
  update.exe download [--dir <path>] [--repo owner/name] [--mirrors u1,u2]   下载并校验更新包（不停止服务）
  update.exe install  [--dir <path>] [--port 8080]                          安装已就绪的更新包（期间停止服务）
  update.exe --check  [--dir <path>] [--repo owner/name] [--version vX.Y.Z] 仅检查是否有可用更新`)
	os.Exit(2)
}

// runCheck 仅检查模式：不停止服务、不动文件，输出 JSON 后退出
func runCheck(args []string) {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	repo := fs.String("repo", update.DefaultRepo, "GitHub 仓库 owner/name")
	dirFlag := fs.String("dir", "", "安装目录（默认为 update.exe 所在目录）")
	targetVer := fs.String("version", "", "指定目标版本（与最新版不一致时仅告警）")
	allowPre := fs.Bool("allow-prerelease", false, "允许预发布版本")
	// 兼容 `update.exe --check` 显式调用：本模式即默认分支，该参数无实际作用
	_ = fs.Bool("check", false, "仅检查是否有可用更新")
	_ = fs.Parse(args)

	dir := resolveDir(*dirFlag)
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
		if *targetVer != "" && *targetVer != info.Target.String() {
			fmt.Printf("提示: 指定版本 %s 与最新版 %s 不一致，按最新版更新\n", *targetVer, info.Target)
		}
		fmt.Printf("发现可热更新版本: %s -> %s\n", info.Current, info.Target)
	}
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

// ---------- 小工具 ----------

func nowStamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func fileSize(p string) int64 {
	info, err := os.Stat(p)
	if err != nil {
		return 0
	}
	return info.Size()
}

// splitMirrors 拆分逗号分隔的镜像列表（去空白、去空项，保序）
func splitMirrors(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	var out []string
	for _, m := range strings.Split(s, ",") {
		if m = strings.TrimSpace(m); m != "" {
			out = append(out, m)
		}
	}
	return out
}

// prioritizeMirror 把上次成功的镜像移到列表最前（省去逐个试错）
func prioritizeMirror(mirrors []string, best string) []string {
	if best == "" {
		return mirrors
	}
	out := []string{best}
	for _, m := range mirrors {
		if m != best {
			out = append(out, m)
		}
	}
	return out
}

// ---------- 阶段一：下载 ----------

// Downloader 下载阶段控制器：进度写入状态文件，由主程序轮询展示。
// 不停止服务、不占用 8080——下载期间部署环境全程正常提供服务。
type Downloader struct {
	mu       sync.Mutex
	dir      string
	repo     string
	allowPre bool
	mirrors  []string
	state    update.DownloadState
	stopHb   chan struct{}
	hbOnce   sync.Once
}

func runDownload(args []string) {
	fs := flag.NewFlagSet("download", flag.ExitOnError)
	dirFlag := fs.String("dir", "", "安装目录（默认为 update.exe 所在目录）")
	repo := fs.String("repo", update.DefaultRepo, "GitHub 仓库 owner/name")
	mirrorsFlag := fs.String("mirrors", "", "镜像加速地址列表，逗号分隔（按顺序尝试，全部失败后回退直连）")
	allowPre := fs.Bool("allow-prerelease", false, "允许预发布版本")
	_ = fs.Parse(args)

	dir := resolveDir(*dirFlag)
	d := &Downloader{
		dir:      dir,
		repo:     *repo,
		allowPre: *allowPre,
		mirrors:  splitMirrors(*mirrorsFlag),
		state: update.DownloadState{
			Phase:     update.DownloadPhaseChecking,
			Current:   update.CurrentVersion(dir).String(),
			PID:       os.Getpid(),
			StartedAt: nowStamp(),
			UpdatedAt: nowStamp(),
		},
	}
	os.Exit(d.run())
}

// run 执行下载流程：检查 →（镜像优先）下载 → 校验 → 就绪。
// 返回进程退出码：0 = 已就绪；1 = 失败。
func (d *Downloader) run() int {
	// 上次成功的镜像优先（须在清理旧状态之前读取）
	if st := update.ReadDownloadState(d.dir); st != nil && st.LastGoodMirror != "" {
		d.mirrors = prioritizeMirror(d.mirrors, st.LastGoodMirror)
	}
	// 清理上一次下载的残留：旧状态、取消标志、未完成或已过期的安装包
	_ = os.RemoveAll(update.DownloadTmpDir(d.dir))
	_ = os.MkdirAll(update.DownloadTmpDir(d.dir), 0755)

	fmt.Printf("排班系统热更新程序（下载阶段）\n安装目录: %s\n当前版本: %s\n\n", d.dir, d.state.Current)
	d.startHeartbeat()
	defer d.stopHeartbeat()

	// ---- 检查更新 ----
	d.setPhase(update.DownloadPhaseChecking, "正在检查 GitHub Release…")
	info, err := update.CheckUpdate(d.dir, d.repo, update.PlatformAssetName(), d.allowPre)
	if err != nil {
		return d.fail(fmt.Sprintf("检查更新失败: %v", err))
	}
	d.mu.Lock()
	d.state.Current = info.Current.String()
	d.mu.Unlock()
	d.flush()
	if !info.HasUpdate {
		return d.fail(fmt.Sprintf("当前已是最新版本（%s），无需更新", info.Current))
	}
	if !info.SameMajor {
		return d.fail(fmt.Sprintf("发现新的大版本 %s（当前 %s）。大版本可能包含数据库结构变更，热更新已禁止，请手动下载整包升级。",
			info.Target, info.Current))
	}
	d.mu.Lock()
	d.state.Target = info.Target.String()
	d.state.AssetName = info.AssetName
	d.state.AssetSize = info.AssetSize
	d.mu.Unlock()
	d.flush()
	d.logf("发现可更新版本: %s -> %s（%s，%.2f MB）",
		info.Current, info.Target, info.AssetName, float64(info.AssetSize)/1048576)

	// ---- 获取校验文件（始终 GitHub 直连，作为信任锚） ----
	var expect string
	var checksumErr error
	if info.ChecksumURL != "" {
		expect, checksumErr = update.DownloadChecksum(info.ChecksumURL)
	}
	if len(d.mirrors) > 0 && expect == "" {
		// 镜像模式强制校验：不允许未经校验的镜像包进入安装
		reason := "该 Release 未提供 sha256 校验文件"
		if checksumErr != nil {
			reason = fmt.Sprintf("获取校验文件失败: %v", checksumErr)
		}
		return d.fail("镜像下载要求 sha256 校验，" + reason + "；已终止以避免安装未经校验的安装包")
	}
	d.mu.Lock()
	d.state.Checksum = expect
	d.mu.Unlock()
	d.flush()

	// ---- 下载（镜像优先、直连兜底；校验不符自动切换下一个镜像） ----
	pkgPath := filepath.Join(update.DownloadTmpDir(d.dir), info.AssetName)
	usedMirror, err := update.DownloadWithMirrors(info.DownloadURL, d.mirrors, pkgPath,
		func(received, total int64) {
			d.mu.Lock()
			d.state.Received, d.state.Total = received, total
			if total > 0 {
				pct := int(received * 100 / total)
				d.state.Message = fmt.Sprintf("正在下载更新包… %d%%", pct)
			}
			d.mu.Unlock()
		},
		func() bool { return fileExists(update.DownloadCancelFlagPath(d.dir)) },
		func(p string) error { return update.VerifyChecksum(p, expect) },
		func(mirror string, err error) { d.logf("镜像 %s 下载失败: %v", mirror, err) },
	)
	if err != nil {
		if fileExists(update.DownloadCancelFlagPath(d.dir)) {
			return d.fail("下载已取消")
		}
		return d.fail(fmt.Sprintf("下载更新包失败: %v", err))
	}
	d.mu.Lock()
	d.state.Mirror = usedMirror
	if usedMirror != "" {
		d.state.LastGoodMirror = usedMirror
	}
	d.mu.Unlock()
	d.flush()
	if usedMirror != "" {
		d.logf("下载完成（镜像 %s）: %s", usedMirror, pkgPath)
	} else {
		d.logf("下载完成（直连）: %s", pkgPath)
	}

	// ---- 校验 ----
	if expect != "" {
		d.setPhase(update.DownloadPhaseVerifying, "正在校验安装包…")
		if err := update.VerifyChecksum(pkgPath, expect); err != nil {
			_ = os.Remove(pkgPath)
			return d.fail(fmt.Sprintf("安装包校验失败: %v", err))
		}
		d.log("sha256 校验通过")
	} else {
		d.log("该 Release 未提供 sha256 校验文件，跳过校验")
	}

	// ---- 就绪 ----
	d.mu.Lock()
	d.state.Phase = update.DownloadPhaseReady
	d.state.Message = fmt.Sprintf("更新包已就绪（%s），可在管理页执行安装", info.Target)
	d.state.Error = ""
	d.state.FinishedAt = nowStamp()
	d.mu.Unlock()
	d.flush()
	d.logf("更新包已就绪: %s -> %s", info.Current, info.Target)
	return 0
}

// startHeartbeat 心跳：每 2 秒刷新 UpdatedAt 并落盘。
// 主程序据 UpdatedAt 停滞超过 30 秒判定下载进程已意外死亡（防止进度假转圈）。
func (d *Downloader) startHeartbeat() {
	d.stopHb = make(chan struct{})
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-d.stopHb:
				return
			case <-ticker.C:
				d.mu.Lock()
				d.state.UpdatedAt = nowStamp()
				d.mu.Unlock()
				d.flush()
			}
		}
	}()
}

func (d *Downloader) stopHeartbeat() {
	d.hbOnce.Do(func() { close(d.stopHb) })
}

// flush 将当前状态原子写入状态文件
func (d *Downloader) flush() {
	d.mu.Lock()
	s := d.state
	d.mu.Unlock()
	if err := update.WriteDownloadState(d.dir, &s); err != nil {
		fmt.Fprintf(os.Stderr, "写入下载状态失败: %v\n", err)
	}
}

func (d *Downloader) setPhase(phase, msg string) {
	d.mu.Lock()
	d.state.Phase = phase
	d.state.Message = msg
	d.state.Error = ""
	d.mu.Unlock()
	d.flush()
	d.log(msg)
}

func (d *Downloader) fail(msg string) int {
	d.mu.Lock()
	d.state.Phase = update.DownloadPhaseFailed
	d.state.Message = "下载失败"
	d.state.Error = msg
	d.state.FinishedAt = nowStamp()
	d.mu.Unlock()
	d.flush()
	d.log("失败: " + msg)
	return 1
}

func (d *Downloader) log(msg string) {
	line := nowStamp() + " " + msg
	fmt.Println(line)
	if f, err := os.OpenFile(filepath.Join(d.dir, "updates", "update.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); err == nil {
		_, _ = f.WriteString(line + "\n")
		_ = f.Close()
	}
}

func (d *Downloader) logf(format string, args ...interface{}) {
	d.log(fmt.Sprintf(format, args...))
}

// ---------- 阶段二：安装 ----------

// Updater 安装阶段控制器：仅消费已就绪的本地安装包，全程不访问外网。
type Updater struct {
	mu    sync.Mutex
	state State
	dir   string
	port  string

	target   string // 目标版本（来自下载状态）
	pkgPath  string // 安装包路径
	checksum string // 期望的 sha256（为空表示无校验文件）

	retryCh    chan struct{} // 用户请求重试时发送
	backupDir  string
	httpServer *http.Server
	doneCh     chan struct{} // 流程结束（成功/失败均算结束）
	doneOnce   sync.Once
	logFile    *os.File
}

func runInstall(args []string) {
	fs := flag.NewFlagSet("install", flag.ExitOnError)
	dirFlag := fs.String("dir", "", "安装目录（默认为 update.exe 所在目录）")
	port := fs.String("port", "8080", "安装进度页面端口")
	_ = fs.Parse(args)

	dir := resolveDir(*dirFlag)

	// ---- 前置检查：未就绪 / 包缺失 / 预检失败时直接退出，服务不受影响 ----
	st := update.ReadDownloadState(dir)
	if st == nil || st.Phase != update.DownloadPhaseReady {
		fmt.Println("未找到已就绪的下载状态（updates/tmp/download-state.json），请先执行下载阶段。")
		os.Exit(1)
	}
	if st.AssetName == "" {
		fmt.Println("下载状态缺少安装包信息，请重新执行下载阶段。")
		os.Exit(1)
	}
	pkgPath := filepath.Join(update.DownloadTmpDir(dir), st.AssetName)
	if !fileExists(pkgPath) {
		fmt.Println("安装包不存在或已被清理，请重新执行下载阶段。")
		os.Exit(1)
	}
	if err := update.CheckInstallPreconditions(dir, fileSize(pkgPath)); err != nil {
		fmt.Printf("安装前预检失败: %v\n", err)
		os.Exit(1)
	}

	u := &Updater{
		state: State{
			Phase:          phaseVerifyPkg,
			CurrentVersion: update.CurrentVersion(dir).String(),
			TargetVersion:  st.Target,
			Logs:           []string{},
		},
		dir:      dir,
		port:     *port,
		target:   st.Target,
		pkgPath:  pkgPath,
		checksum: st.Checksum,
		retryCh:  make(chan struct{}, 1),
		doneCh:   make(chan struct{}),
	}
	u.openLog()
	defer u.closeLog()

	fmt.Printf("排班系统热更新程序（安装阶段）\n安装目录: %s\n当前版本: %s → 目标版本: %s\n\n",
		dir, u.state.CurrentVersion, u.target)

	// 1. 停止服务（随后 8080 由进度页面接管）
	u.setPhase(phaseStopping, "正在停止服务…")
	if _, err := update.StopServer(u.dir); err != nil {
		u.logf("停止服务: %v", err)
	} else {
		u.log("服务已停止")
	}

	// 2. 在 8080 启动安装进度页面（安装不可终止；失败可重试或自动恢复）
	if err := u.startPage(u.port); err != nil {
		// 页面起不来不能阻断安装，退化为命令行模式
		fmt.Printf("警告: 进度页面启动失败（%v），将以命令行模式继续\n", err)
	} else {
		fmt.Printf("安装进度页面: http://localhost:%s\n", u.port)
	}

	// 3. 执行安装流程（失败后可重试，仅重跑本地安装）
	go u.runInstall()

	// 4. 等待流程结束
	<-u.doneCh

	// 给浏览器一点时间展示最终状态并完成自动刷新，然后退出自身进程
	time.Sleep(1500 * time.Millisecond)
	if u.httpServer != nil {
		_ = u.httpServer.Close()
	}
	fmt.Println("更新程序退出。")
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

// ---------- 安装流程 ----------

// runInstall 执行安装流程；失败后等待用户重试，超时则自动恢复更新前的版本
func (u *Updater) runInstall() {
	defer u.finish()

	for attempt := 1; ; attempt++ {
		if attempt > 1 {
			u.logf("==== 第 %d 次尝试 ====", attempt)
			u.backupDir = ""
			u.mu.Lock()
			u.state.Progress = 0
			u.mu.Unlock()
		}

		if u.doInstall() {
			return
		}

		// 失败：保持进度页面（8080 仍由 updater 持有），等待重试或自动恢复
		u.mu.Lock()
		u.state.AutoAbortIn = 60
		u.mu.Unlock()
		u.log("安装失败，60 秒后自动恢复更新前的版本（可点击「重试更新」立即重试）")

		select {
		case <-u.retryCh:
			u.mu.Lock()
			u.state.Phase = phaseVerifyPkg
			u.state.AutoAbortIn = 0
			u.state.Error = ""
			u.mu.Unlock()
			continue
		case <-time.After(60 * time.Second):
			u.log("超过 60 秒未重试，自动恢复更新前的版本")
		}
		u.restartOldServer("已恢复更新前的版本")
		u.writeResult(phaseRolledBack, "安装失败，已自动恢复更新前的版本")
		return
	}
}

// doInstall 执行一次安装（本地包，无网络访问）。
// 返回 true：新版本已重启并验证存活；false：失败，server 未重启（由 runInstall 决定重试或恢复）。
func (u *Updater) doInstall() bool {
	// ---- 校验安装包 ----
	u.setPhase(phaseVerifyPkg, "正在校验安装包…")
	if u.checksum != "" {
		if err := update.VerifyChecksum(u.pkgPath, u.checksum); err != nil {
			return u.fail(fmt.Sprintf("安装包校验失败: %v", err))
		}
		u.log("sha256 校验通过")
	} else {
		u.log("未记录校验值，跳过校验")
	}

	// ---- 备份 ----
	u.setPhase(phaseBackingUp, "正在备份当前版本…")
	backupDir, err := update.Backup(u.dir, update.CurrentVersion(u.dir).String()+"-"+update.NowLabel())
	if err != nil {
		return u.fail(fmt.Sprintf("备份当前版本失败: %v", err))
	}
	u.backupDir = backupDir
	u.logf("已备份到 %s", backupDir)

	// ---- 解压并替换 ----
	u.setPhase(phaseInstalling, "正在应用新版本…")
	stagingDir := filepath.Join(u.dir, "updates", "staging")
	_ = os.RemoveAll(stagingDir)
	if _, err := update.ExtractPackage(u.pkgPath, stagingDir); err != nil {
		return u.failWithRollback(fmt.Sprintf("解压更新包失败: %v", err))
	}
	if err := update.ApplyStaging(u.dir, stagingDir); err != nil {
		u.logf("应用新版本失败: %v", err)
		return u.failWithRollback(fmt.Sprintf("应用新版本失败: %v", err))
	}
	targetVer, err := update.ParseVersion(u.target)
	if err != nil {
		targetVer = update.Version{Raw: u.target}
	}
	if err := update.WriteVersion(u.dir, targetVer); err != nil {
		u.logf("警告: 写入版本号失败: %v", err)
	}
	u.logf("已应用新版本 %s", u.target)

	// ---- 重启 server ----
	u.setPhase(phaseRestarting, "更新完成，正在重启服务…")
	u.log("正在重启服务，页面即将自动刷新…")
	u.stopPage() // 必须先释放 8080，server 才能绑定

	if err := update.StartServer(u.dir); err != nil {
		return u.failAfterRestart(fmt.Sprintf("启动新版本失败: %v", err))
	}

	// ---- 健康检查 ----
	u.setPhase(phaseVerifying, "正在验证服务…")
	if err := update.WaitServerAlive(60 * time.Second); err != nil {
		return u.failAfterRestart(fmt.Sprintf("新版本健康检查失败: %v", err))
	}

	// ---- 成功 ----
	u.setPhase(phaseDone, fmt.Sprintf("更新成功！已升级到 %s，服务恢复正常", u.target))
	u.logf("更新成功: %s -> %s，服务已验证存活", u.state.CurrentVersion, u.target)
	u.writeResult(phaseDone, fmt.Sprintf("更新成功，已升级到 %s", u.target))
	_ = os.RemoveAll(filepath.Join(u.dir, "updates", "staging"))
	// 就绪态随之失效：安装包与下载状态一起清理，避免重复安装同一版本
	_ = os.RemoveAll(update.DownloadTmpDir(u.dir))
	return true
}

// failWithRollback 记录失败并立即回滚文件（不重启 server，页面保持可用可重试）
func (u *Updater) failWithRollback(msg string) bool {
	u.log("正在回滚到更新前的版本…")
	if u.backupDir != "" {
		if err := update.Rollback(u.dir, u.backupDir); err != nil {
			u.logf("回滚失败: %v", err)
			u.fail(msg + "（且回滚失败，请手动处理）")
			return false
		}
		u.log("已回滚到更新前的版本")
	}
	return u.fail(msg)
}

// failAfterRestart 重启/验证阶段失败：此时页面已释放，只能回滚并尽力恢复服务，
// 结果写入 updates/last-result.json 供管理端查看。
func (u *Updater) failAfterRestart(msg string) bool {
	u.logf("%s，正在回滚", msg)
	if u.backupDir != "" {
		if err := update.Rollback(u.dir, u.backupDir); err != nil {
			u.logf("回滚失败: %v", err)
			u.writeResult(phaseFailed, msg+"（回滚失败，请手动处理）")
			u.finish()
			os.Exit(3)
		}
	}
	if err := update.StartServer(u.dir); err != nil {
		u.logf("恢复旧版本失败: %v", err)
		u.writeResult(phaseFailed, msg+"（旧版本恢复失败，请手动启动 server.exe）")
		u.finish()
		os.Exit(3)
	}
	if err := update.WaitServerAlive(60 * time.Second); err != nil {
		u.writeResult(phaseFailed, msg+"（旧版本未就绪，请检查 server.err.log）")
		u.finish()
		os.Exit(3)
	}
	u.log("已回滚并恢复更新前的版本")
	u.writeResult(phaseFailed, msg+"（已自动回滚并恢复更新前的版本）")
	u.finish()
	os.Exit(3)
	return false // 不会执行到这里
}

// restartOldServer 恢复/启动更新前的 server。
// 这是唯一负责「释放 8080 → 启动 server → 验证」的终点路径。
func (u *Updater) restartOldServer(msg string) {
	u.setPhase(phaseRestarting, msg+"，正在重启服务…")
	u.stopPage()

	if update.ServerRunning() {
		u.log("服务已在运行，等待就绪")
	} else if err := update.StartServer(u.dir); err != nil {
		u.fail(msg + "，但服务重启失败，请手动启动 server.exe")
		u.writeResult(phaseFailed, msg+"，服务重启失败")
		return
	}
	if err := update.WaitServerAlive(60 * time.Second); err != nil {
		u.fail(msg + "，但服务未在预期时间内就绪，请检查 server.err.log")
		u.writeResult(phaseFailed, msg+"，服务未就绪")
		return
	}
	u.setPhase(phaseDone, msg)
}

// ---------- 状态管理 ----------

func (u *Updater) setPhase(phase, msg string) {
	u.mu.Lock()
	u.state.Phase = phase
	u.state.Message = msg
	u.state.Error = ""
	u.state.Progress = phaseProgress(phase)
	u.mu.Unlock()
	u.log(msg)
}

func phaseProgress(phase string) int {
	switch phase {
	case phaseVerifyPkg:
		return 10
	case phaseStopping:
		return 20
	case phaseBackingUp:
		return 45
	case phaseInstalling:
		return 70
	case phaseRestarting:
		return 90
	case phaseVerifying:
		return 96
	case phaseDone:
		return 100
	default:
		return 0
	}
}

// fail 记录失败状态（不重启 server，页面保持可用，等待重试或自动恢复）
func (u *Updater) fail(msg string) bool {
	u.mu.Lock()
	u.state.Phase = phaseFailed
	u.state.Message = "安装失败"
	u.state.Error = msg
	u.state.Progress = 0
	u.mu.Unlock()
	u.log("失败: " + msg)
	return false
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
	line := nowStamp() + " " + msg
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

// LastResult 最近一次更新的结果（写入 updates/last-result.json，供管理端查看）
type LastResult struct {
	Phase   string `json:"phase"` // done / failed / rolled-back
	Message string `json:"message"`
	From    string `json:"from"`
	To      string `json:"to"`
	At      string `json:"at"`
}

// writeResult 落盘最近一次更新结果
func (u *Updater) writeResult(phase, message string) {
	r := LastResult{
		Phase:   phase,
		Message: message,
		From:    u.state.CurrentVersion,
		To:      u.target,
		At:      nowStamp(),
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Join(u.dir, "updates"), 0755)
	_ = os.WriteFile(filepath.Join(u.dir, "updates", "last-result.json"), data, 0644)
}
