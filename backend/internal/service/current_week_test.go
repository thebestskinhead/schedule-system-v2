package service

import (
	"schedule-system-v2/backend/internal/model"
	"testing"
	"time"
)

func date(s string) *time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return &t
}

func TestCurrentWeekAt(t *testing.T) {
	tests := []struct {
		name     string
		settings model.ScheduleSettings
		now      string
		want     int
	}{
		{"semester start", model.ScheduleSettings{SemesterStartDate: date("2026-09-07")}, "2026-09-07T00:00:00+08:00", 1},
		{"before new week", model.ScheduleSettings{SemesterStartDate: date("2026-09-07")}, "2026-09-13T23:59:59+08:00", 1},
		{"Beijing midnight on UTC host", model.ScheduleSettings{SemesterStartDate: date("2026-09-07")}, "2026-09-13T16:00:00Z", 2},
		{"manual overrides semester and disabled legacy flag", model.ScheduleSettings{CurrentWeek: 9, CurrentWeekSetAt: date("2026-10-12"), SemesterStartDate: date("2026-09-07")}, "2026-10-17T12:00:00+08:00", 9},
		{"manual continues next week", model.ScheduleSettings{CurrentWeek: 9, CurrentWeekSetAt: date("2026-10-12")}, "2026-10-19T00:00:00+08:00", 10},
		{"cap at thirty", model.ScheduleSettings{CurrentWeek: 30, CurrentWeekSetAt: date("2026-10-12")}, "2026-10-26T00:00:00+08:00", 30},
		{"future semester", model.ScheduleSettings{SemesterStartDate: date("2026-11-02")}, "2026-10-17T12:00:00+08:00", 1},
		{"empty defaults", model.ScheduleSettings{}, "2026-10-17T12:00:00+08:00", 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			now, _ := time.Parse(time.RFC3339, tc.now)
			if got := currentWeekAt(&tc.settings, now); got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestWeekAnchor(t *testing.T) {
	now, _ := time.Parse(time.RFC3339, "2026-10-17T12:00:00+08:00")
	if got := weekAnchor(nil, now).Format("2006-01-02"); got != "2026-10-12" {
		t.Fatal(got)
	}
	// An academic week follows the semester start even when it is not a Monday.
	settings := &model.ScheduleSettings{SemesterStartDate: date("2026-09-09")}
	if got := weekAnchor(settings, now).Format("2006-01-02"); got != "2026-10-14" {
		t.Fatal(got)
	}
}
