package service

import "testing"

// 学期提取：取学年学期下拉框中 selected 的选项，而非列表首项。
func TestExtractSelectedTerm(t *testing.T) {
	html := `<form>
	<select name="xnxq01id" id="xnxq01id">
		<option value="2026-2027-2"  >2026-2027-2</option>
		<option value="2026-2027-1"   selected="selected">2026-2027-1</option>
		<option value="2025-2026-2"  >2025-2026-2</option>
	</select>
	</form>`

	if got := extractSelectedTerm(html); got != "2026-2027-1" {
		t.Fatalf("选中学期 = %q, 期望 2026-2027-1", got)
	}
}

// 无 selected 属性时返回空串（视为未取到当前学期，调用方据此判定会话失效）。
// 刻意不回退首项：列表首项是下一学期，回退会导致导入错误学期的空白课表。
func TestExtractSelectedTermNoSelected(t *testing.T) {
	html := `<select name="xnxq01id">
		<option value="2026-2027-2">2026-2027-2</option>
		<option value="2026-2027-1">2026-2027-1</option>
	</select>`

	if got := extractSelectedTerm(html); got != "" {
		t.Fatalf("无 selected 时应返回空串, 实际 %q", got)
	}
}

// 页面无该下拉框（会话失效返回登录页）时返回空串。
func TestExtractSelectedTermMissing(t *testing.T) {
	if got := extractSelectedTerm(`<html><body>登录</body></html>`); got != "" {
		t.Fatalf("缺失场景应返回空串, 实际 %q", got)
	}
}

// 脚本用单引号、且属性顺序变化时仍能正确提取。
func TestExtractSelectedTermSingleQuote(t *testing.T) {
	html := `<select id="xnxq01id" name='xnxq01id'><option selected value='2025-2026-1'>x</option></select>`
	if got := extractSelectedTerm(html); got != "2025-2026-1" {
		t.Fatalf("单引号场景 = %q, 期望 2025-2026-1", got)
	}
}

// 响应判定：Content-Type 或文件魔数任一命中即视为课表文件。
func TestIsExcelResponse(t *testing.T) {
	xlsBody := append([]byte(xlsMagic), make([]byte, 100)...)

	cases := []struct {
		name        string
		contentType string
		body        []byte
		want        bool
	}{
		{"按 Content-Type", "application/msexcel", []byte("anything"), true},
		{"按 xls 魔数", "application/octet-stream", xlsBody, true},
		{"按 xlsx 魔数", "", append([]byte(zipMagic), make([]byte, 100)...), true},
		{"登录页 HTML", "text/html;charset=utf-8", []byte("<!DOCTYPE html><html>登录</html>"), false},
		{"空响应", "", nil, false},
	}
	for _, c := range cases {
		if got := isExcelResponse(c.contentType, c.body); got != c.want {
			t.Errorf("%s: isExcelResponse = %v, 期望 %v", c.name, got, c.want)
		}
	}
}

// 文件名解析：服务端以 GBK 输出中文名，需正确还原。
func TestFilenameFromDisposition(t *testing.T) {
	// "学生个人课表_.xls" 的 GBK 字节按 latin-1 映射后的形态
	gbk := "\u00d1\u00a7\u00c9\u00fa\u00b8\u00f6\u00c8\u00cb\u00bf\u00ce\u00b1\u00ed_.xls"
	disp := `attachment; filename=` + gbk

	got := filenameFromDisposition(disp)
	if got != "学生个人课表_.xls" {
		t.Fatalf("文件名还原 = %q, 期望 学生个人课表_.xls", got)
	}

	if got := filenameFromDisposition("attachment"); got != "timetable.xls" {
		t.Fatalf("缺省文件名 = %q", got)
	}
}
