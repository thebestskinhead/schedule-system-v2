package update

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// HealthCheckURL 服务健康检查地址（公开接口，无需认证）
const HealthCheckURL = "http://127.0.0.1:8080/api/v1/system/health"

// ServerExeName 服务端可执行文件名（按平台）
func ServerExeName() string {
	if isWindows() {
		return "server.exe"
	}
	return "server"
}

// UpdaterExeName 更新程序文件名（按平台）
func UpdaterExeName() string {
	if isWindows() {
		return "update.exe"
	}
	return "update"
}

func isWindows() bool { return runtime.GOOS == "windows" }

// ServerPort 服务监听端口（用于等待端口释放）
const ServerPort = "8080"

// StopServer 终止安装目录下的 server 进程，并等待其完全退出。
// 返回被终止的进程数；找不到进程时返回 0（服务本就没运行）。
// 注意：判定退出用「进程消失」而非「端口释放」——更新程序自身随后会占用 8080
// 展示进度页，端口等待会误判为 server 未退出。
func StopServer(dir string) (int, error) {
	exePath := filepath.Join(dir, ServerExeName())

	if isWindows() {
		// 优先按可执行路径精确匹配，避免误杀同名进程；
		// PowerShell 查询失败时回退为按映像名终止。
		count, err := stopServerWindows(exePath)
		if err != nil {
			// 回退：按进程名强杀（部署机通常只有本系统一个 server 实例）
			if fallbackErr := exec.Command("taskkill", "/F", "/IM", ServerExeName()).Run(); fallbackErr != nil {
				return 0, fmt.Errorf("终止 server 进程失败: %v / %v", err, fallbackErr)
			}
			waitProcessGoneWindows(ServerExeName(), 30*time.Second)
			return 1, nil
		}
		if count < 0 {
			return 0, fmt.Errorf("等待 server 进程退出超时")
		}
		return count, nil
	}

	// Linux：按完整路径匹配（-f 全命令行匹配）
	cmd := exec.Command("pkill", "-f", exePath)
	if err := cmd.Run(); err != nil {
		// pkill 退出码 1 表示没有匹配进程，视为服务未运行
		if ee, ok := err.(*exec.ExitError); !ok || ee.ExitCode() != 1 {
			return 0, fmt.Errorf("终止 server 进程失败: %w", err)
		}
		return 0, nil
	}
	// 等待进程消失
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if err := exec.Command("pgrep", "-f", exePath).Run(); err != nil {
			return 1, nil // pgrep 退出码 1 = 已无匹配进程
		}
		time.Sleep(300 * time.Millisecond)
	}
	return 1, fmt.Errorf("等待 server 进程退出超时")
}

// stopServerWindows 通过 PowerShell 按可执行路径终止 server.exe：
// 结束进程 → 结束其父批处理窗口 → 等待进程完全退出。
// 输出：0 = 没有匹配进程；-1 = 杀后仍未退出；N = 成功终止并等待退出。
func stopServerWindows(exePath string) (int, error) {
	ps := fmt.Sprintf(`
$procs = @(Get-CimInstance Win32_Process -Filter "Name='%s'" | Where-Object { $_.ExecutablePath -and $_.ExecutablePath -like '%s*' })
if ($procs.Count -eq 0) { Write-Output "0"; exit 0 }
foreach ($p in $procs) {
  Stop-Process -Id $p.ProcessId -Force
  # 结束父进程中的启动批处理窗口（cmd.exe 且命令行含本目录）
  $parent = Get-CimInstance Win32_Process -Filter "ProcessId=$($p.ParentProcessId)" -ErrorAction SilentlyContinue
  if ($parent -and $parent.Name -eq 'cmd.exe' -and $parent.CommandLine -and $parent.CommandLine -like '*%s*') {
    Stop-Process -Id $parent.ProcessId -Force
  }
}
$ids = @($procs | ForEach-Object { $_.ProcessId })
$deadline = (Get-Date).AddSeconds(30)
while ((Get-Date) -lt $deadline) {
  $alive = @(Get-Process -Id $ids -ErrorAction SilentlyContinue)
  if ($alive.Count -eq 0) { Write-Output $procs.Count; exit 0 }
  Start-Sleep -Milliseconds 300
}
Write-Output "-1"
`, ServerExeName(), escapeSingle(exeDir(exePath)), escapeSingle(exeDir(exePath)))

	out, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", ps).Output()
	if err != nil {
		return 0, err
	}
	n := 0
	fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &n)
	return n, nil
}

// waitProcessGoneWindows 轮询等待指定映像名的进程全部消失（taskkill 回退路径用）
func waitProcessGoneWindows(imageName string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		out, err := exec.Command("tasklist", "/FI", "IMAGENAME eq "+imageName, "/NH").Output()
		if err != nil || !strings.Contains(string(out), imageName) {
			return
		}
		time.Sleep(300 * time.Millisecond)
	}
}

func exeDir(p string) string { return filepath.Dir(p) }

func escapeSingle(s string) string { return strings.ReplaceAll(s, "'", "''") }

// StartServer 以分离进程方式启动 server，日志追加到安装目录的 server.out.log / server.err.log。
// 调用方进程退出后 server 继续运行（Windows/Linux 均适用）。
func StartServer(dir string) error {
	exePath := filepath.Join(dir, ServerExeName())
	if _, err := os.Stat(exePath); err != nil {
		return fmt.Errorf("找不到服务可执行文件: %s", exePath)
	}

	outLog, err := os.OpenFile(filepath.Join(dir, "server.out.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开 server.out.log 失败: %w", err)
	}
	errLog, err := os.OpenFile(filepath.Join(dir, "server.err.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		outLog.Close()
		return fmt.Errorf("打开 server.err.log 失败: %w", err)
	}

	cmd := exec.Command(exePath)
	cmd.Dir = dir
	cmd.Stdout = outLog
	cmd.Stderr = errLog
	cmd.Env = os.Environ()
	if isWindows() {
		// 不隐藏窗口：部署机可直接看到服务输出；需要后台运行时可改为 HideWindow: true
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	if err := cmd.Start(); err != nil {
		outLog.Close()
		errLog.Close()
		return fmt.Errorf("启动 server 失败: %w", err)
	}
	// 不等待子进程：updater 退出后 server 作为孤儿进程继续运行。
	// 文件句柄随 updater 退出由 OS 关闭，server 已继承写入端。
	_ = cmd.Process.Release()
	outLog.Close()
	errLog.Close()
	return nil
}

// WaitServerAlive 轮询健康检查接口直到服务存活或超时。
// 判定标准：能收到任意 HTTP 响应即视为存活（不同版本接口路径可能有差异，
// 旧版本没有 /system/health 会返回 403，但能响应说明进程已起来）。
func WaitServerAlive(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 3 * time.Second}
	var lastErr error
	for time.Now().Before(deadline) {
		resp, err := client.Get(HealthCheckURL)
		if err == nil {
			resp.Body.Close()
			return nil
		}
		lastErr = err
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("服务启动超时: %v", lastErr)
}

// ServerRunning 快速探测 server 是否在运行（端口是否可连接）
func ServerRunning() bool {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:"+ServerPort, 500*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// UpdaterPresent 安装目录中是否存在更新程序
func UpdaterPresent(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, UpdaterExeName()))
	return err == nil
}
