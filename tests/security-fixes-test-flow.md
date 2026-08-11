# 权限系统与 JWT 鉴权链路修复 —— 完整测试流程

> 适用对象：排班系统 v2 后端（Go 1.24.0，D:\usexxx 提供）
> 测试目标：验证 2026-08-12 三项安全修复
>   修复1：Login / QR 登录显式校验 is_active（禁用账号不可登录）
>   修复2：AuthMiddleware 验签后从数据库重载用户，权限判定不信任 JWT 快照
>   修复3：ParseToken 固定 HS256 算法（拒绝算法混淆）

---

## 0. 环境信息（已从 backend/configs/config.yaml 读取）

| 项 | 值 |
|---|---|
| 后端地址 | http://localhost:8080 |
| API 前缀 | /api/v1 |
| MySQL | localhost:3306，库 `schedule_test`，用户 `schedule_test`，密码 `Schedule2026!` |
| JWT secret | dev-test-secret-key-change-me（dev 环境，过期 168h） |
| 管理员账号 | 学号 `2505050318`，密码 `123456` |

MySQL 命令行（Git Bash）：

```bash
mysql -h localhost -u schedule_test -p'Schedule2026!' schedule_test
```

---

## 1. 启动后端

```bash
cd /e/workspace/schedule-system-v2/backend
export PATH="/d/usexxx/go/versions/1.24.0/go/bin:$PATH"
export GOPROXY="https://goproxy.cn,direct"
go build -o server-test.exe ./cmd/server
./server-test.exe
# 预期日志：系统已安装，加载配置... 服务器启动，监听端口: 8080
# 测试期间保持该窗口运行，另开 Git Bash 执行下列命令
```

---

## 2. 准备：管理员登录，取 TOKEN

```bash
curl -s -X POST http://localhost:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"student_id":"2505050318","password":"123456"}'
```

- **预期**：`"code":0`（或 200），返回体含 `token` 字段
- 将 token 存入变量：

```bash
ADMIN_TOKEN=<上一步返回的token>
```

验证 token 可用：

```bash
curl -s http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer $ADMIN_TOKEN"
# 预期：返回 2505050318 的用户信息，role=admin
```

---

## 3. 准备：创建测试用户 B（普通成员）

```bash
curl -s -X POST http://localhost:8080/api/v1/admin/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"student_id":"test_b_001","name":"测试用户B","email":"testb@example.com","password":"123456","role":"user","department":"竞赛部","dept_role":"dept_member"}'
# 预期：创建成功，记录返回的用户 id（下文记为 B_ID）
```

用户 B 登录取 token：

```bash
curl -s -X POST http://localhost:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"student_id":"test_b_001","password":"123456"}'
B_TOKEN=<返回的token>
```

---

## 4. 测试 A —— 修复2：禁用账号后，已签发的 token 立即失效

**目的**：修复前，禁用用户后其 7 天内已签发的 token 仍可用；修复后应立即 401。

```bash
# 4.1 禁用前：B 的 token 可正常访问
curl -s http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer $B_TOKEN"
# 预期：200，返回 B 的信息
```

```bash
# 4.2 管理员禁用 B（软删除 is_active=0）
curl -s -X DELETE http://localhost:8080/api/v1/admin/users/$B_ID \
  -H "Authorization: Bearer $ADMIN_TOKEN"
# 预期：成功
```

```bash
# 4.3 禁用后：B 的旧 token 立即被拒
curl -s http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer $B_TOKEN"
# ✅ 预期：401 {"code":401,"message":"用户不存在或已被禁用"}
# ❌ 修复前表现：仍返回 200 用户信息
```

---

## 5. 测试 B —— 修复1：禁用账号无法登录

```bash
curl -s -X POST http://localhost:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"student_id":"test_b_001","password":"123456"}'
# ✅ 预期：401/400，message 为 "账号已被禁用，请联系管理员"（或 "用户不存在"）
# ❌ 若返回 token 则修复失效
```

---

## 6. 测试 C —— 修复2：降权即时生效（不重新登录）

**目的**：权限以数据库为准，角色变更后旧 token 权限立即变化。

```bash
# 6.1 恢复 B（SQL 重新激活）
mysql -h localhost -u schedule_test -p'Schedule2026!' schedule_test \
  -e "UPDATE users SET is_active=1 WHERE student_id='test_b_001';"
```

