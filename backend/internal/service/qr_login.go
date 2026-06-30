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
	eduBaseURL         = "https://kdjw.hnust.edu.cn"
	pollInterval       = 2 * time.Second
	sessionExpire      = 5 * time.Minute
	maxPollRetries     = 180
	maxSessions        = 100
	cleanupInterval    = 10 * time.Minute
)

// QRLoginService 管理所有扫码会话
type QRLoginService struct {
	sessions map[string]*model.QRSession
	mu       sync.RWMutex
	userDAO  *dao.UserDAO
	StopCh   chan struct{}
	once     sync.Once
}

var qrLoginServiceInstance *QRLoginService
var qrLoginServiceOnce sync.Once

// GetQRLoginService 获取扫码登录服务单例
func GetQRLoginService() *QRLoginService {
	qrLoginServiceOnce.Do(func() {
		qrLoginServiceInstance = &QRLoginService{
			sessions: make(map[string]*model.QRSession),
			userDAO:  dao.NewUserDAO(),
			StopCh:   make(chan struct{}),
		}
		go qrLoginServiceInstance.startCleanupTicker()
	})
	return qrLoginServiceInstance
}

// CreateSession 创建新会话并获取二维码
func (s *QRLoginService) CreateSession() (*model.QRStartResponse, error) {
	s.cleanupIfFull()

	sessionID := generateRandHex(16)
	eduUUID := generateRandHex(16)

	// 创建会话专用的 HTTP Client（独立 CookieJar）
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:     jar,
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// 步骤1：从教务网获取二维码
	qrcodeB64, err := s.fetchQRCode(client, eduUUID)
	if err != nil {
		return nil, fmt.Errorf("获取二维码失败: %w", err)
	}

	session := &model.QRSession{
		SessionID: sessionID,
		Status:    model.QRStatusPending,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(sessionExpire),
		StopCh:    make(chan struct{}),
		DoneCh:    make(chan struct{}),
	}

	s.mu.Lock()
	s.sessions[sessionID] = session
	s.mu.Unlock()

	// 后台 goroutine 轮询教务网
	go s.pollEduLoop(session, client, eduUUID)

	return &model.QRStartResponse{
		SessionID: sessionID,
		QRCode:    qrcodeB64,
		ExpiresIn: 300,
	}, nil
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
	if sess, ok := s.sessions[sessionID]; ok {
		select {
		case <-sess.StopCh:
		default:
			close(sess.StopCh)
		}
		delete(s.sessions, sessionID)
	}
}

