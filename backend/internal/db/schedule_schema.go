package db

import "database/sql"

// EnsureScheduleWeekSchema upgrades existing installations without changing semester dates.
// It is also called by the installer after creating schedule_settings.
func EnsureScheduleWeekSchema(database *sql.DB) error {
	var exists int
	if err := database.QueryRow(`SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'schedule_settings'`).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return nil
	}
	if err := database.QueryRow(`SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'schedule_settings' AND column_name = 'current_week_set_at'`).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		if _, err := database.Exec(`ALTER TABLE schedule_settings ADD COLUMN current_week_set_at DATE DEFAULT NULL COMMENT '手动周次的起算日期', ALTER COLUMN auto_increment SET DEFAULT 1`); err != nil {
			return err
		}
	}
	// Freeze the legacy base before later parameter writes can change updated_at.
	if _, err := database.Exec(`UPDATE schedule_settings SET current_week_set_at = DATE_SUB(DATE(updated_at), INTERVAL WEEKDAY(updated_at) DAY) WHERE current_week_set_at IS NULL AND semester_start_date IS NULL`); err != nil {
		return err
	}
	_, err := database.Exec(`UPDATE schedule_settings SET auto_increment = 1 WHERE auto_increment = 0`)
	return err
}
