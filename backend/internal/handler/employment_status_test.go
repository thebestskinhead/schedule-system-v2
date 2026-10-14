package handler

import (
	"testing"

	"schedule-system-v2/backend/internal/model"
)

func TestNormalizeEmploymentStatus(t *testing.T) {
	cases := []struct {
		raw      string
		want     string
		wantOK   bool
	}{
		{"", "", true},
		{"all", "", true},
		{model.EmploymentStatusPending, model.EmploymentStatusPending, true},
		{model.EmploymentStatusActive, model.EmploymentStatusActive, true},
		{model.EmploymentStatusResigned, model.EmploymentStatusResigned, true},
		{"bogus", "", false},
		{"ACTIVE", "", false},
	}
	for _, c := range cases {
		got, ok := normalizeEmploymentStatus(c.raw)
		if got != c.want || ok != c.wantOK {
			t.Fatalf("normalizeEmploymentStatus(%q) = (%q, %v), want (%q, %v)", c.raw, got, ok, c.want, c.wantOK)
		}
	}
}