// fetchQRCode 从教务网获取二维码图片
func (s *QRLoginService) fetchQRCode(client *http.Client, uuid string) (string, error) {
	url := fmt.Sprintf("%s/Logon.do?method=QrCodeCreate&uuid=%s", eduBaseURL, uuid)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

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

// pollEduLoop 后台持续轮询教务网扫码状态
func (s *QRLoginService) pollEduLoop(session *model.QRSession, client *http.Client, uuid string) {
	defer close(session.DoneCh)
	defer s.DeleteSession(session.SessionID)

	for i := 0; i < maxPollRetries; i++ {
		select {
		case <-session.StopCh:
			return
		case <-time.After(pollInterval):
		}

		// 检查过期
		if time.Now().After(session.ExpiresAt) {
			s.setSessionStatus(session, model.QRStatusExpired, "二维码已过期")
			return
		}

		// 轮询教务网扫码状态
		statusText, err := s.checkScanStatus(client, uuid)
		if err != nil {
			continue
		}

		statusText = strings.TrimSpace(statusText)

		switch statusText {
		case "0":
			// 等待扫码，继续轮询
		case "1":
			// 已扫码，等待确认
			s.setSessionStatus(session, model.QRStatusScanned, "已扫码，请在手机上确认登录")
		case "2", "yes":
			s.setSessionStatus(session, model.QRStatusConfirmed, "登录确认成功，正在处理...")

			// 执行登录
			ok, err := s.doLogin(client, uuid)
			if !ok || err != nil {
				errMsg := "教务网登录失败"
				if err != nil {
					errMsg = "教务网登录失败: " + err.Error()
				}
				s.setSessionStatus(session, model.QRStatusError, errMsg)
				return
			}

			// 获取用户信息（学号 + 姓名）
			userInfo, err := s.fetchUserInfo(client)
			if err != nil {
				s.setSessionStatus(session, model.QRStatusError, "获取用户信息失败: "+err.Error())
				return
			}

			session.StudentID = userInfo.StudentID
			session.Name = userInfo.Name

			// 查系统数据库中是否存在该用户
			user, err := s.userDAO.GetByStudentID(userInfo.StudentID)
			if err == nil && user != nil {
				// 用户存在，生成 JWT
				token, err := utils.GenerateToken(user.ID, user.StudentID, user.Role, user.Department, user.DeptRole)
				if err != nil {
					s.setSessionStatus(session, model.QRStatusError, "生成token失败")
					return
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
				s.setSessionStatus(session, model.QRStatusSuccess, "登录成功")

				// 延长 session 存活时间，让前端有足够时间拿到结果
				session.ExpiresAt = time.Now().Add(30 * time.Second)
			} else {
				// 用户不存在，引导注册
				s.setSessionStatus(session, model.QRStatusNeedReg, "请先注册账号")
				// 延长存活时间
				session.ExpiresAt = time.Now().Add(30 * time.Second)
			}
			return
		default:
			// 其他状态（可能是 success 或异常），继续轮询
		}
	}

	s.setSessionStatus(session, model.QRStatusExpired, "轮询超时")
}

// checkScanStatus 轮询教务网扫码状态
func (s *QRLoginService) checkScanStatus(client *http.Client, uuid string) (string, error) {
	timestamp := time.Now().UnixMilli()
	url := fmt.Sprintf("%s/Logon.do?method=checksfhd&sid=%s&_=%d", eduBaseURL, uuid, timestamp)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")

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

// doLogin 执行教务网登录（跟随302重定向）
func (s *QRLoginService) doLogin(client *http.Client, uuid string) (bool, error) {
	url := fmt.Sprintf("%s/Logon.do?method=logon_kd&type=wx&sid=%s", eduBaseURL, uuid)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")

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
		return false, fmt.Errorf("no Location header")
	}

	// 跟随第一次重定向
	if !strings.HasPrefix(location, "http") {
		location = eduBaseURL + location
	}

	req2, _ := http.NewRequest("GET", location, nil)
	req2.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	resp2, err := client.Do(req2)
	if err != nil {
		return false, err
	}
	defer resp2.Body.Close()

	// 继续跟随重定向到主页
	if resp2.StatusCode == 302 {
		homeLocation := resp2.Header.Get("Location")
		if homeLocation == "" {
			return true, nil // 登录流程完成
		}
		if !strings.HasPrefix(homeLocation, "http") {
			homeLocation = eduBaseURL + homeLocation
		}

		req3, _ := http.NewRequest("GET", homeLocation, nil)
		req3.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
		resp3, err := client.Do(req3)
		if err != nil {
			// 非致命错误，登录流程已完成
			return true, nil
		}
		if resp3 != nil {
			resp3.Body.Close()
		}
	}

	return true, nil
}

// fetchUserInfo 从教务网个人信息页提取学号和姓名
func (s *QRLoginService) fetchUserInfo(client *http.Client) (*model.EduUserInfo, error) {
	url := fmt.Sprintf("%s/jsxsd/grsz/grsz_xggrxx.do", eduBaseURL)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")

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

	// 正则提取学号（与 test_login.py 一致）
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

// setSessionStatus 线程安全地更新会话状态
func (s *QRLoginService) setSessionStatus(session *model.QRSession, status model.QRStatus, message string) {
	session.Status = status
	session.Message = message
}

// cleanupIfFull 会话数超过上限时清理最旧的
func (s *QRLoginService) cleanupIfFull() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.sessions) < maxSessions {
		return
	}

	// 找最旧的会话
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
		close(s.sessions[oldestID].StopCh)
		delete(s.sessions, oldestID)
	}
}

// startCleanupTicker 定期清理过期会话
func (s *QRLoginService) startCleanupTicker() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.StopCh:
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
			select {
			case <-sess.StopCh:
			default:
				close(sess.StopCh)
			}
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
