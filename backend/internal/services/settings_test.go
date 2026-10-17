package services

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestAlertThreshold(t *testing.T) {
	for _, tc := range []struct {
		name, raw      string
		queryErr       error
		fallback, want int
		wantErr        bool
	}{
		{name: "configured", raw: "6", fallback: 3, want: 6},
		{name: "absent uses default", queryErr: sql.ErrNoRows, fallback: 3, want: 3},
		{name: "absent invalid default", queryErr: sql.ErrNoRows, fallback: 0, wantErr: true},
		{name: "database failure", queryErr: errors.New("offline"), fallback: 3, wantErr: true},
		{name: "invalid", raw: "abc", fallback: 3, wantErr: true},
		{name: "zero", raw: "0", fallback: 3, wantErr: true},
		{name: "negative", raw: "-2", fallback: 3, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer raw.Close()
			q := mock.ExpectQuery("SELECT value")
			if tc.queryErr != nil {
				q.WillReturnError(tc.queryErr)
			} else {
				q.WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow(tc.raw))
			}
			got, err := AlertThreshold(sqlx.NewDb(raw, "sqlmock"), tc.fallback)
			if (err != nil) != tc.wantErr || got != tc.want {
				t.Fatalf("got (%d,%v), want (%d,error=%t)", got, err, tc.want, tc.wantErr)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSyncAlertLogsRollsBackClosureFailure(t *testing.T) {
	raw, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	failure := errors.New("cannot close stale alerts")
	mock.ExpectQuery("SELECT value").WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("3"))
	mock.ExpectQuery("SELECT \\*").WillReturnRows(sqlmock.NewRows([]string{"nip"}))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE sikapj.alert_log").WillReturnError(failure)
	mock.ExpectRollback()
	count, err := SyncAlertLogs(sqlx.NewDb(raw, "sqlmock"), 3)
	if count != 0 || !errors.Is(err, failure) {
		t.Fatalf("count=%d error=%v", count, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
