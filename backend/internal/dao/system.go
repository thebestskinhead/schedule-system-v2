package dao

import (
	"schedule-system-v2/backend/internal/db"
)

type SystemDAO struct{}

func NewSystemDAO() *SystemDAO {
	return &SystemDAO{}
}

func (d *SystemDAO) GetConfig(key string) (string, error) {
	var value string
	query := `SELECT config_value FROM system_config WHERE config_key = ?`
	err := db.GetDB().Get(&value, query, key)
	return value, err
}

func (d *SystemDAO) SetConfig(key, value string) error {
	query := `INSERT INTO system_config (config_key, config_value) VALUES (?, ?) 
		ON DUPLICATE KEY UPDATE config_value = VALUES(config_value)`
	_, err := db.GetDB().Exec(query, key, value)
	return err
}

func (d *SystemDAO) IsInitialized() (bool, error) {
	value, err := d.GetConfig("system_initialized")
	if err != nil {
		return false, nil
	}
	return value == "true", nil
}

func (d *SystemDAO) SetInitialized() error {
	return d.SetConfig("system_initialized", "true")
}

func (d *SystemDAO) GetSemester() (string, error) {
	return d.GetConfig("current_semester")
}

func (d *SystemDAO) SetSemester(semester string) error {
	return d.SetConfig("current_semester", semester)
}
