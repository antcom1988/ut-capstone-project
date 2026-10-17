package handlers

import (
	"database/sql"
	"sikapj-backend/internal/config"
	"sikapj-backend/internal/models"
	"sikapj-backend/internal/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type DashboardAlertHandler struct {
	DB  *sqlx.DB
	Cfg *config.Config
}

// GetDashboardSummary counts completeness only within the active population.
func (h *DashboardAlertHandler) GetDashboardSummary(c *fiber.Ctx) error {
	thresholdMonths, err := services.AlertThreshold(h.DB, h.Cfg.AlertThresholdMonths)
	if err != nil {
		return internalError(c)
	}
	var totalPegawai int
	if err := h.DB.Get(&totalPegawai, "SELECT COUNT(*) FROM sikapj.pegawai WHERE is_active = TRUE"); err != nil {
		return internalError(c)
	}
	var menungguVerifikasi int
	if err := h.DB.Get(&menungguVerifikasi, "SELECT COUNT(*) FROM sikapj.e_berkas WHERE status_verifikasi = 'Menunggu Verifikasi'"); err != nil {
		return internalError(c)
	}
	var pegawaiLengkapCount int
	// NOT EXISTS avoids duplicate uploads inflating completeness. Zero required
	// documents count as zero completeness, matching ESS and export percentages.
	queryLengkap := `SELECT COUNT(*) FROM sikapj.pegawai p WHERE p.is_active = TRUE
 AND EXISTS (SELECT 1 FROM sikapj.master_dokumen m WHERE ` + services.RequiredDocumentPredicateSQL + `)
 AND NOT EXISTS (SELECT 1 FROM sikapj.master_dokumen m WHERE ` + services.RequiredDocumentPredicateSQL + `
 AND NOT EXISTS (SELECT 1 FROM sikapj.e_berkas b WHERE b.nip=p.nip AND b.dokumen_id=m.id AND b.status_verifikasi='Disetujui'))`
	if err := h.DB.Get(&pegawaiLengkapCount, queryLengkap); err != nil {
		return internalError(c)
	}
	persenLengkap := 0.0
	if totalPegawai > 0 {
		persenLengkap = float64(pegawaiLengkapCount) * 100 / float64(totalPegawai)
	}
	allPegawai := []models.Pegawai{}
	if err := h.DB.Select(&allPegawai, "SELECT * FROM sikapj.pegawai WHERE is_active = TRUE"); err != nil {
		return internalError(c)
	}
	allAlerts := []models.AlertItem{}
	alertMendekatiCount := 0
	for _, p := range allPegawai {
		for _, a := range services.BuildPegawaiAlertItems(&p, thresholdMonths) {
			if a.Status == "SEGERA" || a.Status == "DIINGATKAN" {
				alertMendekatiCount++
			}
			allAlerts = append(allAlerts, a)
		}
	}
	return c.JSON(fiber.Map{"success": true, "data": models.DashboardSummary{
		ThresholdMonths: thresholdMonths, TotalPegawai: totalPegawai, BerkasLengkapPersen: persenLengkap,
		MenungguVerifikasi: menungguVerifikasi, AlertMendekati: alertMendekatiCount, Alerts: allAlerts,
	}})
}

func (h *DashboardAlertHandler) GetAlertList(c *fiber.Ctx) error {
	thresholdMonths, err := services.AlertThreshold(h.DB, h.Cfg.AlertThresholdMonths)
	if err != nil {
		return internalError(c)
	}
	allPegawai := []models.Pegawai{}
	if err := h.DB.Select(&allPegawai, "SELECT * FROM sikapj.pegawai WHERE is_active = TRUE ORDER BY nama_lengkap ASC"); err != nil {
		return internalError(c)
	}
	alerts := []models.AlertItem{}
	for _, p := range allPegawai {
		alerts = append(alerts, services.BuildPegawaiAlertItems(&p, thresholdMonths)...)
	}
	return c.JSON(fiber.Map{"success": true, "threshold_months": thresholdMonths, "data": alerts})
}

// UpdateAlertThreshold mengubah parameter N bulan.
func (h *DashboardAlertHandler) UpdateAlertThreshold(c *fiber.Ctx) error {
	type ThresholdReq struct {
		ThresholdMonths int `json:"threshold_months"`
	}
	var req ThresholdReq
	if err := c.BodyParser(&req); err != nil || req.ThresholdMonths <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Nilai bulan pengingat harus angka positif"})
	}
	_, err := h.DB.Exec("INSERT INTO sikapj.settings (key, value, deskripsi) VALUES ('ALERT_N_MONTHS', $1, 'Jangka waktu ambang batas pengingat') ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()", strconv.Itoa(req.ThresholdMonths))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal menyimpan pengaturan"})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Konfigurasi waktu pengingat berhasil disimpan", "data": fiber.Map{"threshold_months": req.ThresholdMonths}})
}

// GetPegawaiSelfDashboard retains alerts for the employee as required by PRD.
func (h *DashboardAlertHandler) GetPegawaiSelfDashboard(c *fiber.Ctx) error {
	nip := c.Locals("nip").(string)
	var p models.Pegawai
	if err := h.DB.Get(&p, "SELECT * FROM sikapj.pegawai WHERE nip = $1", nip); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Data tidak ditemukan"})
		}
		return internalError(c)
	}
	var totalRequired int
	if err := h.DB.Get(&totalRequired, `SELECT COUNT(*) FROM sikapj.master_dokumen m JOIN sikapj.pegawai p ON p.nip=$1 WHERE `+services.RequiredDocumentPredicateSQL, nip); err != nil {
		return internalError(c)
	}
	var disetujuiCount int
	if err := h.DB.Get(&disetujuiCount, `SELECT COUNT(DISTINCT b.dokumen_id) FROM sikapj.e_berkas b
 JOIN sikapj.master_dokumen m ON m.id=b.dokumen_id JOIN sikapj.pegawai p ON p.nip=b.nip
 WHERE b.nip=$1 AND b.status_verifikasi='Disetujui' AND `+services.RequiredDocumentPredicateSQL, nip); err != nil {
		return internalError(c)
	}
	persen := 0.0
	if totalRequired > 0 {
		persen = float64(disetujuiCount) * 100 / float64(totalRequired)
	}
	thresholdMonths, err := services.AlertThreshold(h.DB, h.Cfg.AlertThresholdMonths)
	if err != nil {
		return internalError(c)
	}
	alerts := services.BuildPegawaiAlertItems(&p, thresholdMonths)
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"pegawai": p, "kelengkapan_berkas": persen, "threshold_months": thresholdMonths, "alerts": alerts}})
}
