package service

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"
)

// 教务在深层路径下发的 Cookie 同样参与鉴权，缓存时应完整保留，
// 否则重建会话后下载课表会缺少凭证而被判定为登录失效。
func TestEduSessionCacheKeepsDeepPathCookies(t *testing.T) {
	jar, _ := cookiejar.New(nil)
	loginURL, _ := url.Parse(eduBaseURL + "/jsxsd/framework/xsMain.jsp")
	jar.SetCookies(loginURL, []*http.Cookie{
		{Name: "bzb_jsxsd", Value: "AAA", Path: "/jsxsd"},
		{Name: "bzb_njw", Value: "BBB", Path: "/jsxsd"},
		{Name: "SERVERID", Value: "CCC", Path: "/"},
	})

	cache := &EduSessionCache{sessions: make(map[string]*eduSession)}
	cache.Put("2021001", jar)

	restored := cache.Get("2021001")
	if restored == nil {
		t.Fatal("会话应已缓存")
	}

	listURL, _ := url.Parse(eduTimetableListURL)
	got := map[string]string{}
	for _, ck := range restored.Cookies(listURL) {
		got[ck.Name] = ck.Value
	}
	if len(got) == 0 {
		t.Fatal("重建后的会话在请求课表页时未携带任何 Cookie")
	}
	for _, want := range []string{"bzb_jsxsd", "bzb_njw", "SERVERID"} {
		if got[want] == "" {
			t.Errorf("重建后的会话在请求课表页时缺少 Cookie %s（实得 %v）", want, got)
		}
	}
}

// 未持有任何 Cookie 的会话不应写入缓存。
func TestEduSessionCacheSkipsEmptyJar(t *testing.T) {
	jar, _ := cookiejar.New(nil)

	cache := &EduSessionCache{sessions: make(map[string]*eduSession)}
	cache.Put("2021002", jar)

	if cache.Get("2021002") != nil {
		t.Error("无 Cookie 的会话不应被缓存")
	}
}
