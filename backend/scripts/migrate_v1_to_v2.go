package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// V1User v1版本用户结构
type V1User struct {
	ID        int
	StudentID string
	Name      string
	Email     string
	Password  string
	Role      string
	IsActive  int
	CreatedAt string
	UpdatedAt string
}

// V1Availability v1版本无课表结构
type V1Availability struct {
	ID        int
	UserID    int
	Week      int
	Weekday   int
	Period    int
	CreatedAt string
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	Charset  string
}

func main() {
	// 从配置文件加载数据库配置
	cfg := loadConfig()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.Charset)
	if len(os.Args) > 1 {
		dsn = os.Args[1]
	}

	fmt.Printf("数据库配置: %s@%s:%s/%s\n", cfg.User, cfg.Host, cfg.Port, cfg.DBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("连接数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		fmt.Printf("Ping数据库失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("数据库连接成功")

	// 解析v1数据
	users, err := parseV1Users("/workspace/schedule-system-v2/v1-data/users_export/db_dump.sql")
	if err != nil {
		fmt.Printf("解析v1用户数据失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("解析到 %d 个v1用户\n", len(users))

	availabilities, err := parseV1Availability("/workspace/schedule-system-v2/v1-data/availability_export/db_dump.sql")
	if err != nil {
		fmt.Printf("解析v1无课表数据失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("解析到 %d 条v1无课表记录\n", len(availabilities))

	// 开始迁移
	if err := migrateUsers(db, users); err != nil {
		fmt.Printf("迁移用户失败: %v\n", err)
		os.Exit(1)
	}

	if err := migrateAvailability(db, availabilities); err != nil {
		fmt.Printf("迁移无课表失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n数据迁移完成!")
}

// loadConfig 加载配置文件
func loadConfig() DatabaseConfig {
	// 尝试读取配置文件
	data, err := os.ReadFile("configs/config.yaml")
	if err == nil {
		cfg := parseYAML(string(data))
		if cfg != nil {
			return *cfg
		}
	}

	// 使用默认配置
	return DatabaseConfig{
		Host:     "localhost",
		Port:     "3306",
		User:     "root",
		Password: "Schedule@2024",
		DBName:   "schedule_system_v2",
		Charset:  "utf8mb4",
	}
}

// parseYAML 简单解析YAML配置
func parseYAML(content string) *DatabaseConfig {
	cfg := &DatabaseConfig{
		Charset: "utf8mb4",
	}

	lines := strings.Split(content, "\n")
	inDatabase := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "database:") {
			inDatabase = true
			continue
		}

		if inDatabase {
			if strings.HasPrefix(line, "host:") {
				cfg.Host = strings.Trim(strings.SplitN(line, ":", 2)[1], " \"")
			}
			if strings.HasPrefix(line, "port:") {
				cfg.Port = strings.Trim(strings.SplitN(line, ":", 2)[1], " \"")
			}
			if strings.HasPrefix(line, "user:") {
				cfg.User = strings.Trim(strings.SplitN(line, ":", 2)[1], " \"")
			}
			if strings.HasPrefix(line, "password:") {
				cfg.Password = strings.Trim(strings.SplitN(line, ":", 2)[1], " \"")
			}
			if strings.HasPrefix(line, "dbname:") {
				cfg.DBName = strings.Trim(strings.SplitN(line, ":", 2)[1], " \"")
			}
			// 检查是否进入其他section
			if !strings.HasPrefix(line, "") && strings.Contains(line, ":") && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "host") && !strings.HasPrefix(line, "port") && !strings.HasPrefix(line, "user") && !strings.HasPrefix(line, "pass") && !strings.HasPrefix(line, "db") && !strings.HasPrefix(line, "char") {
				break
			}
		}
	}

	return cfg
}

// parseV1Users 解析v1用户SQL文件
func parseV1Users(filename string) ([]V1User, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var users []V1User
	// 匹配 VALUES ('1','2505050318','name','email','pass','role','1','created','updated')
	re := regexp.MustCompile(`VALUES \('([0-9]+)','([^']*)','([^']*)','([^']*)','([^']*)','([^']*)','([0-9]+)','([^']*)','([^']*)'\)`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "INSERT INTO `users`") {
			continue
		}

		matches := re.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) >= 10 {
				user := V1User{
					ID:        parseInt(match[1]),
					StudentID: match[2],
					Name:      match[3],
					Email:     match[4],
					Password:  match[5],
					Role:      match[6],
					IsActive:  parseInt(match[7]),
					CreatedAt: match[8],
					UpdatedAt: match[9],
				}
				users = append(users, user)
			}
		}
	}

	return users, scanner.Err()
}

