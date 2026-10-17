package services

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"sikapj-backend/internal/models"
)

func TestSyncAlertLogsRejectsSettingsQueryFailure(t *testing.T) {
	raw, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	failure := errors.New("database unavailable")
	mock.ExpectQuery("SELECT value").WillReturnError(failure)
	if _, err := SyncAlertLogs(sqlx.NewDb(raw, "sqlmock"), 3); !errors.Is(err, failure) {
		t.Fatalf("expected settings failure, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSyncAlertLogsClosesKPForFormerPNS(t *testing.T) {
	raw, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	mock.ExpectQuery("SELECT value").WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("3"))
	mock.ExpectQuery("SELECT \\*").WillReturnRows(sqlmock.NewRows([]string{"nip", "status_pegawai", "tmt_kgb_terakhir", "tmt_pangkat_terakhir"}).AddRow("123456789012345678", "PPPK", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)))
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO sikapj.alert_log").WithArgs("123456789012345678", "KGB", sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE sikapj.alert_log SET").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE sikapj.alert_log l.*l.tipe_alert.*KP.*p.status_pegawai.*PNS").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	count, err := SyncAlertLogs(sqlx.NewDb(raw, "sqlmock"), 3)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("count=%d want 1", count)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPPPKHasKGBButNoKP(t *testing.T) {
	p := models.Pegawai{StatusPegawai: "PPPK", TMTKGBTerakhir: time.Now(), TMTPangkatTerakhir: time.Now()}
	alerts := BuildPegawaiAlertItems(&p, 3)
	if len(alerts) != 1 || alerts[0].TipeAlert != "KGB" {
		t.Fatalf("unexpected alerts: %+v", alerts)
	}
}

func TestCalendarDaysIgnoreDSTAndTimeZones(t *testing.T) {
	// Consecutive civil dates are one day apart despite different time zones.
	now := time.Date(2027, 3, 28, 12, 0, 0, 0, time.FixedZone("local", 7*3600))
	due := time.Date(2027, 3, 29, 0, 0, 0, 0, time.UTC)
	_, _, days := EvaluateAlertStatusAt(now, due, 3)
	if days != 1 {
		t.Fatalf("days=%d want 1", days)
	}
	// Spring-forward midnight to midnight is only 23 elapsed hours, but one civil day.
	now = time.Date(2027, 3, 28, 0, 0, 0, 0, time.FixedZone("winter", 3600))
	due = time.Date(2027, 3, 29, 0, 0, 0, 0, time.FixedZone("summer", 7200))
	_, _, days = EvaluateAlertStatusAt(now, due, 3)
	if days != 1 {
		t.Fatalf("DST days=%d want 1", days)
	}
}
