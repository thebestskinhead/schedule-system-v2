//go:build integration

package service

import (
	"database/sql"
	"fmt"
	"schedule-system-v2/backend/internal/config"
	"schedule-system-v2/backend/internal/db"
	"schedule-system-v2/backend/internal/model"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// Only creates tables in a uniquely named, disposable local database.
func TestWeekAndCandidatesDatabase(t *testing.T) {
	cfg, err := config.LoadConfig("../../configs/config.yaml")
	if err != nil {
		t.Skip("没有本地数据库配置")
	}
	if cfg.Database.Host != "localhost" && cfg.Database.Host != "127.0.0.1" {
		t.Skip("仅在本地数据库运行")
	}
	conn := mysql.NewConfig()
	conn.User, conn.Passwd = cfg.Database.User, cfg.Database.Password
	conn.Net, conn.Addr = "tcp", cfg.Database.Host+":"+cfg.Database.Port
	conn.ParseTime = true
	conn.Timeout = 3 * time.Second
	admin, err := sql.Open("mysql", conn.FormatDSN())
	if err != nil {
		t.Fatal("无法打开本地数据库连接")
	}
	defer admin.Close()
	if err := admin.Ping(); err != nil {
		t.Skip("本地MySQL未运行")
	}
	name := fmt.Sprintf("schedule_regression_%d", time.Now().UnixNano())
	if _, err := admin.Exec("CREATE DATABASE `" + name + "` CHARACTER SET utf8mb4"); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && (mysqlErr.Number == 1044 || mysqlErr.Number == 1045) {
			t.Skip("本地测试账号无权创建隔离数据库，未修改现有数据库")
		}
		t.Fatal(err)
	}
	defer admin.Exec("DROP DATABASE `" + name + "`")
	conn.DBName = name
	database, err := sqlx.Connect("mysql", conn.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	original := db.DB
	db.DB = database
	defer func() { db.DB = original }()
	exec := func(query string, args ...interface{}) {
		t.Helper()
		if _, err := database.Exec(query, args...); err != nil {
			t.Fatal(err)
		}
	}
	// Deliberately start with the old schema and disabled flag to exercise the upgrade.
	exec(`CREATE TABLE schedule_settings (
		id INT PRIMARY KEY AUTO_INCREMENT, config_key VARCHAR(20) UNIQUE,
		admin_id INT, current_week INT DEFAULT 1, auto_increment TINYINT DEFAULT 0,
		need_per_cell INT, min_per_cell INT, max_per_day INT, max_per_week INT,
		export_title VARCHAR(255), semester_start_date DATE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP)`)
	exec(`INSERT INTO schedule_settings (config_key, admin_id, current_week, auto_increment, need_per_cell, min_per_cell, max_per_day, max_per_week, export_title, semester_start_date)
		VALUES ('global', 1, 1, 0, 2, 0, 1, 2, '排班', '2026-09-07')`)
	if err := db.EnsureScheduleWeekSchema(database.DB); err != nil {
		t.Fatal(err)
	}
	if err := db.EnsureScheduleWeekSchema(database.DB); err != nil {
		t.Fatal("重复升级失败:", err)
	}
	s := NewScheduleService()
	if err := s.UpdateCurrentWeek(1, 9); err != nil {
		t.Fatal(err)
	}
	settings, err := s.GetScheduleSettings()
	if err != nil {
		t.Fatal(err)
	}
	if settings.CurrentWeek != 9 || !settings.AutoIncrement {
		t.Fatalf("校准未生效: %+v", settings)
	}
	settings.AutoIncrement = false // Older clients cannot turn progression off.
	settings.MaxPerWeek = 3
	if err := s.SaveScheduleSettings(1, settings); err != nil {
		t.Fatal(err)
	}
	settings, err = s.GetScheduleSettings()
	if err != nil || settings.CurrentWeek != 9 || !settings.AutoIncrement || settings.SemesterStartDate == nil {
		t.Fatalf("参数保存丢失周次或学期: %v %+v", err, settings)
	}
	start, err := s.GetSemesterStartDate()
	if err != nil || start == nil || *start != "2026-09-07" {
		t.Fatalf("日期格式错误: %v %v", start, err)
	}

	exec(`CREATE TABLE users (id INT PRIMARY KEY, student_id VARCHAR(50), name VARCHAR(50), email VARCHAR(100), role VARCHAR(20), department VARCHAR(50), dept_role VARCHAR(20), employment_status VARCHAR(20), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)`)
	exec(`CREATE TABLE availability (user_id INT, week INT, weekday INT, period INT)`)
	exec(`CREATE TABLE leave_requests (user_id INT, status VARCHAR(20), start_date DATE, end_date DATE)`)
	exec(`INSERT INTO users (id, student_id, name, role, department, dept_role, employment_status) VALUES
		(7, 'S7', '无课成员', 'user', '测试部', 'dept_member', 'active'),
		(8, 'S8', '有课成员', 'user', '测试部', 'dept_member', 'active')`)
	exec(`INSERT INTO availability VALUES (7, 9, 1, 1), (7, 9, 1, 2)`)
	req := &model.CandidateRequest{Week: 9, Weekday: 1, Period: 1, Department: "测试部", MaxPerDay: 1, MaxPerWeek: 2}
	candidates, err := s.GetScheduleCandidates(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 2 {
		t.Fatal(candidates)
	}
	for _, candidate := range candidates {
		if candidate.Eligible != (candidate.ID == 7) {
			t.Fatal(candidate)
		}
	}
	draft := &model.ConfirmScheduleRequest{Week: 9, MaxPerDay: 1, MaxPerWeek: 2, Cells: []model.ConfirmCell{{Weekday: 1, Period: 1, UserIDs: []int{7}}}}
	if err := s.validateSchedule(draft); err != nil {
		t.Fatal(err)
	}
	draft.Cells = append(draft.Cells, model.ConfirmCell{Weekday: 1, Period: 2, UserIDs: []int{7}})
	if err := s.validateSchedule(draft); err == nil {
		t.Fatal("未拦截超过日上限的草稿")
	}
	draft.Cells = []model.ConfirmCell{{Weekday: 1, Period: 1, UserIDs: []int{8}}}
	if err := s.validateSchedule(draft); err == nil {
		t.Fatal("未拦截无无课记录的成员")
	}
}
