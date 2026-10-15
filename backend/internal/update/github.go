package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// GitHubHint 受限网络（GitHub 不可达）时的处置建议，统一追加到网络类错误信息后
const GitHubHint = "；若服务器处于受限网络，可设置 HTTPS_PROXY 环境变量指向代理后重启服务或更新程序"

// friendlyNetErr 将底层网络错误转换为可操作的中文提示。
// 原始错误形如 "dial tcp: lookup api.github.com: no such host"，
// 对管理员没有指导意义，这里按错误类型分类并附上处置建议。
func friendlyNetErr(err error) error {
	if err == nil {
		return nil
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return fmt.Errorf("无法解析 api.github.com（DNS 解析失败: %s）%s", dnsErr.Err, GitHubHint)
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, os.ErrDeadlineExceeded) {
		return fmt.Errorf("连接 GitHub 超时（可能被防火墙或网络策略拦截）%s", GitHubHint)
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return fmt.Errorf("连接 GitHub 超时（可能被防火墙或网络策略拦截）%s", GitHubHint)
	}
	return fmt.Errorf("无法连接 GitHub（%v）%s", err, GitHubHint)
}

// DefaultRepo 默认 GitHub 仓库（owner/name）
const DefaultRepo = "thebestskinhead/schedule-system-v2"

// ghAPIBaseURL GitHub API 基础地址。测试时可覆盖为 httptest 服务器地址，
// 避免单测依赖外网（生产路径不变）。
var ghAPIBaseURL = "https://api.github.com"

// SetGitHubAPIBaseForTest 覆盖 GitHub API 地址并返回还原函数。
// 仅供测试注入 mock 服务器使用，生产代码不应调用。
func SetGitHubAPIBaseForTest(base string) (restore func()) {
	orig := ghAPIBaseURL
	ghAPIBaseURL = base
	return func() { ghAPIBaseURL = orig }
}

// Release GitHub Release 信息（仅取所需字段）
type Release struct {
	TagName    string  `json:"tag_name"`
	Name       string  `json:"name"`
	Prerelease bool    `json:"prerelease"`
	Draft      bool    `json:"draft"`
	Body       string  `json:"body"`
	Assets     []Asset `json:"assets"`
}

// Asset Release 资产
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// Version 解析并返回该 Release 的版本号；无法解析时返回错误
func (r Release) Version() (Version, error) {
	return ParseVersion(r.TagName)
}

// UpdateInfo 一次更新检查的结果
type UpdateInfo struct {
	Current      Version `json:"current"`      // 当前部署版本
	Target       Version `json:"target"`       // 目标版本；无可更新时等于 Current，跨大版本时为最新大版本
	HasUpdate    bool    `json:"hasUpdate"`    // 是否存在可热更新的同大版本新版本
	SameMajor    bool    `json:"sameMajor"`    // 目标与当前同大版本（可热更新）
	CrossMajor   bool    `json:"crossMajor"`   // 无可热更新的版本，但存在更新的大版本（需手动整包升级）
	AssetName    string  `json:"assetName"`    // 目标平台对应的资产文件名
	DownloadURL  string  `json:"downloadUrl"`  // 资产下载地址
	AssetSize    int64   `json:"assetSize"`    // 资产大小（字节）
	ChecksumURL  string  `json:"checksumUrl"`  // 可选：sha256 校验文件地址
	ReleaseNotes string  `json:"releaseNotes"` // Release 说明（可能很长，按需截断）
}

// ghClient GitHub API 客户端
type ghClient struct {
	repo    string
	token   string
	baseURL string // <api>/repos/<repo>，测试时可指向 httptest 服务器
	client  *http.Client
}

func newGHClient(repo, token string) *ghClient {
	return &ghClient{
		repo:    repo,
		token:   token,
		baseURL: fmt.Sprintf("%s/repos/%s", ghAPIBaseURL, repo),
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (g *ghClient) apiURL(path string) string {
	return g.baseURL + path
}

func (g *ghClient) newReq(method, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "schedule-system-updater")
	if g.token != "" {
		req.Header.Set("Authorization", "Bearer "+g.token)
	}
	return req, nil
}

// ListReleases 拉取最近的 releases（按发布时间倒序，GitHub 已排序）
func (g *ghClient) ListReleases() ([]Release, error) {
	req, err := g.newReq("GET", g.apiURL("/releases?per_page=30"))
	if err != nil {
		return nil, err
	}
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, friendlyNetErr(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API 返回 %d: %s", resp.StatusCode, truncate(string(body), 200))
	}
	var releases []Release
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, fmt.Errorf("解析 GitHub 响应失败: %w", err)
	}
	return releases, nil
}

