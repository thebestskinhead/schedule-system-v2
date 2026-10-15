package service

import (
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
)

// eduSessionTTL 教务会话在内存中的保留时长。登录成功后缓存该时长，
// 期间用户可反复发起课表导入而无需重新扫码；到期或失效后需重新登录。
const eduSessionTTL = 2 * time.Hour

// eduCookie 序列化后的单条 Cookie
type eduCookie struct {
	Name     string
	Value    string
	Domain   string
	Path     string
	Secure   bool
	HttpOnly bool
	Expires  time.Time
}

// eduSession 缓存的教务登录会话
type eduSession struct {
	StudentID string
	Cookies   []eduCookie
	ExpiresAt time.Time
}

// EduSessionCache 按学号缓存教务登录会话（进程内存，带 TTL）。
// 课表导入服务据此复用扫码登录得到的 Cookie，无需用户重复扫码。
type EduSessionCache struct {
	mu       sync.RWMutex
	sessions map[string]*eduSession // key: 学号
}

var (
	eduCacheInstance *EduSessionCache
	eduCacheOnce     sync.Once
)

// GetEduSessionCache 获取教务会话缓存单例
func GetEduSessionCache() *EduSessionCache {
	eduCacheOnce.Do(func() {
		eduCacheInstance = &EduSessionCache{
			sessions: make(map[string]*eduSession),
		}
	})
	return eduCacheInstance
}

// Put 缓存指定学号的教务会话，从 CookieJar 中抽取教务站 Cookie。
//
// 教务的 Cookie 分布在多个路径下，而 Cookie 是否随请求发送由其 Path 决定。
// 这里按课表导入实际访问的 URL 逐一遍历汇总，避免遗漏深层路径下的鉴权 Cookie。
func (c *EduSessionCache) Put(studentID string, jar *cookiejar.Jar) {
	if studentID == "" || jar == nil {
		return
	}

	cookies := collectEduCookies(jar)
	if len(cookies) == 0 {
		log.Printf("[edu-session] 学号 %s 的教务会话未发现任何 Cookie，跳过缓存", studentID)
		return
	}

	serialized := make([]eduCookie, 0, len(cookies))
	for _, ck := range cookies {
		serialized = append(serialized, eduCookie{
			Name:     ck.Name,
			Value:    ck.Value,
			Domain:   ck.Domain,
			Path:     ck.Path,
			Secure:   ck.Secure,
			HttpOnly: ck.HttpOnly,
			Expires:  ck.Expires,
		})
	}

	c.mu.Lock()
	c.sessions[studentID] = &eduSession{
		StudentID: studentID,
		Cookies:   serialized,
		ExpiresAt: time.Now().Add(eduSessionTTL),
	}
	c.mu.Unlock()

	names := make([]string, 0, len(serialized))
	for _, ck := range serialized {
		names = append(names, ck.Name)
	}
	log.Printf("[edu-session] 已缓存学号 %s 的教务会话，共 %d 枚 Cookie: %s", studentID, len(serialized), strings.Join(names, ","))
}

// collectEduCookies 汇总教务各业务页面会携带的 Cookie 并按 名称+域+路径 去重。
// 仅取根路径会漏掉深层路径下发的鉴权 Cookie，故按实际访问地址逐一遍历。
func collectEduCookies(jar *cookiejar.Jar) []*http.Cookie {
	seen := make(map[string]bool)
	var out []*http.Cookie
	for _, raw := range []string{eduBaseURL, eduTimetableListURL, eduTimetablePrintURL} {
		u, err := url.Parse(raw)
		if err != nil {
			continue
		}
		for _, ck := range jar.Cookies(u) {
			key := ck.Name + "\x00" + ck.Domain + "\x00" + ck.Path
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, ck)
		}
	}
	return out
}

// Get 取出指定学号的可用会话，返回重建的 CookieJar。未命中或已过期返回 nil。
func (c *EduSessionCache) Get(studentID string) *cookiejar.Jar {
	c.mu.RLock()
	sess := c.sessions[studentID]
	c.mu.RUnlock()

	if sess == nil {
		return nil
	}
	if time.Now().After(sess.ExpiresAt) {
		c.Delete(studentID)
		return nil
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil
	}

	eduURL, _ := url.Parse(eduBaseURL)
	for _, ck := range sess.Cookies {
		jar.SetCookies(eduURL, []*http.Cookie{{
			Name:     ck.Name,
			Value:    ck.Value,
			Path:     ck.Path,
			Domain:   ck.Domain,
			Secure:   ck.Secure,
			HttpOnly: ck.HttpOnly,
			Expires:  ck.Expires,
		}})
	}
	return jar
}

// Delete 移除指定学号的缓存会话
func (c *EduSessionCache) Delete(studentID string) {
	c.mu.Lock()
	delete(c.sessions, studentID)
	c.mu.Unlock()
}

// cleanup 清理过期会话，由定时器驱动
func (c *EduSessionCache) cleanup() {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	for id, sess := range c.sessions {
		if now.After(sess.ExpiresAt) {
			delete(c.sessions, id)
		}
	}
}
