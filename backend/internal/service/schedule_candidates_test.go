package service

import (
	"schedule-system-v2/backend/internal/model"
	"strings"
	"testing"
)

func TestCandidateRules(t *testing.T) {
	user := model.User{ID: 7, Name: "测试成员"}
	cells := []model.ConfirmCell{{Weekday: 1, Period: 1, UserIDs: []int{7}}, {Weekday: 2, Period: 1, UserIDs: []int{7}}}
	tests := []struct {
		name                         string
		day, period, maxDay, maxWeek int
		available                    bool
		cells                        []model.ConfirmCell
		reason                       string
	}{
		{"available", 3, 1, 1, 3, true, cells, ""},
		{"daily limit", 1, 2, 1, 3, true, cells, "日最大"},
		{"weekly limit", 3, 1, 2, 2, true, cells, "周最大"},
		{"class or missing availability", 3, 1, 2, 3, false, cells, "有课"},
		{"duplicate slot", 1, 1, 2, 3, true, cells, "已添加"},
		{"removal frees capacity", 1, 1, 1, 2, true, cells[1:], ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := candidateFor(user, tc.day, tc.period, tc.maxDay, tc.maxWeek, tc.cells, tc.available)
			if tc.reason == "" && !c.Eligible {
				t.Fatal(c.Reasons)
			}
			if tc.reason != "" && (c.Eligible || !strings.Contains(strings.Join(c.Reasons, " "), tc.reason)) {
				t.Fatal(c)
			}
		})
	}
}
