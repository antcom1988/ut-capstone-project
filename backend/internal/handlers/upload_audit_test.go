package handlers

import (
	"bytes"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sikapj-backend/internal/config"
	"testing"
)

func multipartCall(t *testing.T, h fiber.Handler, fields map[string]string, key, name string, data []byte) int {
	t.Helper()
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for k, v := range fields {
		w.WriteField(k, v)
	}
	if key != "" {
		f, _ := w.CreateFormFile(key, name)
		f.Write(data)
	}
	w.Close()
	r := httptest.NewRequest("POST", "/", &b)
	r.Header.Set("Content-Type", w.FormDataContentType())
	res, e := testApp(h).Test(r)
	if e != nil {
		t.Fatal(e)
	}
	defer res.Body.Close()
	return res.StatusCode
}
func TestUploadAuditFailureRemovesNewFile(t *testing.T) {
	for _, kind := range []string{"berkas", "kinerja", "diklat"} {
		t.Run(kind, func(t *testing.T) {
			db, m := testDB(t)
			dir := t.TempDir()
			cfg := &config.Config{UploadDir: dir}
			fields := map[string]string{"nip": "199003202014022005", "dokumen_id": "1", "tahun_periode": "2025", "jenis_dokumen": "PK", "nama_diklat": "Course", "penyelenggara": "Org", "jumlah_jp": "8", "tanggal_mulai": "2025-01-01", "tanggal_selesai": "2025-01-02"}
			var h fiber.Handler
			key := "file"
			table := "sikapj.e_berkas"
			switch kind {
			case "berkas":
				h = (&BerkasHandler{DB: db, Cfg: cfg}).UploadBerkas
				m.ExpectQuery("SELECT id, nama_dokumen").WillReturnRows(sqlmock.NewRows([]string{"id", "nama_dokumen", "max_size_mb"}).AddRow(1, "KTP", 1))
			case "kinerja":
				h = (&KinerjaDiklatHandler{DB: db, Cfg: cfg}).UploadArsipKinerja
				table = "sikapj.arsip_kinerja"
			case "diklat":
				h = (&KinerjaDiklatHandler{DB: db, Cfg: cfg}).CreateRiwayatDiklat
				table = "sikapj.riwayat_diklat"
				key = "file_sertifikat"
			}
			if kind != "diklat" {
				m.ExpectQuery("SELECT file_path").WillReturnRows(sqlmock.NewRows([]string{"file_path"}))
			}
			m.ExpectBegin()
			m.ExpectExec("INSERT INTO " + table).WillReturnResult(sqlmock.NewResult(1, 1))
			m.ExpectExec("INSERT INTO sikapj.audit_log").WillReturnError(errors.New("audit unavailable"))
			m.ExpectRollback()
			if status := multipartCall(t, h, fields, key, "file.pdf", []byte("%PDF-1.4\n%%EOF")); status != 500 {
				t.Fatalf("got %d", status)
			}
			if e := m.ExpectationsWereMet(); e != nil {
				t.Error(e)
			}
			filepath.Walk(dir, func(p string, i os.FileInfo, e error) error {
				if e == nil && !i.IsDir() {
					t.Errorf("orphan file %s", p)
				}
				return e
			})
		})
	}
}
