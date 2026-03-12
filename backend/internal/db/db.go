package db

import (
	"fmt"
	"schedule-system-v2/backend/internal/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
)

var DB *sqlx.DB

func InitDB(cfg *config.DatabaseConfig) error {
	// 确保 charset 有默认值
	charset := cfg.Charset
	if charset == "" {
		charset = "utf8mb4"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		charset,
	)

	var err error
	DB, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}

	DB.SetMaxOpenConns(100)
	DB.SetMaxIdleConns(10)

	return nil
}

func GetDB() *sqlx.DB {
	return DB
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