// parseV1Availability 解析v1无课表SQL文件
func parseV1Availability(filename string) ([]V1Availability, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var availabilities []V1Availability
	re := regexp.MustCompile(`VALUES \('([0-9]+)','([0-9]+)','([0-9]+)','([0-9]+)','([0-9]+)','([^']*)'\)`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "INSERT INTO `availability`") {
			continue
		}

		matches := re.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) >= 7 {
				avail := V1Availability{
					ID:        parseInt(match[1]),
					UserID:    parseInt(match[2]),
					Week:      parseInt(match[3]),
					Weekday:   parseInt(match[4]),
					Period:    parseInt(match[5]),
					CreatedAt: match[6],
				}
				availabilities = append(availabilities, avail)
			}
		}
	}

	return availabilities, scanner.Err()
}

func parseInt(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

// migrateUsers 迁移用户数据
func migrateUsers(db *sql.DB, users []V1User) error {
	fmt.Println("\n开始迁移用户数据...")

	// 清空现有用户数据（包括id=1，因为v1数据中也有id=1的用户）
	_, err := db.Exec("DELETE FROM users")
	if err != nil {
		return fmt.Errorf("清空用户表失败: %v", err)
	}
	fmt.Println("  已清空现有用户数据")

	// 准备插入语句
	stmt, err := db.Prepare(`
		INSERT INTO users (id, student_id, name, email, password, role, department, dept_role, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("准备插入语句失败: %v", err)
	}
	defer stmt.Close()

	for _, u := range users {
		// 映射角色
		role := u.Role
		deptRole := "dept_member"

		// 特殊处理：2505050318 为系统管理员
		if u.StudentID == "2505050318" {
			role = "admin"
			deptRole = "dept_admin"
			fmt.Printf("  设置 %s(%s) 为系统管理员+部门管理员\n", u.Name, u.StudentID)
		} else if u.Role == "admin" {
			// 其他v1的admin设为系统管理员和部门管理员
			role = "admin"
			deptRole = "dept_admin"
			fmt.Printf("  设置 %s(%s) 为系统管理员+部门管理员\n", u.Name, u.StudentID)
		} else {
			fmt.Printf("  添加用户 %s(%s) 为科普部成员\n", u.Name, u.StudentID)
		}

		_, err := stmt.Exec(
			u.ID,
			u.StudentID,
			u.Name,
			u.Email,
			u.Password,
			role,
			"科普部", // 所有人均为科普部
			deptRole,
			u.IsActive,
			u.CreatedAt,
			u.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("插入用户 %s 失败: %v", u.StudentID, err)
		}
	}

	fmt.Printf("成功迁移 %d 个用户\n", len(users))
	return nil
}

// migrateAvailability 迁移无课表数据
func migrateAvailability(db *sql.DB, availabilities []V1Availability) error {
	fmt.Println("\n开始迁移无课表数据...")

	// 清空现有无课表数据
	_, err := db.Exec("DELETE FROM availability")
	if err != nil {
		return fmt.Errorf("清空无课表失败: %v", err)
	}

	// 准备插入语句
	stmt, err := db.Prepare(`
		INSERT INTO availability (id, user_id, week, weekday, period, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("准备插入语句失败: %v", err)
	}
	defer stmt.Close()

	batchSize := 100
	count := 0

	for _, a := range availabilities {
		_, err := stmt.Exec(a.ID, a.UserID, a.Week, a.Weekday, a.Period, a.CreatedAt)
		if err != nil {
			return fmt.Errorf("插入无课表记录 %d 失败: %v", a.ID, err)
		}
		count++
		if count%batchSize == 0 {
			fmt.Printf("  已迁移 %d 条记录...\n", count)
		}
	}

	fmt.Printf("成功迁移 %d 条无课表记录\n", len(availabilities))
	return nil
}
