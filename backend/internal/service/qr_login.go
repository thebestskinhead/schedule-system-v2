package service

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strings"
	"sync"
	"time"

	"schedule-system-v2/backend/internal/dao"
	"schedule-system-v2/backend/internal/model"
	"schedule-system-v2/backend/internal/utils"
)

const (
	eduBaseURL      = "https://kdjw.hnust.edu.cn"
	sessionExpire   = 5 * time.Minute
	maxSessions     = 100
	cleanupInterval = 10 * time.Minute
)

// QRLoginService 管理所有扫码会话（前端驱动轮询模式）
type QRLoginService struct {
	sessions map[string]*model.QRSession
	mu       sync.RWMutex
	userDAO  *dao.UserDAO
	stopCh   chan struct{}
}

var qrLoginServiceInstance *QRLoginService
var qrLoginServiceOnce sync.Once

// GetQRLoginService 获取扫码登录服务单例
func GetQRLoginService() *QRLoginService {
	qrLoginServiceOnce.Do(func() {
		qrLoginServiceInstance = &QRLoginService{
			sessions: make(map[string]*model.QRSession),
			userDAO:  dao.NewUserDAO(),
			stopCh:   make(chan struct{}),
		}
		go qrLoginServiceInstance.startCleanupTicker()
	})
	return qrLoginServiceInstance
}

// commonHeaders 与 Python test_login.py 一致的请求头
func commonHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
}

// CreateSession 创建新会话：向教务网获取二维码，保存 CookieJar 和 UUID
func (s *QRLoginService) CreateSession() (*model.QRStartResponse, error) {
	s.cleanupIfFull()

	sessionID := generateRandHex(16)
	eduUUID := generateRandHex(16)

	// 创建独立 CookieJar 存储教务网 Cookie
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("创建CookieJar失败: %w", err)
	}

	// 请求教务网二维码（跟随重定向，与 Python 一致）
	qrcodeB64, err := s.fetchQRCode(jar, eduUUID)
	if err != nil {
		return nil, fmt.Errorf("获取二维码失败: %w", err)
	}

	session := &model.QRSession{
		SessionID: sessionID,
		UUID:      eduUUID,
		CookieJar: jar,
		Status:    model.QRStatusPending,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(sessionExpire),
	}

	s.mu.Lock()
	s.sessions[sessionID] = session
	s.mu.Unlock()

	return &model.QRStartResponse{
		SessionID: sessionID,
		QRCode:    qrcodeB64,
		ExpiresIn: 300,
	}, nil
}

