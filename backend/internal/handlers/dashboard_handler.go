package handlers

import (
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

// GetDashboardSummary mengembalikan metrik aggregat dashboard admin
func (h *DashboardAlertHandler) GetDashboardSummary(c *fiber.Ctx) error {
	// 1. Ambil setting threshold N bulan
	thresholdMonths := h.Cfg.AlertThresholdMonths
	var settingVal string
	if err := h.DB.Get(&settingVal, "SELECT value FROM sikapj.settings WHERE key = 'ALERT_N_MONTHS'"); err == nil {
		if val, err := strconv.Atoi(settingVal); err == nil {
			thresholdMonths = val
		}
	}

	// 2. Total Pegawai Aktif
	var totalPegawai int
	_ = h.DB.Get(&totalPegawai, "SELECT COUNT(*) FROM sikapj.pegawai WHERE is_active = TRUE")

	// 3. Menunggu Verifikasi Berkas
	var menungguVerifikasi int
	_ = h.DB.Get(&menungguVerifikasi, "SELECT COUNT(*) FROM sikapj.e_berkas WHERE status_verifikasi = 'Menunggu Verifikasi'")

	// 4. Hitung Persentase Berkas Lengkap
	// Berkas lengkap didefinisikan jika pegawai memiliki seluruh 11 dokumen required berstatus 'Disetujui'
	var pegawaiLengkapCount int
	queryLengkap := `
		SELECT COUNT(*) FROM (
			SELECT b.nip FROM sikapj.e_berkas b
			JOIN sikapj.master_dokumen m ON b.dokumen_id = m.id
			WHERE m.is_required = TRUE AND b.status_verifikasi = 'Disetujui'
			GROUP BY b.nip
			HAVING COUNT(DISTINCT b.dokumen_id) = (SELECT COUNT(*) FROM sikapj.master_dokumen WHERE is_required = TRUE)
		) sub
	`
	_ = h.DB.Get(&pegawaiLengkapCount, queryLengkap)
	var persenLengkap float64 = 0
	if totalPegawai > 0 {
		persenLengkap = (float64(pegawaiLengkapCount) / float64(totalPegawai)) * 100
	}

	// 5. Evaluasi Proyeksi Alert KGB & KP untuk Seluruh Pegawai
	var allPegawai []models.Pegawai
	_ = h.DB.Select(&allPegawai, "SELECT * FROM sikapj.pegawai WHERE is_active = TRUE")

	var allAlerts []models.AlertItem
	alertMendekatiCount := 0

	for _, p := range allPegawai {
		alerts := services.BuildPegawaiAlertItems(&p, thresholdMonths)
		for _, a := range alerts {
			if a.Status == "SEGERA" || a.Status == "DIINGATKAN" {
				alertMendekatiCount++
			}
			allAlerts = append(allAlerts, a)
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": models.DashboardSummary{
			TotalPegawai:        totalPegawai,
			BerkasLengkapPersen: persenLengkap,
			MenungguVerifikasi:  menungguVerifikasi,
			AlertMendekati:      alertMendekatiCount,
			Alerts:              allAlerts,
		},
	})
}

// GetAlertList khusus mengambil list tabel alert KGB/KP
func (h *DashboardAlertHandler) GetAlertList(c *fiber.Ctx) error {
	thresholdMonths := h.Cfg.AlertThresholdMonths
	var settingVal string
	if err := h.DB.Get(&settingVal, "SELECT value FROM sikapj.settings WHERE key = 'ALERT_N_MONTHS'"); err == nil {
		if val, err := strconv.Atoi(settingVal); err == nil {
			thresholdMonths = val
		}
	}

	var allPegawai []models.Pegawai
	_ = h.DB.Select(&allPegawai, "SELECT * FROM sikapj.pegawai WHERE is_active = TRUE ORDER BY nama_lengkap ASC")

	var alerts []models.AlertItem
	for _, p := range allPegawai {
		items := services.BuildPegawaiAlertItems(&p, thresholdMonths)
		alerts = append(alerts, items...)
	}

	return c.JSON(fiber.Map{
		"success":          true,
		"threshold_months": thresholdMonths,
		"data":             alerts,
	})
}

// UpdateAlertThreshold mengubah parameter N bulan
func (h *DashboardAlertHandler) UpdateAlertThreshold(c *fiber.Ctx) error {
	type ThresholdReq struct {
		ThresholdMonths int `json:"threshold_months"`
	}

	var req ThresholdReq
	if err := c.BodyParser(&req); err != nil || req.ThresholdMonths <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Nilai bulan pengingat harus angka positif",
		})
	}

	_, err := h.DB.Exec(
		"INSERT INTO sikapj.settings (key, value, deskripsi) VALUES ('ALERT_N_MONTHS', $1, 'Jangka waktu ambang batas pengingat') ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()",
		strconv.Itoa(req.ThresholdMonths),
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menyimpan pengaturan",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Konfigurasi waktu pengingat berhasil disimpan",
		"data": fiber.Map{
			"threshold_months": req.ThresholdMonths,
		},
	})
}

// GetPegawaiSelfDashboard data dashboard ringkas untuk portal mandiri pegawai
func (h *DashboardAlertHandler) GetPegawaiSelfDashboard(c *fiber.Ctx) error {
	nip := c.Locals("nip").(string)

	var p models.Pegawai
	if err := h.DB.Get(&p, "SELECT * FROM sikapj.pegawai WHERE nip = $1", nip); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Data tidak ditemukan"})
	}

	// Hitung kelengkapan berkas pegawai ini
	var totalRequired int
	_ = h.DB.Get(&totalRequired, "SELECT COUNT(*) FROM sikapj.master_dokumen WHERE is_required = TRUE")

	var disetujuiCount int
	_ = h.DB.Get(&disetujuiCount, `
		SELECT COUNT(DISTINCT b.dokumen_id) 
		FROM sikapj.e_berkas b 
		JOIN sikapj.master_dokumen m ON b.dokumen_id = m.id 
		WHERE b.nip = $1 AND b.status_verifikasi = 'Disetujui' AND m.is_required = TRUE
	`, nip)

	persen := 0.0
	if totalRequired > 0 {
		persen = (float64(disetujuiCount) / float64(totalRequired)) * 100
	}

	// Alert milik pegawai ini
	alerts := services.BuildPegawaiAlertItems(&p, h.Cfg.AlertThresholdMonths)

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"pegawai":             p,
			"kelengkapan_berkas": persen,
			"alerts":              alerts,
		},
	})
}
