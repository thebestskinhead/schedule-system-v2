package utils

import "testing"

func TestIsEmail(t *testing.T) {
	valid := []string{
		"a@qq.com",
		"a@qq.com.cn",
		"li.si+tag@example.org",
		"user_1@mail.example.com",
	}
	for _, v := range valid {
		if !IsEmail(v) {
			t.Errorf("IsEmail(%q) = false, 期望 true", v)
		}
	}

	invalid := []string{
		"",
		"abc",
		"张三",
		"a@qq",
		"a@qq.com ",
		" a@qq.com",
		"a b@qq.com",
		"a@@qq.com",
	}
	for _, v := range invalid {
		if IsEmail(v) {
			t.Errorf("IsEmail(%q) = true, 期望 false", v)
		}
	}
}

func TestNormalizeEmail(t *testing.T) {
	cases := map[string]string{
		"  a@qq.com  ": "a@qq.com",
		"\ta@qq.com\n": "a@qq.com",
		"a@qq.com":     "a@qq.com",
	}
	for in, want := range cases {
		if got := NormalizeEmail(in); got != want {
			t.Errorf("NormalizeEmail(%q) = %q, 期望 %q", in, got, want)
		}
	}
}

func TestNormalizeEmailEnablesValidation(t *testing.T) {
	// 带首尾空格的邮箱：规范化后应通过校验
	raw := "  a@qq.com "
	if !IsEmail(NormalizeEmail(raw)) {
		t.Errorf("规范化后仍不通过校验: %q", raw)
	}
	if IsEmail(raw) {
		t.Errorf("未规范化时应不通过校验: %q", raw)
	}
}