// Poll 前端触发一次轮询：代理请求教务网，返回当前状态
func (s *QRLoginService) Poll(sessionID string) *model.QRPollResponse {
	session := s.GetSession(sessionID)
	if session == nil {
		return &model.QRPollResponse{
			Status:  model.QRStatusExpired,
			Message: "二维码已过期，请重新获取",
		}
	}

	// 检查会话超时
	if time.Now().After(session.ExpiresAt) {
		s.DeleteSession(sessionID)
		return &model.QRPollResponse{
			Status:  model.QRStatusExpired,
			Message: "二维码已过期，请重新获取",
		}
	}

	// 已完成的会话直接返回结果
	if session.IsDone() {
		status, msg := session.GetStatus()
		switch status {
		case model.QRStatusSuccess:
			return &model.QRPollResponse{
				Status: model.QRStatusSuccess,
				Token:  session.Token,
				User:   session.User,
			}
		case model.QRStatusNeedReg:
			return &model.QRPollResponse{
				Status:    model.QRStatusNeedReg,
				StudentID: session.StudentID,
				Name:      session.Name,
			}
		case model.QRStatusError:
			return &model.QRPollResponse{
				Status:  model.QRStatusError,
				Message: msg,
			}
		}
	}

	// 防止并发请求同时访问教务网
	if !session.Processing.TryLock() {
		status, msg := session.GetStatus()
		return &model.QRPollResponse{Status: status, Message: msg}
	}
	defer session.Processing.Unlock()

	// 双重检查
	if session.IsDone() {
		status, msg := session.GetStatus()
		switch status {
		case model.QRStatusSuccess:
			s.DeleteSession(sessionID)
			return &model.QRPollResponse{Status: model.QRStatusSuccess, Token: session.Token, User: session.User}
		case model.QRStatusNeedReg:
			s.DeleteSession(sessionID)
			return &model.QRPollResponse{Status: model.QRStatusNeedReg, StudentID: session.StudentID, Name: session.Name}
		case model.QRStatusError:
			return &model.QRPollResponse{Status: model.QRStatusError, Message: msg}
		}
	}

	// 代理请求教务网 checksfhd（跟随重定向，与 Python 一致）
	statusText, err := s.checkScanStatus(session.CookieJar, session.UUID)
	if err != nil {
		// 网络错误不改变状态，继续返回 pending
		return &model.QRPollResponse{
			Status:  model.QRStatusPending,
			Message: "等待扫码",
		}
	}

	statusText = strings.TrimSpace(statusText)

	switch statusText {
	case "0":
		// 等待扫码
		return &model.QRPollResponse{
			Status:  model.QRStatusPending,
			Message: "等待扫码",
		}

	case "1":
		// 已扫码等待确认
		session.SetStatus(model.QRStatusScanned, "已扫码，请在手机上确认登录")
		return &model.QRPollResponse{
			Status:  model.QRStatusScanned,
			Message: "已扫码，请在手机上确认登录",
		}

	case "2", "yes":
		// 用户已确认，执行登录流程
		session.SetStatus(model.QRStatusConfirmed, "登录确认成功，正在处理...")

		// 执行教务网登录（需要阻止重定向，与 Python allow_redirects=False 一致）
		ok, err := s.doLogin(session.CookieJar, session.UUID)
		if !ok || err != nil {
			errMsg := "教务网登录失败"
			if err != nil {
				errMsg = "教务网登录失败: " + err.Error()
			}
			session.SetResult(model.QRStatusError, errMsg)
			// 延长存活，让前端能读到错误状态
			session.ExpiresAt = time.Now().Add(30 * time.Second)
			return &model.QRPollResponse{
				Status:  model.QRStatusError,
				Message: errMsg,
			}
		}

		// 获取用户信息（跟随重定向，与 Python 一致）
		userInfo, err := s.fetchUserInfo(session.CookieJar)
		if err != nil {
			session.SetResult(model.QRStatusError, "获取用户信息失败: "+err.Error())
			session.ExpiresAt = time.Now().Add(30 * time.Second)
			return &model.QRPollResponse{
				Status:  model.QRStatusError,
				Message: "获取用户信息失败: " + err.Error(),
			}
		}

		session.StudentID = userInfo.StudentID
		session.Name = userInfo.Name

		// 查系统数据库
		user, err := s.userDAO.GetByStudentID(userInfo.StudentID)
		if err == nil && user != nil {
			// 已注册，生成 JWT
			token, err := utils.GenerateToken(user.ID, user.StudentID, user.Role, user.Department, user.DeptRole)
			if err != nil {
				session.SetResult(model.QRStatusError, "生成token失败")
				session.ExpiresAt = time.Now().Add(30 * time.Second)
				return &model.QRPollResponse{
					Status:  model.QRStatusError,
					Message: "生成token失败",
				}
			}
			session.Token = token
			session.User = model.UserInfo{
				ID:         user.ID,
				StudentID:  user.StudentID,
				Name:       user.Name,
				Email:      user.Email,
				Role:       user.Role,
				Department: user.Department,
				DeptRole:   user.DeptRole,
			}
			session.SetResult(model.QRStatusSuccess, "登录成功")
			session.ExpiresAt = time.Now().Add(30 * time.Second)
			return &model.QRPollResponse{
				Status: model.QRStatusSuccess,
				Token:  session.Token,
				User:   session.User,
			}
		}

		// 未注册
		session.SetResult(model.QRStatusNeedReg, "请先注册账号")
		session.ExpiresAt = time.Now().Add(30 * time.Second)
		return &model.QRPollResponse{
			Status:    model.QRStatusNeedReg,
			StudentID: session.StudentID,
			Name:      session.Name,
		}

	default:
		// 包含 "success" 字符串的处理（与 Python 一致）
		if strings.Contains(strings.ToLower(statusText), "success") {
			session.SetResult(model.QRStatusError, "教务网返回异常状态")
			session.ExpiresAt = time.Now().Add(30 * time.Second)
			return &model.QRPollResponse{
				Status:  model.QRStatusError,
				Message: "教务网返回异常状态: " + statusText,
			}
		}
		// 未知状态，继续等待
		return &model.QRPollResponse{
			Status:  model.QRStatusPending,
			Message: "等待扫码",
		}
	}
}

// GetSession 获取会话
func (s *QRLoginService) GetSession(sessionID string) *model.QRSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[sessionID]
}

// DeleteSession 删除会话
func (s *QRLoginService) DeleteSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
}

// ---------- 教务网 HTTP 请求 ----------

// newRedirectClient 创建跟随重定向的 HTTP Client（用于 QrCodeCreate / checksfhd / fetchUserInfo）
func newRedirectClient(jar *cookiejar.Jar) *http.Client {
	return &http.Client{
		Jar:     jar,
		Timeout: 15 * time.Second,
	}
}

