package config

import (
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// 配置文件写入锁，防止并发写入导致文件损坏
var configMutex sync.Mutex

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	JWT       JWTConfig       `yaml:"jwt"`
	Site      SiteConfig      `yaml:"site"`      // 网站配置
	Installed bool            `yaml:"installed"` // 是否已完成安装
}

type SiteConfig struct {
	Domain string `yaml:"domain" json:"domain"` // 网站域名，用于生成重置密码链接
}

type ServerConfig struct {
	Port string `yaml:"port"`
	Mode string `yaml:"mode"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     string `yaml:"port" json:"port"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
	DBName   string `yaml:"dbname" json:"dbname"`
	Charset  string `yaml:"charset" json:"charset"`
}

type JWTConfig struct {
	Secret string `yaml:"secret"`
	Expire int    `yaml:"expire"`
}

var GlobalConfig *Config

// 配置文件路径（固定）
const ConfigFilePath = "configs/config.yaml"

// IsInstalled 检查系统是否已安装（通过配置文件的 installed 字段）
func IsInstalled() bool {
	cfg, err := LoadConfig(ConfigFilePath)
	if err != nil {
		return false
	}
	return cfg.Installed
}

// LoadConfig 加载配置文件
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	GlobalConfig = &config
	return &config, nil
}

// SaveConfig 保存配置到文件（线程安全）
func SaveConfig(cfg *Config) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	// 确保目录存在
	dir := filepath.Dir(ConfigFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	// 使用临时文件+重命名的方式保证原子性写入
	tempFile := ConfigFilePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return err
	}

	// 原子重命名，防止写入过程中文件损坏
	if err := os.Rename(tempFile, ConfigFilePath); err != nil {
		os.Remove(tempFile)
		return err
	}

	GlobalConfig = cfg
	return nil
}

// GetConfig 获取配置（如果没有加载则返回默认配置）
func GetConfig() *Config {
	if GlobalConfig != nil {
		return GlobalConfig
	}
	return &Config{
		Server: ServerConfig{
			Port: "8080",
			Mode: "debug",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     "3306",
			User:     "root",
			Password: "Schedule@2024",
			DBName:   "schedule_system_v2",
			Charset:  "utf8mb4",
		},
		JWT: JWTConfig{
			Secret: "schedule-system-secret-key",
			Expire: 168,
		},
	}
}
