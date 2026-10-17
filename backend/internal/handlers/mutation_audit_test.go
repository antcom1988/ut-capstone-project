package handlers

import (
	"encoding/json"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestContactAuditFailureRollsBack(t *testing.T) {
	db, m := testDB(t)
	m.ExpectBegin()
	m.ExpectExec("UPDATE sikapj.pegawai").WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectExec("INSERT INTO sikapj.audit_log").WillReturnError(errors.New("audit unavailable"))
	m.ExpectRollback()
	r := httptest.NewRequest("PUT", "/", strings.NewReader(`{"email":"valid@example.org"}`))
	r.Header.Set("Content-Type", "application/json")
	res, e := testApp((&PegawaiHandler{DB: db}).UpdateMyContact).Test(r)
	if e != nil {
		t.Fatal(e)
	}
	res.Body.Close()
	if res.StatusCode != 500 {
		t.Fatalf("got %d", res.StatusCode)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
func TestCreateAuditFailureRollsBack(t *testing.T) {
	db, m := testDB(t)
	m.ExpectBegin()
	m.ExpectQuery("INSERT INTO sikapj.users .*must_change_password.*TRUE, TRUE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	m.ExpectExec("INSERT INTO sikapj.pegawai").WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectExec("INSERT INTO sikapj.audit_log").WithArgs(1, "CREATE_PEGAWAI", jsonArg{}, sqlmock.AnyArg()).WillReturnError(errors.New("audit unavailable"))
	m.ExpectRollback()
	p := validEmployeePayload()
	b, _ := json.Marshal(p)
	r := httptest.NewRequest("POST", "/", strings.NewReader(string(b)))
	r.Header.Set("Content-Type", "application/json")
	res, e := testApp((&PegawaiHandler{DB: db}).CreatePegawai).Test(r)
	if e != nil {
		t.Fatal(e)
	}
	res.Body.Close()
	if res.StatusCode != 500 {
		t.Fatalf("got %d", res.StatusCode)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
func TestVerificationStaleAndAuditRollback(t *testing.T) {
	for _, stale := range []bool{true, false} {
		db, m := testDB(t)
		m.ExpectBegin()
		n := int64(1)
		if stale {
			n = 0
		}
		m.ExpectExec("UPDATE sikapj.e_berkas .*WHERE id = \\$4 AND uploaded_at = \\$5").WillReturnResult(sqlmock.NewResult(0, n))
		if !stale {
			m.ExpectExec("INSERT INTO sikapj.audit_log").WillReturnError(errors.New("audit unavailable"))
		}
		m.ExpectRollback()
		r := httptest.NewRequest("PUT", "/1", strings.NewReader(`{"status_verifikasi":"Disetujui","uploaded_at":"2026-01-01T00:00:00Z"}`))
		r.Header.Set("Content-Type", "application/json")
		res, e := testApp((&BerkasHandler{DB: db}).VerifikasiBerkas).Test(r)
		if e != nil {
			t.Fatal(e)
		}
		res.Body.Close()
		want := 500
		if stale {
			want = 409
		}
		if res.StatusCode != want {
			t.Errorf("got %d want %d", res.StatusCode, want)
		}
		if e = m.ExpectationsWereMet(); e != nil {
			t.Error(e)
		}
	}
}
