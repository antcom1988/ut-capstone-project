package services

import (
	"testing"
	"time"
)

func TestEvaluateAlertStatusAtUsesCalendarMonths(t *testing.T) {
	due := time.Date(2027, 4, 30, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		now  time.Time
		want string
	}{
		{"before threshold", time.Date(2027, 1, 29, 0, 0, 0, 0, time.UTC), "AMAN"},
		{"at threshold", time.Date(2027, 1, 30, 0, 0, 0, 0, time.UTC), "DIINGATKAN"},
		{"at urgent month", time.Date(2027, 3, 30, 0, 0, 0, 0, time.UTC), "SEGERA"},
		{"overdue", time.Date(2027, 5, 1, 0, 0, 0, 0, time.UTC), "SEGERA"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, _ := EvaluateAlertStatusAt(tt.now, due, 3)
			if got != tt.want {
				t.Fatalf("status=%s want=%s", got, tt.want)
			}
		})
	}
}

func TestCalculateDueDateClampsLeapDay(t *testing.T) {
	leap := time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)
	got := CalculateKGBDueDate(leap, leap)
	if got.Format(time.DateOnly) != "2026-02-28" {
		t.Fatalf("got %s", got.Format(time.DateOnly))
	}
}
