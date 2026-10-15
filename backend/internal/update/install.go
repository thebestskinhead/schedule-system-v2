package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ReplaceTargets 热更新允许替换的部署物。
// 注意：绝不动 configs/（数据库密码等配置）、migrations/、init_db.sql、
// 日志与备份目录——热更新只覆盖「前后端构建产物 + 服务可执行文件」。
var ReplaceTargets = []string{"server.exe", "server", "dist", "dist-mobile"}

// Backup 将部署物备份到 <dir>/updates/backup/<label>/，返回备份目录路径。
func Backup(dir, label string) (string, error) {
	backupDir := filepath.Join(dir, "updates", "backup", label)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("创建备份目录失败: %w", err)
	}
	for _, name := range ReplaceTargets {
		src := filepath.Join(dir, name)
		if _, err := os.Stat(src); os.IsNotExist(err) {
			continue // 该部署物不存在（如未构建移动端），跳过
		}
		dst := filepath.Join(backupDir, name)
		if err := copyPath(src, dst); err != nil {
			return backupDir, fmt.Errorf("备份 %s 失败: %w", name, err)
		}
	}
	// 同时备份当前版本号，便于回滚后恢复
	_ = copyFile(VersionFilePath(dir), filepath.Join(backupDir, VersionFileName))
	return backupDir, nil
}

// Rollback 从备份目录恢复部署物并写回旧版本号
func Rollback(dir, backupDir string) error {
	if backupDir == "" {
		return fmt.Errorf("备份目录为空，无法回滚")
	}
	for _, name := range ReplaceTargets {
		src := filepath.Join(backupDir, name)
		if _, err := os.Stat(src); os.IsNotExist(err) {
			continue
		}
		dst := filepath.Join(dir, name)
		if err := os.RemoveAll(dst); err != nil {
			return fmt.Errorf("回滚删除 %s 失败: %w", name, err)
		}
		if err := copyPath(src, dst); err != nil {
			return fmt.Errorf("回滚恢复 %s 失败: %w", name, err)
		}
	}
	if vf := filepath.Join(backupDir, VersionFileName); fileExists(vf) {
		data, err := os.ReadFile(vf)
		if err == nil {
			_ = os.WriteFile(VersionFilePath(dir), data, 0644)
		}
	}
	return nil
}

// ExtractPackage 解压下载的发布包到 staging 目录，返回 staging 路径。
// 支持 .zip 与 .tar.gz。
func ExtractPackage(archivePath, stagingDir string) (string, error) {
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return "", err
	}
	lower := strings.ToLower(archivePath)
	switch {
	case strings.HasSuffix(lower, ".zip"):
		if err := extractZip(archivePath, stagingDir); err != nil {
			return "", err
		}
	case strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz"):
		if err := extractTarGz(archivePath, stagingDir); err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("不支持的压缩包格式: %s", filepath.Base(archivePath))
	}
	return stagingDir, nil
}

// ApplyStaging 用 staging 中的产物替换安装目录的部署物。
// 跳过 updater 自身（Windows 上运行中的 exe 不可覆盖，重命名后由下次更新生效）。
func ApplyStaging(dir, stagingDir string) error {
	for _, name := range ReplaceTargets {
		src := filepath.Join(stagingDir, name)
		if _, err := os.Stat(src); os.IsNotExist(err) {
			continue // 新包未包含该产物，保留现有
		}
		dst := filepath.Join(dir, name)
		if err := os.RemoveAll(dst); err != nil {
			return fmt.Errorf("替换前清理 %s 失败: %w", name, err)
		}
		if err := os.Rename(src, dst); err != nil {
			// 跨卷时 rename 失败，回退复制
			if cerr := copyPath(src, dst); cerr != nil {
				return fmt.Errorf("替换 %s 失败: %w", name, cerr)
			}
		}
	}

	// 更新程序自身：若新版包内带有 updater，重命名运行中的文件后落位（Windows 允许重命名运行中的 exe）
	newUpdater := filepath.Join(stagingDir, UpdaterExeName())
	curUpdater := filepath.Join(dir, UpdaterExeName())
	if fileExists(newUpdater) && fileExists(curUpdater) {
		if !sameFile(newUpdater, curUpdater) {
			oldUpdater := curUpdater + ".old"
			_ = os.RemoveAll(oldUpdater)
			if err := os.Rename(curUpdater, oldUpdater); err == nil {
				if err := copyPath(newUpdater, curUpdater); err != nil {
					// 新 updater 落位失败不影响本次更新，还原旧文件
					_ = os.Rename(oldUpdater, curUpdater)
				}
			}
		}
	}
	return nil
}

// VerifyChecksum 校验文件 sha256 与期望值是否一致
func VerifyChecksum(filePath, expect string) error {
	if expect == "" {
		return nil
	}
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(got, expect) {
		return fmt.Errorf("sha256 校验失败: 期望 %s，实际 %s", expect, got)
	}
	return nil
}

// ---------- 内部工具 ----------

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// sameFile 判断两个文件内容是否相同（按大小+sha256）
func sameFile(a, b string) bool {
	fa, err := os.ReadFile(a)
	if err != nil {
		return false
	}
	fb, err := os.ReadFile(b)
	if err != nil {
		return false
	}
	return string(fa) == string(fb)
}

// copyPath 复制文件或目录（递归）
func copyPath(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := copyPath(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func extractZip(archivePath, destDir string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("打开 zip 失败: %w", err)
	}
	defer r.Close()
	for _, f := range r.File {
		target := filepath.Join(destDir, f.Name)
		// 防路径穿越
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("压缩包含非法路径: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		if err := extractOne(f.Open, target, f.Mode()); err != nil {
			return err
		}
	}
	return nil
}

func extractTarGz(archivePath, destDir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("打开 gzip 失败: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取 tar 失败: %w", err)
		}
		target := filepath.Join(destDir, hdr.Name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("压缩包含非法路径: %s", hdr.Name)
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			if err := extractOne(func() (io.ReadCloser, error) { return io.NopCloser(tr), nil }, target, os.FileMode(hdr.Mode)); err != nil {
				return err
			}
		}
	}
	return nil
}

func extractOne(open func() (io.ReadCloser, error), target string, mode os.FileMode) error {
	rc, err := open()
	if err != nil {
		return err
	}
	defer rc.Close()
	out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode.Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, rc); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// NowLabel 生成备份目录用的时间戳标签
func NowLabel() string {
	return time.Now().Format("20060102-150405")
}
