package handlers

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUploadRejectsUnsafeNIP(t *testing.T) {
	for name, h := range map[string]fiber.Handler{"berkas": (&BerkasHandler{}).UploadBerkas, "kinerja": (&KinerjaDiklatHandler{}).UploadArsipKinerja, "diklat": (&KinerjaDiklatHandler{}).CreateRiwayatDiklat} {
		t.Run(name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/", strings.NewReader("nip=..%2Fescape"))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			res, e := testApp(h).Test(r)
			if e != nil {
				t.Fatal(e)
			}
			defer res.Body.Close()
			b := make([]byte, 4096)
			n, _ := res.Body.Read(b)
			if res.StatusCode != 400 || !strings.Contains(string(b[:n]), "NIP") {
				t.Fatalf("%d %s", res.StatusCode, b[:n])
			}
		})
	}
}
func TestVerificationRequiresUploadVersion(t *testing.T) {
	db, m := testDB(t)
	m.ExpectExec("UPDATE sikapj.e_berkas").WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectExec("INSERT INTO sikapj.audit_log").WillReturnResult(sqlmock.NewResult(1, 1))
	r := httptest.NewRequest("PUT", "/1", strings.NewReader(`{"status_verifikasi":"Disetujui"}`))
	r.Header.Set("Content-Type", "application/json")
	res, e := testApp((&BerkasHandler{DB: db}).VerifikasiBerkas).Test(r)
	if e != nil {
		t.Fatal(e)
	}
	defer res.Body.Close()
	if res.StatusCode != 400 {
		t.Fatalf("got %d want 400", res.StatusCode)
	}
}
func TestDiklatRequiresCertificate(t *testing.T) {
	db, m := testDB(t)
	m.ExpectExec("INSERT INTO sikapj.riwayat_diklat").WillReturnResult(sqlmock.NewResult(1, 1))
	r := httptest.NewRequest("POST", "/", strings.NewReader("nip=199003202014022005&nama_diklat=Test&penyelenggara=Test&jumlah_jp=8&tanggal_mulai=2026-01-01&tanggal_selesai=2026-01-02"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, e := testApp((&KinerjaDiklatHandler{DB: db}).CreateRiwayatDiklat).Test(r)
	if e != nil {
		t.Fatal(e)
	}
	defer res.Body.Close()
	if res.StatusCode != 400 {
		t.Fatalf("got %d want 400", res.StatusCode)
	}
}
