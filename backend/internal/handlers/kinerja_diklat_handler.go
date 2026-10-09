package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"sikapj-backend/internal/config"
	"sikapj-backend/internal/models"
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

	tahunStr := c.FormValue("tahun_periode")
	tahun, err := strconv.Atoi(tahunStr)
	if err != nil || tahun < 2000 {
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

	// Maksimal 2MB untuk berkas kinerja
	if file.Size > 2*1024*1024 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Ukuran file kinerja melebihi batas maksimum 2 MB",
		})
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".pdf" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Dokumen kinerja wajib dalam format PDF",
		})
	}

	destDir := filepath.Join(h.Cfg.UploadDir, "kinerja", nip)
	_ = os.MkdirAll(destDir, os.ModePerm)
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
	if _, err := h.DB.Exec(query, nip, tahun, jenisDokumen, savedPath, file.Filename, file.Size); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menyimpan data arsip kinerja: " + err.Error(),
		})
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

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, doc.FileName))
	return c.SendFile(doc.FilePath)
}

// GetRiwayatDiklatByNIP mengambil daftar rekam jejak pelatihan
func (h *KinerjaDiklatHandler) GetRiwayatDiklatByNIP(c *fiber.Ctx) error {
	nip := c.Params("nip")
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
	currentRole := c.Locals("role").(string)
	currentNIP := c.Locals("nip").(string)

	nip := c.FormValue("nip")
	if currentRole == "PEGAWAI" {
		nip = currentNIP
	}

	namaDiklat := strings.TrimSpace(c.FormValue("nama_diklat"))
	penyelenggara := strings.TrimSpace(c.FormValue("penyelenggara"))
	jumlahJP, _ := strconv.Atoi(c.FormValue("jumlah_jp"))
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

	// Handle unggah sertifikat jika ada
	var certPath *string
	var certName *string
	file, err := c.FormFile("file_sertifikat")
	if err == nil && file != nil {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		destDir := filepath.Join(h.Cfg.UploadDir, "diklat", nip)
		_ = os.MkdirAll(destDir, os.ModePerm)
		saved := filepath.Join(destDir, fmt.Sprintf("cert_%s%s", uuid.New().String()[:8], ext))
		if err := c.SaveFile(file, saved); err == nil {
			certPath = &saved
			certName = &file.Filename
		}
	}

	query := `
		INSERT INTO sikapj.riwayat_diklat (nip, nama_diklat, penyelenggara, jumlah_jp, tanggal_mulai, tanggal_selesai, file_sertifikat_path, file_name)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = h.DB.Exec(query, nip, namaDiklat, penyelenggara, jumlahJP, tglMulai, tglSelesai, certPath, certName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menyimpan riwayat diklat: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Data diklat berhasil disimpan",
	})
}

// StreamPreviewSertifikat melihat file sertifikat diklat inline di browser
func (h *KinerjaDiklatHandler) StreamPreviewSertifikat(c *fiber.Ctx) error {
	id := c.Params("id")
	var diklat models.RiwayatDiklat
	if err := h.DB.Get(&diklat, "SELECT id, file_sertifikat_path, file_name FROM sikapj.riwayat_diklat WHERE id = $1", id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Data tidak ditemukan"})
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
