//go:build integration

package service

import (
	"testing"

	"schedule-system-v2/backend/internal/config"
	"schedule-system-v2/backend/internal/db"
	"schedule-system-v2/backend/internal/model"
)

// 排班核心算法（PreviewSchedule）集成测试。需真实数据库，以 build tag integration 隔离：
//
//	go test -tags integration ./internal/service/ -run TestSchedulePreview -v
//
// 使用独立测试部门，测试用户与其无课/值班数据级联清理，不留残留。

const (
	schedTestDept  = "排班测试部"
	schedTestWeek  = 29
	schedHistWeek  = 28 // 历史班次所在周，与预览周错开
	schedStudentID = "TESTSCHED"
)

// setupScheduleTestDB 幂等建表 + 清理本测试的历史残留，返回 ScheduleService
func setupScheduleTestDB(t *testing.T) *ScheduleService {
	t.Helper()

	cfg, err := config.LoadConfig("../../configs/config.yaml")
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	// 幂等建表（force=false 不会删任何既有表）
	if err := NewSystemService().InitDatabaseTables(&cfg.Database, false); err != nil {
		t.Fatalf("初始化数据库表失败: %v", err)
	}

	// 清理历史残留（users 级联删除 availability / duty_records）
	_, _ = db.GetDB().Exec(`DELETE FROM users WHERE department = ?`, schedTestDept)

	t.Cleanup(func() {
		_, _ = db.GetDB().Exec(`DELETE FROM users WHERE department = ?`, schedTestDept)
		db.Close()
	})

	return NewScheduleService()
}

func createSchedTestUser(t *testing.T, sid, name string) int {
	t.Helper()
	res, err := db.GetDB().Exec(
		`INSERT INTO users (student_id, name, email, password, role, department, dept_role, employment_status)
		 VALUES (?, ?, ?, 'x', 'user', ?, 'dept_member', 'active')`,
		schedStudentID+sid, name, schedStudentID+sid+"@test.local", schedTestDept)
	if err != nil {
		t.Fatalf("创建测试用户 %s 失败: %v", name, err)
	}
	uid64, _ := res.LastInsertId()
	return int(uid64)
}

func addAvailability(t *testing.T, uid, weekday, period int) {
	t.Helper()
	if _, err := db.GetDB().Exec(
		`INSERT INTO availability (user_id, week, weekday, period) VALUES (?, ?, ?, ?)`,
		uid, schedTestWeek, weekday, period); err != nil {
		t.Fatalf("写入无课记录 (uid=%d, %d-%d) 失败: %v", uid, weekday, period, err)
	}
}

func addDutyRecord(t *testing.T, uid, week, weekday, period int, status string) {
	t.Helper()
	if _, err := db.GetDB().Exec(
		`INSERT INTO duty_records (week, weekday, period, user_id, department, assigned_by, status)
		 VALUES (?, ?, ?, ?, ?, NULL, ?)`,
		week, weekday, period, uid, schedTestDept, status); err != nil {
		t.Fatalf("写入值班记录 (uid=%d, %d周%d-%d, %s) 失败: %v", uid, week, weekday, period, status, err)
	}
}

// countUserInGrid 统计某用户在预览网格中出现的次数
func countUserInGrid(preview *model.SchedulePreview, uid int) int {
	n := 0
	for _, day := range preview.Grid {
		for _, cell := range day {
			for _, u := range cell.Users {
				if u.ID == uid {
					n++
				}
			}
		}
	}
	return n
}

