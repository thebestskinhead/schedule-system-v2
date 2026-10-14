//go:build integration

package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"

	"schedule-system-v2/backend/internal/auth"
	"schedule-system-v2/backend/internal/config"
	"schedule-system-v2/backend/internal/db"
)

// checkerHasManageAll 直接构造 Checker 校验临时权限是否被正确加载。
func checkerHasManageAll(userID int) bool {
	c := auth.NewChecker(userID, "", "user", "竞赛部", "dept_member")
	return c.HasPermission(auth.PermUserManageAll)
}

// 端到端验证 P0-4：持临时 user:manage:all 的普通成员不得把自身身份升级为超出自身范围的身份。
// 需真实数据库；以 build tag integration 隔离：
//
//	go test -tags integration ./internal/handler/ -run TestP04 -v
//
// 自建临时用户与权限，测试结束即清理，不留残留。
func TestP04_IdentityEscalationViaTempPermissionBlocked(t *testing.T) {
	cfg, err := config.LoadConfig("../../configs/config.yaml")
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	if err := db.InitDB(&cfg.Database); err != nil {
		t.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	const sid = "TESTP04ESC"
	// 清理历史残留
	_, _ = db.GetDB().Exec(`DELETE FROM user_permissions_temp WHERE user_id IN (SELECT id FROM users WHERE student_id = ?)`, sid)
	_, _ = db.GetDB().Exec(`DELETE FROM users WHERE student_id = ?`, sid)

	res, err := db.GetDB().Exec(
		`INSERT INTO users (student_id, name, email, password, role, department, dept_role, employment_status)
		 VALUES (?, ?, ?, ?, 'user', '竞赛部', 'dept_member', 'active')`,
		sid, "P04越权测试", "p04esc@test.local", "x")
	if err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}
	uid64, _ := res.LastInsertId()
	uid := int(uid64)

	// 授予临时 user:manage:all（唯一能越过路由闸门又有增益的身份，见报告 P0-4）
	_, err = db.GetDB().Exec(
		`INSERT INTO user_permissions_temp (user_id, permission, resource_type, resource_id, granted_by, reason, expires_at, is_active)
		 VALUES (?, 'user:manage:all', 'all', 0, ?, '', NULL, 1)`,
		uid, uid)
	if err != nil {
		t.Fatalf("授予临时权限失败: %v", err)
	}

	// 前置断言：临时权限必须已被加载（否则第一道权限闸门就会拦下，测不到边界逻辑）
	if !checkerHasManageAll(uid) {
		t.Fatal("临时权限未生效：权限加载失败，测试前提不成立")
	}

	defer func() {
		_, _ = db.GetDB().Exec(`DELETE FROM user_permissions_temp WHERE user_id = ?`, uid)
		_, _ = db.GetDB().Exec(`DELETE FROM users WHERE id = ?`, uid)
	}()

	gin.SetMode(gin.TestMode)
	h := NewUserHandler()

	newCtx := func(path, body string) (*gin.Context, *httptest.ResponseRecorder) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, path, bytes.NewBufferString(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(uid)}}
		c.Set("userID", uid)
		c.Set("studentID", sid)
		c.Set("role", "user")
		c.Set("department", "竞赛部")
		c.Set("deptRole", "dept_member")
		return c, w
	}

	// 1) 把自己改成「办公室」+ dept_admin → 应 403
	c1, w1 := newCtx("/api/v1/admin/users/"+strconv.Itoa(uid)+"/department", `{"department":"办公室"}`)
	h.SetUserDepartment(c1)
	if w1.Code != http.StatusForbidden {
		t.Fatalf("SetUserDepartment 越界未被拦截，got %d body=%s", w1.Code, w1.Body.String())
	}
	t.Logf("SetUserDepartment 越界拦截 OK: %d %s", w1.Code, w1.Body.String())

	// 2) 把自己设为 dept_admin → 应 403
	c2, w2 := newCtx("/api/v1/admin/users/"+strconv.Itoa(uid)+"/dept-role", `{"dept_role":"dept_admin"}`)
	h.SetUserDeptRole(c2)
	if w2.Code != http.StatusForbidden {
		t.Fatalf("SetUserDeptRole 越界未被拦截，got %d body=%s", w2.Code, w2.Body.String())
	}
	t.Logf("SetUserDeptRole 越界拦截 OK: %d %s", w2.Code, w2.Body.String())

	// 3) 正向对照：改为同部门内的普通成员身份（不扩大权限）→ 应放行
	c3, w3 := newCtx("/api/v1/admin/users/"+strconv.Itoa(uid)+"/dept-role", `{"dept_role":"dept_member"}`)
	h.SetUserDeptRole(c3)
	if w3.Code == http.StatusForbidden {
		t.Fatalf("同级身份变更被误拦，body=%s", w3.Body.String())
	}
	t.Logf("同级身份变更放行 OK: %d %s", w3.Code, w3.Body.String())
}
