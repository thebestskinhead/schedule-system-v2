package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// doTestEmail 调用 TestEmail 并返回响应信封中的 code 与 message
func doTestEmail(t *testing.T, body string) (int, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/admin/smtp/test", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	NewSMTPHandler().TestEmail(c)

	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("响应解析失败: %v (body=%s)", err, w.Body.String())
	}
	return resp.Code, resp.Message
}

func TestTestEmailValidation(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		wantMsg string
	}{
		{"缺少收件人", `{}`, "请填写收件邮箱"},
		{"收件人格式错误", `{"to":"abc"}`, "收件邮箱格式不正确"},
		{"收件人带空格被去除后进入配置校验", `{"to":" a@qq.com ","host":"smtp.qq.com"}`, "SMTP 配置不完整，请填写服务器、用户名与发件人邮箱"},
		{"内联配置缺少用户名", `{"to":"a@qq.com","host":"smtp.qq.com","from_email":"a@qq.com"}`, "SMTP 配置不完整，请填写服务器、用户名与发件人邮箱"},
		{"内联配置发件人邮箱非法", `{"to":"a@qq.com","host":"smtp.qq.com","username":"u","from_email":"bad"}`, "发件人邮箱格式不正确"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, msg := doTestEmail(t, tc.body)
			if code != 400 {
				t.Fatalf("code = %d, want 400 (msg=%s)", code, msg)
			}
			if msg != tc.wantMsg {
				t.Fatalf("msg = %q, want %q", msg, tc.wantMsg)
			}
		})
	}
}
