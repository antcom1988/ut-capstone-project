package handlers

import (
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"net/http/httptest"
	"strings"
	"testing"
)

func validEmployeePayload() map[string]any {
	return map[string]any{"nip": "199003202014022005", "nama_lengkap": "Employee", "status_pegawai": "PNS", "bidang": "Sekretariat", "jenis_jabatan": "Pelaksana", "nama_jabatan": "Analis", "pangkat": "Penata", "golongan": "III/c", "tmt_pengangkatan": "2020-01-01", "tmt_pangkat_terakhir": "2020-01-01", "tmt_kgb_terakhir": "2020-01-01", "initial_password": "password-123"}
}
func TestCreateRejectsInvalidMasterBeforeDatabase(t *testing.T) {
	for _, tt := range []struct {
		key   string
		value any
	}{{"initial_password", ""}, {"initial_password", "short"}, {"status_pegawai", "INVALID"}, {"jenis_jabatan", "INVALID"}, {"jenis_kelamin", "X"}, {"nama_lengkap", strings.Repeat("a", 151)}, {"email", "not-email"}, {"no_hp", strings.Repeat("1", 21)}} {
		t.Run(tt.key+"/"+tt.value.(string), func(t *testing.T) {
			db, m := testDB(t)
			m.ExpectBegin().WillReturnError(nil)
			p := validEmployeePayload()
			p[tt.key] = tt.value
			b, _ := json.Marshal(p)
			r := httptest.NewRequest("POST", "/", strings.NewReader(string(b)))
			r.Header.Set("Content-Type", "application/json")
			res, e := testApp((&PegawaiHandler{DB: db}).CreatePegawai).Test(r)
			if e != nil {
				t.Fatal(e)
			}
			defer res.Body.Close()
			if res.StatusCode != 400 {
				t.Errorf("got %d want 400", res.StatusCode)
			}
		})
	}
}
func TestContactRejectsInvalidEmail(t *testing.T) {
	db, m := testDB(t)
	m.ExpectExec("UPDATE sikapj.pegawai").WillReturnResult(sqlmock.NewResult(0, 1))
	r := httptest.NewRequest("PUT", "/", strings.NewReader(`{"email":"bad"}`))
	r.Header.Set("Content-Type", "application/json")
	res, e := testApp((&PegawaiHandler{DB: db}).UpdateMyContact).Test(r)
	if e != nil {
		t.Fatal(e)
	}
	defer res.Body.Close()
	if res.StatusCode != 400 {
		t.Errorf("got %d want 400", res.StatusCode)
	}
}
