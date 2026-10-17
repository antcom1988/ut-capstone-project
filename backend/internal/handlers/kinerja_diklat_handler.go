package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"sikapj-backend/internal/audit"
	"sikapj-backend/internal/config"
	"sikapj-backend/internal/models"
	"sikapj-backend/internal/validation"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type KinerjaDiklatHandler struct {
	DB  *sqlx.DB
	Cfg *config.Config
}

// GetArsipKinerjaByNIP mengambil daftar dokumen PK & SKP tahunan milik pegawai
func (h *KinerjaDiklatHandler) GetArsipKinerjaByNIP(c *fiber.Ctx) error {
	nip := c.Params("nip")
	currentRole := c.Locals("role").(string)
	currentNIP := c.Locals("nip").(string)

	if currentRole == "PEGAWAI" && currentNIP != nip {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Akses ditolak",
		})
	}

	var list []models.ArsipKinerja
	query := `SELECT id, nip, tahun_periode, jenis_dokumen, file_path, file_name, file_size_bytes, uploaded_at 
	          FROM sikapj.arsip_kinerja WHERE nip = $1 ORDER BY tahun_periode DESC, jenis_dokumen ASC`
	if err := h.DB.Select(&list, query, nip); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil arsip kinerja: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    list,
	})
}

// UploadArsipKinerja mengunggah dokumen PK atau SKP untuk tahun tertentu
func (h *KinerjaDiklatHandler) UploadArsipKinerja(c *fiber.Ctx) error {
	currentRole := c.Locals("role").(string)
	currentNIP := c.Locals("nip").(string)

	nip := c.FormValue("nip")
	if currentRole == "PEGAWAI" {
		nip = currentNIP
	}
	if !validation.ValidNIP(nip) {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "NIP harus berjumlah 18 digit angka"})
	}

	tahunStr := c.FormValue("tahun_periode")
	tahun, err := strconv.Atoi(tahunStr)
	if err != nil || tahun < 2000 || tahun > time.Now().Year()+1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Tahun periode tidak valid",
		})
	}

	jenisDokumen := strings.ToUpper(strings.TrimSpace(c.FormValue("jenis_dokumen")))
	if jenisDokumen != "PK" && jenisDokumen != "SKP" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Jenis dokumen harus PK atau SKP",
		})
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "File dokumen wajib disertakan",
		})
	}

	// Maksimal 1MB untuk berkas kinerja.
	if file.Size > 1024*1024 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Ukuran file kinerja melebihi batas maksimum 1 MB",
		})
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	opened, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "File tidak dapat dibaca"})
	}
	_, _, fileTypeErr := validation.DetectAllowedFileType(opened, ext)
	_ = opened.Close()
	if fileTypeErr != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Dokumen kinerja harus berupa PDF, JPG, atau PNG yang valid",
		})
	}

	destDir := filepath.Join(h.Cfg.UploadDir, "kinerja", nip)
	if err := os.MkdirAll(destDir, 0750); err != nil {
		return internalError(c)
	}
	savedName := fmt.Sprintf("%s_%d_%s%s", jenisDokumen, tahun, uuid.New().String()[:6], ext)
	savedPath := filepath.Join(destDir, savedName)

	if err := c.SaveFile(file, savedPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menyimpan file ke disk",
		})
	}

	query := `
		INSERT INTO sikapj.arsip_kinerja (nip, tahun_periode, jenis_dokumen, file_path, file_name, file_size_bytes, uploaded_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (nip, tahun_periode, jenis_dokumen)
		DO UPDATE SET file_path = EXCLUDED.file_path, file_name = EXCLUDED.file_name, file_size_bytes = EXCLUDED.file_size_bytes, uploaded_at = NOW()
	`
	var oldPath string
	_ = h.DB.Get(&oldPath, `SELECT file_path FROM sikapj.arsip_kinerja WHERE nip=$1 AND tahun_periode=$2 AND jenis_dokumen=$3`, nip, tahun, jenisDokumen)
	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(savedPath)
		}
	}()
	tx, err := h.DB.Beginx()
	if err != nil {
		return internalError(c)
	}
	defer tx.Rollback()
	if _, err := tx.Exec(query, nip, tahun, jenisDokumen, savedPath, filepath.Base(file.Filename), file.Size); err != nil {
		_ = os.Remove(savedPath)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal menyimpan data arsip kinerja"})
	}
	actor := c.Locals("user_id").(int)
	if err = (audit.Service{DB: tx}).Log(&actor, "UPLOAD_KINERJA", map[string]any{"nip": nip, "tahun_periode": tahun, "jenis_dokumen": jenisDokumen}, c.IP()); err != nil {
		return internalError(c)
	}
	if err = tx.Commit(); err != nil {
		return internalError(c)
	}
	committed = true
	if oldPath != "" && oldPath != savedPath {
		_ = os.Remove(oldPath)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Dokumen kinerja berhasil disimpan",
	})
}

