package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
)

// 教务系统课表下载。基于扫码登录得到的 Cookie，无需用户手工导出文件。
//
// 两步：
//  1. GET  /jsxsd/xskb/xskb_list.do        —— 读学年学期下拉框的选中项，即当前学期
//  2. POST /jsxsd/xskb/xskb_print.do       —— 下载课表（application/msexcel 的 .xls）
//
// 注意：会话失效时第 2 步仍返回 HTTP 200，但响应体是登录页 HTML，
// 因此判定成败以 Content-Type 或文件魔数为准，不能只看状态码。

const (
	eduTimetableListURL  = eduBaseURL + "/jsxsd/xskb/xskb_list.do"
	eduTimetablePrintURL = eduBaseURL + "/jsxsd/xskb/xskb_print.do"

	// 默认节次模式模板 ID，与课表页「导出」按钮提交的一致
	eduDefaultKbjcmsid = "0B841F8A531A4C05BE8DB7DB4B40AEF1"

	// OLE2 复合文档（旧版 .xls）与 ZIP（.xlsx）文件头
	xlsMagic = "\xd0\xcf\x11\xe0\xa1\xb1\x1a\xe1"
	zipMagic = "PK\x03\x04"
)

// ErrEduSessionInvalid 表示教务会话已失效，需要重新扫码登录
var ErrEduSessionInvalid = fmt.Errorf("教务登录已失效，请重新扫码登录")

// EduTimetableService 教务课表下载服务
type EduTimetableService struct{}

// NewEduTimetableService 创建课表下载服务
func NewEduTimetableService() *EduTimetableService {
	return &EduTimetableService{}
}

// Download 用给定 Cookie 下载当前学期课表，返回文件字节与服务端文件名。
// 会话失效时返回 ErrEduSessionInvalid；ctx 用于中断阻塞的网络请求。
func (s *EduTimetableService) Download(ctx context.Context, jar *cookiejar.Jar) ([]byte, string, error) {
	client := newEduClient(jar)

	if u, err := url.Parse(eduTimetableListURL); err == nil {
		names := make([]string, 0, 8)
		for _, ck := range jar.Cookies(u) {
			names = append(names, ck.Name)
		}
		log.Printf("[edu-timetable] 请求课表页将携带 %d 枚 Cookie: %s", len(names), strings.Join(names, ","))
	}

	term, err := s.detectCurrentTerm(ctx, client)
	if err != nil {
		return nil, "", err
	}

	return s.fetchTimetable(ctx, client, term)
}

// detectCurrentTerm 读课表列表页，取学年学期下拉框中 selected 的选项。
// 该选中项由服务端按当前日期判定，且并非列表首项，必须读 selected。
func (s *EduTimetableService) detectCurrentTerm(ctx context.Context, client *http.Client) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", eduTimetableListURL, nil)
	commonHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("访问课表页失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取课表页失败: %w", err)
	}

	term := extractSelectedTerm(string(body))
	if term == "" {
		head := body
		if len(head) > 120 {
			head = head[:120]
		}
		log.Printf("[edu-timetable] 课表页未取到当前学期: status=%d 字节数=%d 头部=%q",
			resp.StatusCode, len(body), strings.ReplaceAll(string(head), "\n", " "))
		return "", ErrEduSessionInvalid
	}
	return term, nil
}

// extractSelectedTerm 从课表列表页 HTML 中提取学年学期下拉框的选中项。
func extractSelectedTerm(html string) string {
	selectRe := regexp.MustCompile(`(?is)<select[^>]*name=["']xnxq01id["'][^>]*>(.*?)</select>`)
	m := selectRe.FindStringSubmatch(html)
	if len(m) < 2 {
		return ""
	}

	optionRe := regexp.MustCompile(`(?is)<option[^>]*value=["']([^"']+)["'][^>]*>`)
	selectedRe := regexp.MustCompile(`(?is)selected`)

	for _, opt := range optionRe.FindAllStringSubmatch(m[1], -1) {
		if len(opt) < 2 {
			continue
		}
		if selectedRe.MatchString(opt[0]) {
			return opt[1]
		}
	}
	return ""
}

// fetchTimetable 提交导出请求，取回课表文件。
func (s *EduTimetableService) fetchTimetable(ctx context.Context, client *http.Client, term string) ([]byte, string, error) {
	req, _ := http.NewRequestWithContext(ctx, "POST", eduTimetablePrintURL, nil)
	commonHeaders(req)
	req.Header.Set("Referer", eduTimetableListURL)

	q := req.URL.Query()
	q.Set("xnxq01id", term)
	q.Set("zc", "")
	q.Set("kbjcmsid", eduDefaultKbjcmsid)
	q.Set("wkbkc", "1")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("下载课表失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("读取课表内容失败: %w", err)
	}

	if !isExcelResponse(resp.Header.Get("Content-Type"), body) {
		// 会话失效时服务端返回登录页 HTML 且状态码仍为 200
		head := body
		if len(head) > 120 {
			head = head[:120]
		}
		log.Printf("[edu-timetable] 导出响应不是课表文件: status=%d content-type=%q 字节数=%d 头部=%q",
			resp.StatusCode, resp.Header.Get("Content-Type"), len(body), strings.ReplaceAll(string(head), "\n", " "))
		return nil, "", ErrEduSessionInvalid
	}

	return body, filenameFromDisposition(resp.Header.Get("Content-disposition")), nil
}

// isExcelResponse 判断响应是否为可用的课表文件
func isExcelResponse(contentType string, body []byte) bool {
	if strings.Contains(strings.ToLower(contentType), "msexcel") {
		return true
	}
	head := string(body)
	if len(head) >= len(xlsMagic) && head[:len(xlsMagic)] == xlsMagic {
		return true
	}
	if len(head) >= len(zipMagic) && head[:len(zipMagic)] == zipMagic {
		return true
	}
	return false
}

// filenameFromDisposition 从 Content-disposition 解析文件名，兼容 GBK 编码。
func filenameFromDisposition(disp string) string {
	const fallback = "timetable.xls"
	if !strings.Contains(disp, "filename=") {
		return fallback
	}

	raw := strings.TrimSpace(strings.Split(disp, "filename=")[len(strings.Split(disp, "filename="))-1])
	raw = strings.Trim(raw, `"`)

	// 服务端以 GBK 输出中文文件名，按 latin-1 再解 GBK 还原
	if decoded, err := gbkDecode(raw); err == nil && decoded != "" {
		return decoded
	}
	if raw == "" {
		return fallback
	}
	return raw
}

// newEduClient 创建携带指定 CookieJar 的 HTTP 客户端
func newEduClient(jar *cookiejar.Jar) *http.Client {
	return &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}
}

// gbkDecode 还原以 GBK 编码、被 HTTP 头按 latin-1 逐字符映射的中文字符串。
// HTTP 头以 latin-1 解析，GBK 的双字节被拆成两个 0-255 的码点，
// 这里先按码点取回原始字节，再用 GBK 解码还原中文。
func gbkDecode(s string) (string, error) {
	raw := make([]byte, 0, len(s))
	for _, r := range s {
		if r > 0xFF {
			// 已含非 latin-1 字符，说明并非该编码形态，原样返回
			return s, nil
		}
		raw = append(raw, byte(r))
	}
	out, err := simplifiedchinese.GBK.NewDecoder().Bytes(raw)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
