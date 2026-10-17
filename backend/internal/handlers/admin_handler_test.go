package handlers

import (
	"testing"
	"time"
)

func TestParseAuditBoundaryAcceptsDateAndDateTime(t *testing.T) {
	date, hasTime, err := parseAuditBoundary("2026-10-13", false)
	if err != nil || hasTime || date.Format(time.DateOnly) != "2026-10-13" {
		t.Fatalf("date boundary invalid: %v %v %v", date, hasTime, err)
	}
	dateTime, hasTime, err := parseAuditBoundary("2026-10-13T17:30", false)
	if err != nil || !hasTime || dateTime.Format("2006-01-02T15:04") != "2026-10-13T17:30" {
		t.Fatalf("datetime boundary invalid: %v %v %v", dateTime, hasTime, err)
	}
}

func TestParseAuditBoundaryMakesDateOnlyEndExclusive(t *testing.T) {
	end, hasTime, err := parseAuditBoundary("2026-10-13", true)
	if err != nil || hasTime || end.Format(time.DateOnly) != "2026-10-14" {
		t.Fatalf("end boundary invalid: %v %v %v", end, hasTime, err)
	}
}
