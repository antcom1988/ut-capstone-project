package handlers

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"net/http/httptest"
	"sikapj-backend/internal/config"
	"testing"
)

func TestKinerjaAcceptsImageAndOneMBLimit(t *testing.T) {
	for _, large := range []bool{false, true} {
		db, m := testDB(t)
		if !large {
			m.ExpectQuery("SELECT file_path").WillReturnRows(sqlmock.NewRows([]string{"file_path"}))
			m.ExpectBegin()
			m.ExpectExec("INSERT INTO sikapj.arsip_kinerja").WillReturnResult(sqlmock.NewResult(1, 1))
			m.ExpectExec("INSERT INTO sikapj.audit_log").WillReturnResult(sqlmock.NewResult(1, 1))
			m.ExpectCommit()
		}
		data := []byte{0x89, 'P', 'N', 'G', 13, 10, 26, 10}
		if large {
			data = append(data, make([]byte, 1024*1024)...)
		}
		status := multipartCall(t, (&KinerjaDiklatHandler{DB: db, Cfg: &config.Config{UploadDir: t.TempDir()}}).UploadArsipKinerja, map[string]string{"nip": "199003202014022005", "tahun_periode": "2025", "jenis_dokumen": "PK"}, "file", "image.png", data)
		want := 200
		if large {
			want = 400
		}
		if status != want {
			t.Errorf("got %d want %d", status, want)
		}
		if e := m.ExpectationsWereMet(); e != nil {
			t.Error(e)
		}
	}
}
func TestDiklatEditRetainsCertificateWithoutNIP(t *testing.T) {
	db, m := testDB(t)
	m.ExpectQuery("SELECT .*FROM sikapj.riwayat_diklat").WillReturnRows(sqlmock.NewRows([]string{"id", "nip", "file_sertifikat_path", "file_name"}).AddRow(1, "199003202014022005", "existing.pdf", "existing.pdf"))
	m.ExpectBegin()
	m.ExpectExec("UPDATE sikapj.riwayat_diklat").WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectExec("INSERT INTO sikapj.audit_log").WillReturnResult(sqlmock.NewResult(1, 1))
	m.ExpectCommit()
	status := multipartCall(t, (&KinerjaDiklatHandler{DB: db}).UpdateRiwayatDiklat, map[string]string{"nama_diklat": "Updated", "penyelenggara": "Org", "jumlah_jp": "8", "tanggal_mulai": "2025-01-01", "tanggal_selesai": "2025-01-02"}, "", "", nil)
	if status != 200 {
		t.Errorf("got %d", status)
	}
	if e := m.ExpectationsWereMet(); e != nil {
		t.Error(e)
	}
}

func TestDiklatEditDeleteRejectOtherEmployee(t *testing.T) {
	for _, method := range []string{"PUT", "DELETE"} {
		db, m := testDB(t)
		m.ExpectQuery("SELECT .*FROM sikapj.riwayat_diklat").WillReturnRows(sqlmock.NewRows([]string{"id", "nip"}).AddRow(1, "199003202014022099"))
		h := &KinerjaDiklatHandler{DB: db}
		a := fiber.New()
		a.Use(func(c *fiber.Ctx) error {
			c.Locals("role", "PEGAWAI")
			c.Locals("nip", "199003202014022005")
			c.Locals("user_id", 1)
			return c.Next()
		})
		a.Put("/:id", h.UpdateRiwayatDiklat)
		a.Delete("/:id", h.DeleteRiwayatDiklat)
		r := httptest.NewRequest(method, "/1", nil)
		res, e := a.Test(r)
		if e != nil {
			t.Fatal(e)
		}
		res.Body.Close()
		if res.StatusCode != 403 {
			t.Errorf("got %d", res.StatusCode)
		}
		if e = m.ExpectationsWereMet(); e != nil {
			t.Error(e)
		}
	}
}