// Test1 过滤能力：MaxPerDay（历史同天已有班）与 MaxPerWeek（历史本周班数已满）应把候选人过滤掉。
//
// 布局：周一 4 个节次。
// 过滤能力：MaxPerDay / MaxPerWeek 以「本次预览的本地计数」为准，且不回查当周库数据。
//
// 布局：周一 4 个节次，A、B 均无课覆盖全部节次，MaxPerDay=1、MaxPerWeek=1。
//   - 新语义：每人本周最多 1 班 → A、B 各排 1 次，其余 2 格 conflict；
//   - 同时给 A 预置一条「当周」库内记录（历史草稿），期望不影响本次结果
//     （若仍把当周库计数并入过滤，A 会被误滤，冲突数会变 4、A 出格 0 次）。
func TestSchedulePreview_FiltersByDayAndWeekLimits(t *testing.T) {
	svc := setupScheduleTestDB(t)

	uidA := createSchedTestUser(t, "A", "过滤A")
	uidB := createSchedTestUser(t, "B", "过滤B")

	for p := 1; p <= 4; p++ {
		addAvailability(t, uidA, 1, p)
		addAvailability(t, uidB, 1, p)
	}

	// A 的当周草稿记录：不应参与过滤判定
	addDutyRecord(t, uidA, schedTestWeek, 1, 2, "pending")

	preview, err := svc.PreviewSchedule(&model.ScheduleRequest{
		Week:        schedTestWeek,
		Days:        []int{1},
		NeedPerCell: 1,
		MinPerCell:  1,
		Periods:     4,
		MaxPerDay:   1,
		MaxPerWeek:  1,
		Department:  schedTestDept,
	})
	if err != nil {
		t.Fatalf("PreviewSchedule 失败: %v", err)
	}

	aCount := countUserInGrid(preview, uidA)
	bCount := countUserInGrid(preview, uidB)
	if aCount != 1 {
		t.Errorf("MaxPerWeek/MaxPerDay 本地计数失效：A 应恰好排 1 次，实际 %d 次（当周库记录不应参与过滤）", aCount)
	}
	if bCount != 1 {
		t.Errorf("MaxPerWeek/MaxPerDay 本地计数失效：B 应恰好排 1 次，实际 %d 次", bCount)
	}
	if len(preview.Conflicts) != 2 {
		t.Errorf("4 格中 2 格无人可排，应标记 2 个 conflict，实际 %d 个: %+v", len(preview.Conflicts), preview.Conflicts)
	} else {
		t.Logf("过滤 OK：A=1 次、B=1 次（本地计数生效），conflict=%d", len(preview.Conflicts))
	}
}

// Test2 按历史排班次数调整：全历史有效班次少者优先（cancelled 不计数）。
//
// 布局：周三第 1 节仅有 F、H 两人无课，只排 1 人。
//   - F：历史 3 班（有效，status=pending）
//   - H：历史 3 班但全部 status=cancelled → 有效班次 0
//
// 多启动 GRASP 取总分最小的最优解，最优解必然选 H（负担 0+1 < 3+1）。
func TestSchedulePreview_PrefersFewerHistoricalDuties(t *testing.T) {
	svc := setupScheduleTestDB(t)

	uidF := createSchedTestUser(t, "F", "高频F")
	uidH := createSchedTestUser(t, "H", "低频H")

	addAvailability(t, uidF, 3, 1)
	addAvailability(t, uidH, 3, 1)

	// F：3 条有效历史班次
	for p := 1; p <= 3; p++ {
		addDutyRecord(t, uidF, schedHistWeek, 1, p, "pending")
	}
	// H：3 条已取消班次 → 有效计数应为 0
	for p := 1; p <= 3; p++ {
		addDutyRecord(t, uidH, schedHistWeek, 2, p, "cancelled")
	}

	preview, err := svc.PreviewSchedule(&model.ScheduleRequest{
		Week:       schedTestWeek,
		Days:       []int{3},
		NeedPerCell: 1,
		MinPerCell: 1,
		Periods:    1,
		MaxPerDay:  1,
		MaxPerWeek: 5,
		Department: schedTestDept,
	})
	if err != nil {
		t.Fatalf("PreviewSchedule 失败: %v", err)
	}

	cell := preview.Grid[2][0] // weekday=3, period=1
	if len(cell.Users) != 1 {
		t.Fatalf("唯一时段应排 1 人，实际 %d 人: %+v", len(cell.Users), cell.Users)
	}
	if cell.Users[0].ID == uidH {
		t.Logf("次数调整 OK：有效班次少的 H（0 次，cancelled 不计）被优先排上，F（3 次）未排")
	} else if cell.Users[0].ID == uidF {
		t.Errorf("次数调整失败：排上了历史负担更重的 F，cancelled 班次可能被误计入全历史次数")
	} else {
		t.Errorf("排上了未知用户 id=%d", cell.Users[0].ID)
	}
}
