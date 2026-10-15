//go:build e2e

package service

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"testing"
)

// 端到端验证：用真实教务会话下载课表并与参考文件比对。
//
//	E2E_COOKIE="bzb_jsxsd=...; bzb_njw=...; SERVERID=124" \
//	go test -tags e2e ./internal/service/ -run TestEduDownloadE2E -v
//
// 可选 E2E_REF 指定参考文件路径，比对字节一致性。
func TestEduDownloadE2E(t *testing.T) {
	cookieStr := os.Getenv("E2E_COOKIE")
	if cookieStr == "" {
		t.Skip("未设置 E2E_COOKIE，跳过端到端测试")
	}

	jar, _ := cookiejar.New(nil)
	eduURL, _ := url.Parse(eduBaseURL)
	for _, pair := range splitCookieHeader(cookieStr) {
		jar.SetCookies(eduURL, []*http.Cookie{{Name: pair[0], Value: pair[1], Path: "/"}})
	}

	data, filename, err := NewEduTimetableService().Download(context.Background(), jar)
	if err != nil {
		t.Fatalf("下载失败: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("下载内容为空")
	}
	t.Logf("下载成功: %s, %d 字节", filename, len(data))

	if ref := os.Getenv("E2E_REF"); ref != "" {
		want, err := os.ReadFile(ref)
		if err != nil {
			t.Fatalf("读取参考文件失败: %v", err)
		}
		if len(want) != len(data) {
			t.Fatalf("字节数不一致: 下载 %d, 参考 %d", len(data), len(want))
		}
		for i := range want {
			if want[i] != data[i] {
				t.Fatalf("内容不一致，首个差异位于第 %d 字节", i)
			}
		}
		t.Log("与参考文件逐字节一致")
	}
}

// splitCookieHeader 将 "k=v; k2=v2" 拆为键值对。
func splitCookieHeader(s string) [][2]string {
	var out [][2]string
	for _, part := range strings.Split(s, ";") {
		part = strings.TrimSpace(part)
		if eq := strings.IndexByte(part, '='); eq > 0 {
			out = append(out, [2]string{part[:eq], part[eq+1:]})
		}
	}
	return out
}
