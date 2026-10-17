package services_test

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"sikapj-backend/internal/config"
	"sikapj-backend/internal/handlers"
)

const requiredPredicate = `(CASE WHEN m.id = 9 THEN p.jenis_jabatan = 'Fungsional' ELSE m.is_required END)`

func dashboard(t *testing.T) (*fiber.App, sqlmock.Sqlmock) {
	t.Helper()
	raw, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { raw.Close() })
	db := sqlx.NewDb(raw, "sqlmock")
	h := &handlers.DashboardAlertHandler{DB: db, Cfg: &config.Config{AlertThresholdMonths: 3}}
	a := fiber.New()
	a.Get("/dashboard", h.GetDashboardSummary)
	a.Get("/alerts", h.GetAlertList)
	a.Get("/self", func(c *fiber.Ctx) error { c.Locals("nip", "123456789012345678"); return h.GetPegawaiSelfDashboard(c) })
	admin := &handlers.AdminHandler{DB: db}
	a.Get("/export", admin.ExportEmployees)
	return a, mock
}

func TestDashboardReturnsThresholdAndUsesActiveCompleteness(t *testing.T) {
	app, mock := dashboard(t)
	mock.ExpectQuery("SELECT value").WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("4"))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM sikapj.pegawai WHERE is_active = TRUE`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery("SELECT COUNT.*e_berkas").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT COUNT.*sikapj.pegawai p WHERE p.is_active.*" + regexp.QuoteMeta(requiredPredicate)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT \\* FROM sikapj.pegawai").WillReturnRows(sqlmock.NewRows([]string{"nip"}))
	res, err := app.Test(httptest.NewRequest("GET", "/dashboard", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("status=%d", res.StatusCode)
	}
	var out struct {
		Data struct {
			Threshold int     `json:"threshold_months"`
			Percent   float64 `json:"berkas_lengkap_persen"`
			Alerts    []any   `json:"alerts"`
		}
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Data.Threshold != 4 || out.Data.Percent != 50 || out.Data.Alerts == nil {
		t.Fatalf("unexpected data: %+v", out.Data)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDashboardFailsClosedOnQueryErrors(t *testing.T) {
	for _, route := range []string{"/dashboard", "/alerts", "/self"} {
		t.Run(route, func(t *testing.T) {
			app, mock := dashboard(t)
			mock.ExpectQuery("SELECT").WillReturnError(errors.New("offline"))
			res, err := app.Test(httptest.NewRequest("GET", route, nil))
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != 500 {
				t.Fatalf("status=%d want 500", res.StatusCode)
			}
		})
	}
}

func TestSelfAndExportShareRequiredDocumentPredicate(t *testing.T) {
	for _, route := range []string{"/self", "/export?format=csv"} {
		t.Run(route, func(t *testing.T) {
			app, mock := dashboard(t)
			if route == "/self" {
				mock.ExpectQuery("SELECT \\* FROM sikapj.pegawai").WillReturnRows(sqlmock.NewRows([]string{"nip", "jenis_jabatan", "status_pegawai"}).AddRow("123456789012345678", "Fungsional", "PPPK"))
				mock.ExpectQuery("SELECT COUNT.*" + regexp.QuoteMeta(requiredPredicate)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(12))
				mock.ExpectQuery("SELECT COUNT.*" + regexp.QuoteMeta(requiredPredicate)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(6))
				mock.ExpectQuery("SELECT value").WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("3"))
			} else {
				mock.ExpectQuery("SELECT p.nip.*" + regexp.QuoteMeta(requiredPredicate) + ".*" + regexp.QuoteMeta(requiredPredicate)).WillReturnRows(sqlmock.NewRows([]string{"nip"}))
			}
			res, err := app.Test(httptest.NewRequest("GET", route, nil))
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != 200 {
				t.Fatalf("status=%d want 200", res.StatusCode)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