// StreamPreviewKinerja melihat preview PDF kinerja inline tanpa download
func (h *KinerjaDiklatHandler) StreamPreviewKinerja(c *fiber.Ctx) error {
	id := c.Params("id")
	var doc models.ArsipKinerja
	if err := h.DB.Get(&doc, "SELECT id, nip, file_path, file_name FROM sikapj.arsip_kinerja WHERE id = $1", id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Dokumen tidak ditemukan"})
	}
	if !validation.CanAccessNIP(c.Locals("role").(string), c.Locals("nip").(string), doc.NIP) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": "Akses ditolak"})
	}

	switch strings.ToLower(filepath.Ext(doc.FilePath)) {
	case ".png":
		c.Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		c.Set("Content-Type", "image/jpeg")
	default:
		c.Set("Content-Type", "application/pdf")
	}
	c.Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, doc.FileName))
	return c.SendFile(doc.FilePath)
}

// GetRiwayatDiklatByNIP mengambil daftar rekam jejak pelatihan
func (h *KinerjaDiklatHandler) GetRiwayatDiklatByNIP(c *fiber.Ctx) error {
	nip := c.Params("nip")
	if !validation.CanAccessNIP(c.Locals("role").(string), c.Locals("nip").(string), nip) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": "Akses ditolak"})
	}
	var list []models.RiwayatDiklat
	query := `SELECT id, nip, nama_diklat, penyelenggara, jumlah_jp, tanggal_mulai, tanggal_selesai, file_sertifikat_path, file_name, created_at 
	          FROM sikapj.riwayat_diklat WHERE nip = $1 ORDER BY tanggal_mulai DESC`
	if err := h.DB.Select(&list, query, nip); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data diklat",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    list,
	})
}