// newNoRedirectClient 创建不跟随重定向的 HTTP Client（用于 logon_kd）
func newNoRedirectClient(jar *cookiejar.Jar) *http.Client {
	return &http.Client{
		Jar:     jar,
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// fetchQRCode 从教务网获取二维码图片（跟随重定向，与 Python create_qrcode 一致）
func (s *QRLoginService) fetchQRCode(jar *cookiejar.Jar, uuid string) (string, error) {
	url := fmt.Sprintf("%s/Logon.do?method=QrCodeCreate&uuid=%s", eduBaseURL, uuid)
	client := newRedirectClient(jar)

	req, _ := http.NewRequest("GET", url, nil)
	commonHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	b64 := base64.StdEncoding.EncodeToString(data)
	return "data:image/jpeg;base64," + b64, nil
}

// checkScanStatus 轮询教务网扫码状态（跟随重定向，与 Python poll_status 一致）
func (s *QRLoginService) checkScanStatus(jar *cookiejar.Jar, uuid string) (string, error) {
	timestamp := time.Now().UnixMilli()
	url := fmt.Sprintf("%s/Logon.do?method=checksfhd&sid=%s&_=%d", eduBaseURL, uuid, timestamp)
	client := newRedirectClient(jar)

	req, _ := http.NewRequest("GET", url, nil)
	commonHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// doLogin 执行教务网登录，手动跟随 302 重定向（与 Python do_login 一致）
func (s *QRLoginService) doLogin(jar *cookiejar.Jar, uuid string) (bool, error) {
	url := fmt.Sprintf("%s/Logon.do?method=logon_kd&type=wx&sid=%s", eduBaseURL, uuid)
	client := newNoRedirectClient(jar) // ← 与 Python allow_redirects=False 一致

	req, _ := http.NewRequest("GET", url, nil)
	commonHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 302 {
		return false, fmt.Errorf("expected 302, got %d", resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return false, fmt.Errorf("no Location header in 302 response")
	}

	// 跟随第一次重定向（跟随重定向，获取第三枚 Cookie）
	if !strings.HasPrefix(location, "http") {
		location = eduBaseURL + location
	}

	client2 := newRedirectClient(jar) // 跟随重定向
	req2, _ := http.NewRequest("GET", location, nil)
	commonHeaders(req2)
	resp2, err := client2.Do(req2)
	if err != nil {
		return false, err
	}
	defer resp2.Body.Close()

	// 继续跟随到主页
	if resp2.StatusCode == 302 {
		homeLocation := resp2.Header.Get("Location")
		if homeLocation == "" {
			return true, nil
		}
		if !strings.HasPrefix(homeLocation, "http") {
			homeLocation = eduBaseURL + homeLocation
		}
		req3, _ := http.NewRequest("GET", homeLocation, nil)
		commonHeaders(req3)
		resp3, _ := client2.Do(req3)
		if resp3 != nil {
			resp3.Body.Close()
		}
	}

	return true, nil
}

// fetchUserInfo 从教务网个人信息页提取学号和姓名（跟随重定向，与 Python get_user_info 一致）
func (s *QRLoginService) fetchUserInfo(jar *cookiejar.Jar) (*model.EduUserInfo, error) {
	url := fmt.Sprintf("%s/jsxsd/grsz/grsz_xggrxx.do", eduBaseURL)
	client := newRedirectClient(jar)

	req, _ := http.NewRequest("GET", url, nil)
	commonHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	html, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	text := string(html)

	// 正则提取学号和姓名（与 Python 一致）
	studentIDRe := regexp.MustCompile(`name="account"[^>]*value="([^"]*)"`)
	nameRe := regexp.MustCompile(`name="realName"[^>]*value="([^"]*)"`)

	studentIDMatch := studentIDRe.FindStringSubmatch(text)
	nameMatch := nameRe.FindStringSubmatch(text)

	studentID := ""
	if len(studentIDMatch) > 1 {
		studentID = studentIDMatch[1]
	}

	name := ""
	if len(nameMatch) > 1 {
		name = nameMatch[1]
	}

	if studentID == "" {
		return nil, fmt.Errorf("未能从教务网提取学号")
	}

	return &model.EduUserInfo{
		StudentID: studentID,
		Name:      name,
	}, nil
}

// ---------- 会话管理 ----------

// cleanupIfFull 会话数超过上限时清理最旧的
func (s *QRLoginService) cleanupIfFull() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.sessions) < maxSessions {
		return
	}

	var oldestID string
	var oldestTime time.Time
	first := true
	for id, sess := range s.sessions {
		if first || sess.CreatedAt.Before(oldestTime) {
			oldestID = id
			oldestTime = sess.CreatedAt
			first = false
		}
	}
	if oldestID != "" {
		delete(s.sessions, oldestID)
	}
}

// startCleanupTicker 定期清理过期会话
func (s *QRLoginService) startCleanupTicker() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.cleanupExpired()
		}
	}
}

// cleanupExpired 清理所有过期会话
func (s *QRLoginService) cleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, sess := range s.sessions {
		if now.After(sess.ExpiresAt) {
			delete(s.sessions, id)
		}
	}
}

// generateRandHex 生成随机十六进制字符串
func generateRandHex(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}
