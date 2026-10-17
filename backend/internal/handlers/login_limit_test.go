package handlers

import (
	"database/sql"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoginLimitsRepeatedFailures(t *testing.T) {
	db, m := testDB(t)
	h := &AuthHandler{DB: db}
	a := testApp(h.Login)
	for i := 0; i < 11; i++ {
		if i < 10 {
			m.ExpectQuery("SELECT .*FROM sikapj.users").WillReturnError(sql.ErrNoRows)
		}
		r := httptest.NewRequest("POST", "/login", strings.NewReader(`{"username":"missing","password":"wrong"}`))
		r.Header.Set("Content-Type", "application/json")
		res, e := a.Test(r)
		if e != nil {
			t.Fatal(e)
		}
		res.Body.Close()
		want := 401
		if i == 10 {
			want = 429
		}
		if res.StatusCode != want {
			t.Fatalf("attempt %d: got %d want %d", i, res.StatusCode, want)
		}
	}
	if e := m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