// CheckUpdate 检查是否存在可热更新的版本。
// 规则：同大版本且版本号严格大于当前版本；跳过 draft / prerelease（除非 allowPrerelease）。
// 找不到匹配当前平台资产的 release 直接跳过（继续向更早的 release 找）。
func CheckUpdate(dir, repo, platformAssetName string, allowPrerelease bool) (*UpdateInfo, error) {
	token := os.Getenv("GITHUB_TOKEN")
	g := newGHClient(repo, token)

	current := CurrentVersion(dir)
	releases, err := g.ListReleases()
	if err != nil {
		return nil, err
	}

	// 防御：本地按版本号再排一次序（GitHub 默认按创建时间，个别仓库可能不严格）
	sort.SliceStable(releases, func(i, j int) bool {
		vi, ei := releases[i].Version()
		vj, ej := releases[j].Version()
		if ei != nil || ej != nil {
			return false
		}
		return vi.Compare(vj) > 0
	})

	var newest Version // 跨大版本也记录，用于提示手动升级
	var newestFound bool
	for _, r := range releases {
		if r.Draft {
			continue
		}
		if r.Prerelease && !allowPrerelease {
			continue
		}
		v, err := r.Version()
		if err != nil {
			continue
		}
		if !newestFound || v.Compare(newest) > 0 {
			newest, newestFound = v, true
		}
		if v.Compare(current) <= 0 {
			continue
		}
		asset, ok := findAsset(r.Assets, platformAssetName)
		if !ok {
			continue // 该 release 没有本平台资产，继续找更早的
		}
		info := &UpdateInfo{
			Current:      current,
			Target:       v,
			HasUpdate:    true,
			SameMajor:    current.SameMajor(v),
			AssetName:    asset.Name,
			DownloadURL:  asset.BrowserDownloadURL,
			AssetSize:    asset.Size,
			ReleaseNotes: truncate(r.Body, 4000),
		}
		// 可选的 sha256 校验文件：<asset>.sha256
		if ck, ok := findAsset(r.Assets, asset.Name+".sha256"); ok {
			info.ChecksumURL = ck.BrowserDownloadURL
		}
		return info, nil
	}

	// 没有可热更新的版本。若有更新但跨大版本，标记出来提示手动升级
	info := &UpdateInfo{Current: current, Target: current, HasUpdate: false, SameMajor: true}
	if newestFound && newest.Compare(current) > 0 && !current.SameMajor(newest) {
		info.CrossMajor = true
		info.Target = newest
		info.SameMajor = false
	}
	return info, nil
}

// findAsset 在资产列表中查找指定文件名的资产
func findAsset(assets []Asset, name string) (Asset, bool) {
	for _, a := range assets {
		if strings.EqualFold(a.Name, name) {
			return a, true
		}
	}
	return Asset{}, false
}

// PlatformAssetName 返回当前平台对应的发布资产文件名
func PlatformAssetName() string {
	if isWindows() {
		return "schedule-system-v2-windows-amd64.zip"
	}
	return "schedule-system-v2-linux-amd64.tar.gz"
}

// DownloadFile 带进度回调的下载。onProgress(接收字节数, 总字节数) 每收到一块调用一次。
func DownloadFile(url, dest string, onProgress func(received, total int64)) error {
	return DownloadFileCancelable(url, dest, onProgress, nil)
}

// DownloadFileCancelable 可中止的下载。shouldAbort 返回 true 时立即中断并删除临时文件。
func DownloadFileCancelable(url, dest string, onProgress func(received, total int64), shouldAbort func() bool) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "schedule-system-updater")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := (&http.Client{Timeout: 0}).Do(req)
	if err != nil {
		return friendlyNetErr(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载返回 %d", resp.StatusCode)
	}

	tmp := dest + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	total := resp.ContentLength
	var received int64
	buf := make([]byte, 32*1024)
	for {
		if shouldAbort != nil && shouldAbort() {
			f.Close()
			os.Remove(tmp)
			return fmt.Errorf("下载已被用户中止")
		}
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				f.Close()
				os.Remove(tmp)
				return fmt.Errorf("写入失败: %w", werr)
			}
			received += int64(n)
			if onProgress != nil {
				onProgress(received, total)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			f.Close()
			os.Remove(tmp)
			return fmt.Errorf("下载中断: %w", err)
		}
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

// DownloadChecksum 下载并解析 sha256 校验文件，返回期望的十六进制摘要
func DownloadChecksum(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "schedule-system-updater")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return "", friendlyNetErr(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载校验文件返回 %d", resp.StatusCode)
	}
	// 文件内容可能是 "<hash>  <filename>" 或纯 hash
	fields := strings.Fields(string(body))
	if len(fields) == 0 {
		return "", fmt.Errorf("校验文件内容为空")
	}
	return strings.ToLower(fields[0]), nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

// AssetFileName 从下载地址推导本地文件名
func AssetFileName(downloadURL string) string {
	if i := strings.LastIndex(downloadURL, "/"); i >= 0 {
		return downloadURL[i+1:]
	}
	return "package.bin"
}

// JoinPath 小工具：拼接路径（统一走 filepath，避免各平台分隔符问题）
func JoinPath(parts ...string) string {
	return filepath.Join(parts...)
}