```bash
# 6.2 B 重新登录取新 token，并提升 B 为部门管理员
curl -s -X POST http://localhost:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"student_id":"test_b_001","password":"123456"}'
B_TOKEN2=<返回的token>

curl -s -X PUT http://localhost:8080/api/v1/admin/users/$B_ID/dept-role \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"dept_role":"dept_admin"}'

# 6.3 B 的旧 token（B_TOKEN2 签发时 dept_role=dept_member）
#     现在应已拥有部门管理员权限（数据库已变更，无需重新登录）
curl -s "http://localhost:8080/api/v1/admin/users/by-dept?department=竞赛部" \
  -H "Authorization: Bearer $B_TOKEN2"
# ✅ 预期：200，返回竞赛部用户列表
# ❌ 修复前表现：403（token 快照仍是 dept_member）

# 6.4 反向验证：把 B 降回成员，旧 token 立即失去权限
curl -s -X PUT http://localhost:8080/api/v1/admin/users/$B_ID/dept-role \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"dept_role":"dept_member"}'

curl -s "http://localhost:8080/api/v1/admin/users/by-dept?department=竞赛部" \
  -H "Authorization: Bearer $B_TOKEN2"
# ✅ 预期：403 无权访问
```

> 同理可验证 role 降级：`PUT /api/v1/admin/users/:id/role {"role":"user"}` 把某管理员降为普通用户后，其旧 token 访问 `/api/v1/admin/smtp/configs` 应立即 403。

---

## 7. 测试 D —— 修复3：算法混淆攻击被拒绝

单元测试已覆盖（`backend/internal/utils/jwt_security_smoke_test.go`，全部 PASS）。
黑盒验证：用 Go 一次性脚本以 HS512 伪造 token，请求受保护接口：

```bash
cd /e/workspace/schedule-system-v2/backend
cat > /tmp/forge.go << 'EOF'
package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	claims := jwt.MapClaims{
		"user_id": 1, "student_id": "2505050318", "role": "admin",
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	s, _ := t.SignedString([]byte("dev-test-secret-key-change-me"))
	fmt.Println(s)
}
EOF
export PATH="/d/usexxx/go/versions/1.24.0/go/bin:$PATH"
FORGED=$(go run /tmp/forge.go)

curl -s http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer $FORGED"
# ✅ 预期：401 {"code":401,"message":"无效的token"}
# ❌ 若返回用户信息则算法固定未生效
```

错误密钥伪造（对照组）：

```bash
# 将 forge.go 中 secret 改为 "wrong-secret" 再跑一次，同样应 401
```

---

## 8. 回归测试 —— 正常功能不受影响

```bash
# 8.1 管理员正常功能
curl -s http://localhost:8080/api/v1/admin/users \
  -H "Authorization: Bearer $ADMIN_TOKEN"
# 预期：200 用户列表

curl -s http://localhost:8080/api/v1/departments \
  -H "Authorization: Bearer $ADMIN_TOKEN"
# 预期：200 部门列表

# 8.2 普通用户 B 正常功能
curl -s http://localhost:8080/api/v1/schedule \
  -H "Authorization: Bearer $B_TOKEN2"
# 预期：200 排班数据（普通查看权限正常）

# 8.3 无 token / 畸形 token
curl -s http://localhost:8080/api/v1/user/profile
# 预期：401 未提供认证信息
curl -s http://localhost:8080/api/v1/user/profile -H "Authorization: Bearer abc.def.ghi"
# 预期：401 无效的token
```

---

## 9. 清理测试数据

```bash
mysql -h localhost -u schedule_test -p'Schedule2026!' schedule_test \
  -e "DELETE FROM users WHERE student_id='test_b_001';"
# 停掉后端：在运行 server-test.exe 的窗口按 Ctrl+C，删除 server-test.exe
```

---

## 10. 验收清单

| # | 用例 | 修复前 | 修复后预期 |
|---|---|---|---|
| A | 禁用后旧 token 访问 | 200（7 天残留） | 401 立即失效 |
| B | 禁用账号登录 | 可登录/报错含混 | 明确拒绝"账号已被禁用" |
| C1 | 升职后旧 token | 403（快照不更新） | 立即拥有新权限 |
| C2 | 降职后旧 token | 保留旧权限 | 立即 403 |
| D | HS512/none/错密钥 token | HS512 可能被接受 | 全部 401 |
| E | 正常登录/查看/管理 | 正常 | 正常（回归） |