// CreateRiwayatDiklat menambahkan rekaman diklat mandiri + sertifikat
func (h *KinerjaDiklatHandler) CreateRiwayatDiklat(c *fiber.Ctx) error {
	existing, _ := c.Locals("editing_diklat").(*models.RiwayatDiklat)
	currentRole := c.Locals("role").(string)
	currentNIP := c.Locals("nip").(string)

	nip := c.FormValue("nip")
	if currentRole == "PEGAWAI" {
		nip = currentNIP
	}
	if existing != nil {
		nip = existing.NIP
	}
	if !validation.ValidNIP(nip) {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "NIP harus berjumlah 18 digit angka"})
	}

	namaDiklat := strings.TrimSpace(c.FormValue("nama_diklat"))
	penyelenggara := strings.TrimSpace(c.FormValue("penyelenggara"))
	jumlahJP, jpErr := strconv.Atoi(c.FormValue("jumlah_jp"))
	if jpErr != nil || jumlahJP <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Jumlah JP harus berupa angka positif"})
	}
	tglMulaiStr := c.FormValue("tanggal_mulai")
	tglSelesaiStr := c.FormValue("tanggal_selesai")

	if namaDiklat == "" || penyelenggara == "" || tglMulaiStr == "" || tglSelesaiStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Semua field diklat (Nama, Penyelenggara, Tanggal) wajib diisi",
		})
	}

	tglMulai, err1 := time.Parse("2006-01-02", tglMulaiStr)
	tglSelesai, err2 := time.Parse("2006-01-02", tglSelesaiStr)
	if err1 != nil || err2 != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format tanggal harus YYYY-MM-DD",
		})
	}
	if tglSelesai.Before(tglMulai) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Tanggal selesai tidak boleh sebelum tanggal mulai"})
	}

	// Handle unggah sertifikat jika ada
	var certPath *string
	var certName *string
	file, err := c.FormFile("file_sertifikat")
	if existing != nil {
		certPath = existing.FileSertifikatPath
		certName = existing.FileName
	}
	newCertificate := err == nil && file != nil
	if !newCertificate && certPath == nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Sertifikat diklat wajib disertakan"})
	}
	if newCertificate {
		if file.Size > 2*1024*1024 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Ukuran sertifikat melebihi batas maksimum 2 MB"})
		}
		ext := strings.ToLower(filepath.Ext(file.Filename))
		opened, openErr := file.Open()
		if openErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Sertifikat tidak dapat dibaca"})
		}
		_, _, typeErr := validation.DetectAllowedFileType(opened, ext)
		_ = opened.Close()
		if typeErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Sertifikat harus berupa PDF, JPG, atau PNG yang valid"})
		}
		destDir := filepath.Join(h.Cfg.UploadDir, "diklat", nip)
		if err := os.MkdirAll(destDir, 0750); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal membuat direktori sertifikat"})
		}
		saved := filepath.Join(destDir, fmt.Sprintf("cert_%s%s", uuid.New().String()[:8], ext))
		if err := c.SaveFile(file, saved); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal menyimpan sertifikat"})
		}
		certPath = &saved
		certName = &file.Filename
	}

	query := `
		INSERT INTO sikapj.riwayat_diklat (nip, nama_diklat, penyelenggara, jumlah_jp, tanggal_mulai, tanggal_selesai, file_sertifikat_path, file_name)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	committed := false
	defer func() {
		if !committed && newCertificate && certPath != nil {
			_ = os.Remove(*certPath)
		}
	}()
	tx, err := h.DB.Beginx()
	if err != nil {
		return internalError(c)
	}
	defer tx.Rollback()
	if existing == nil {
		_, err = tx.Exec(query, nip, namaDiklat, penyelenggara, jumlahJP, tglMulai, tglSelesai, certPath, certName)
	} else {
		result, updateErr := tx.Exec(`UPDATE sikapj.riwayat_diklat SET nama_diklat=$1,penyelenggara=$2,jumlah_jp=$3,tanggal_mulai=$4,tanggal_selesai=$5,file_sertifikat_path=$6,file_name=$7 WHERE id=$8 AND nip=$9 AND file_sertifikat_path IS NOT DISTINCT FROM $10`, namaDiklat, penyelenggara, jumlahJP, tglMulai, tglSelesai, certPath, certName, existing.ID, existing.NIP, existing.FileSertifikatPath)
		err = updateErr
		if err == nil {
			n, rowsErr := result.RowsAffected()
			if rowsErr != nil {
				return internalError(c)
			}
			if n != 1 {
				return c.Status(409).JSON(fiber.Map{"success": false, "message": "Diklat berubah; muat ulang data"})
			}
		}
	}
	if err != nil {
		if newCertificate && certPath != nil {
			_ = os.Remove(*certPath)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal menyimpan riwayat diklat"})
	}

	actor := c.Locals("user_id").(int)
	action := "CREATE_DIKLAT"
	if existing != nil {
		action = "UPDATE_DIKLAT"
	}
	if err = (audit.Service{DB: tx}).Log(&actor, action, map[string]string{"nip": nip}, c.IP()); err != nil {
		return internalError(c)
	}
	if err = tx.Commit(); err != nil {
		return internalError(c)
	}
	committed = true
	if existing != nil && newCertificate && existing.FileSertifikatPath != nil {
		_ = os.Remove(*existing.FileSertifikatPath)
	}
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Data diklat berhasil disimpan",
	})
}

// UpdateRiwayatDiklat preserves ownership; an omitted certificate retains the current one.
func (h *KinerjaDiklatHandler) UpdateRiwayatDiklat(c *fiber.Ctx) error {
	var existing models.RiwayatDiklat
	if err := h.DB.Get(&existing, `SELECT id,nip,file_sertifikat_path,file_name FROM sikapj.riwayat_diklat WHERE id=$1`, c.Params("id")); err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "message": "Diklat tidak ditemukan"})
	}
	if !validation.CanAccessNIP(c.Locals("role").(string), c.Locals("nip").(string), existing.NIP) {
		return c.Status(403).JSON(fiber.Map{"success": false, "message": "Akses ditolak"})
	}
	c.Locals("editing_diklat", &existing)
	return h.CreateRiwayatDiklat(c)
}
func (h *KinerjaDiklatHandler) DeleteRiwayatDiklat(c *fiber.Ctx) error {
	var existing models.RiwayatDiklat
	if err := h.DB.Get(&existing, `SELECT id,nip,file_sertifikat_path FROM sikapj.riwayat_diklat WHERE id=$1`, c.Params("id")); err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "message": "Diklat tidak ditemukan"})
	}
	if !validation.CanAccessNIP(c.Locals("role").(string), c.Locals("nip").(string), existing.NIP) {
		return c.Status(403).JSON(fiber.Map{"success": false, "message": "Akses ditolak"})
	}
	tx, err := h.DB.Beginx()
	if err != nil {
		return internalError(c)
	}
	defer tx.Rollback()
	var deletedPath *string
	if err = tx.QueryRow(`DELETE FROM sikapj.riwayat_diklat WHERE id=$1 AND nip=$2 RETURNING file_sertifikat_path`, existing.ID, existing.NIP).Scan(&deletedPath); err != nil {
		return internalError(c)
	}
	actor := c.Locals("user_id").(int)
	if err = (audit.Service{DB: tx}).Log(&actor, "DELETE_DIKLAT", map[string]any{"nip": existing.NIP, "id": existing.ID}, c.IP()); err != nil {
		return internalError(c)
	}
	if err = tx.Commit(); err != nil {
		return internalError(c)
	}
	if deletedPath != nil {
		_ = os.Remove(*deletedPath)
	}
	return c.JSON(fiber.Map{"success": true, "message": "Diklat berhasil dihapus"})
}

// StreamPreviewSertifikat melihat file sertifikat diklat inline di browser
func (h *KinerjaDiklatHandler) StreamPreviewSertifikat(c *fiber.Ctx) error {
	id := c.Params("id")
	var diklat models.RiwayatDiklat
	if err := h.DB.Get(&diklat, "SELECT id, nip, file_sertifikat_path, file_name FROM sikapj.riwayat_diklat WHERE id = $1", id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Data tidak ditemukan"})
	}

	if !validation.CanAccessNIP(c.Locals("role").(string), c.Locals("nip").(string), diklat.NIP) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": "Akses ditolak"})
	}
	if diklat.FileSertifikatPath == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Sertifikat belum diunggah"})
	}

	ext := strings.ToLower(filepath.Ext(*diklat.FileSertifikatPath))
	if ext == ".pdf" {
		c.Set("Content-Type", "application/pdf")
	} else if ext == ".png" {
		c.Set("Content-Type", "image/png")
	} else {
		c.Set("Content-Type", "image/jpeg")
	}

	fileName := "sertifikat" + ext
	if diklat.FileName != nil {
		fileName = *diklat.FileName
	}
	c.Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, fileName))
	return c.SendFile(*diklat.FileSertifikatPath)
}
