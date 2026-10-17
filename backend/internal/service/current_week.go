package service

import (
	"schedule-system-v2/backend/internal/model"
	"time"
)

// Use civil dates in Beijing, independent of the host timezone and time of day.
var schoolLocation = time.FixedZone("Asia/Shanghai", 8*60*60)

func civilDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func weekAnchor(settings *model.ScheduleSettings, now time.Time) time.Time {
	today := civilDate(now.In(schoolLocation))
	if settings != nil && settings.SemesterStartDate != nil {
		start := civilDate(*settings.SemesterStartDate)
		if !today.Before(start) {
			days := int(today.Sub(start).Hours() / 24)
			return today.AddDate(0, 0, -(days % 7))
		}
	}
	return today.AddDate(0, 0, -((int(today.Weekday()) + 6) % 7))
}

func currentWeekAt(settings *model.ScheduleSettings, now time.Time) int {
	base := settings.CurrentWeek
	anchor := settings.CurrentWeekSetAt
	if anchor == nil && settings.SemesterStartDate != nil {
		base, anchor = 1, settings.SemesterStartDate
	}
	// Legacy installations without either date start advancing from their saved week.
	if anchor == nil && !settings.UpdatedAt.IsZero() {
		date := weekAnchor(nil, settings.UpdatedAt)
		anchor = &date
	}
	if anchor != nil {
		days := int(civilDate(now.In(schoolLocation)).Sub(civilDate(*anchor)).Hours() / 24)
		if days >= 0 {
			base += days / 7
		} else {
			base = 1
		}
	}
	if base < 1 {
		return 1
	}
	if base > 30 {
		return 30
	}
	return base
}
