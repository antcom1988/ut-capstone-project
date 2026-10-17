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

type BerkasHandler struct {
	DB  *sqlx.DB
	Cfg *config.Config
}

// GetMasterDokumen mengembalikan 12 jenis dokumen wajib
func (h *BerkasHandler) GetMasterDokumen(c *fiber.Ctx) error {
	var docs []models.MasterDokumen
	query := `SELECT id, nama_dokumen, max_size_mb, is_required, deskripsi 
	          FROM sikapj.master_dokumen ORDER BY id ASC`
	if err := h.DB.Select(&docs, query); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil master dokumen",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    docs,
	})
}

// GetBerkasByNIP mengambil status 12 e-berkas milik pegawai tertentu
func (h *BerkasHandler) GetBerkasByNIP(c *fiber.Ctx) error {
	nip := c.Params("nip")
	currentRole := c.Locals("role").(string)
	currentNIP := c.Locals("nip").(string)

	// Pastikan pegawai hanya bisa melihat berkasnya sendiri
	if currentRole == "PEGAWAI" && currentNIP != nip {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Anda tidak memiliki akses ke berkas pegawai lain",
		})
	}

	type BerkasStatusItem struct {
		DokumenID        int        `db:"dokumen_id" json:"dokumen_id"`
		NamaDokumen      string     `db:"nama_dokumen" json:"nama_dokumen"`
		MaxSizeMB        int        `db:"max_size_mb" json:"max_size_mb"`
		IsRequired       bool       `db:"is_required" json:"is_required"`
		BerkasID         *int       `db:"berkas_id" json:"berkas_id"`
		FileName         *string    `db:"file_name" json:"file_name"`
		FileSizeBytes    *int64     `db:"file_size_bytes" json:"file_size_bytes"`
		MimeType         *string    `db:"mime_type" json:"mime_type"`
		StatusVerifikasi string     `db:"status_verifikasi" json:"status_verifikasi"`
		CatatanAdmin     *string    `db:"catatan_admin" json:"catatan_admin"`
		UploadedAt       *time.Time `db:"uploaded_at" json:"uploaded_at"`
		VerifiedAt       *time.Time `db:"verified_at" json:"verified_at"`
	}

	query := `
		SELECT 
			m.id AS dokumen_id,
			m.nama_dokumen,
			m.max_size_mb,
			m.is_required,
			b.id AS berkas_id,
			b.file_name,
			b.file_size_bytes,
			b.mime_type,
			COALESCE(b.status_verifikasi, 'Belum Diunggah') AS status_verifikasi,
			b.catatan_admin,
			b.uploaded_at,
			b.verified_at
		FROM sikapj.master_dokumen m
		LEFT JOIN sikapj.e_berkas b ON m.id = b.dokumen_id AND b.nip = $1
		ORDER BY m.id ASC
	`

	var list []BerkasStatusItem
	if err := h.DB.Select(&list, query, nip); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil daftar e-berkas pegawai",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    list,
	})
}

// UploadBerkas menangani unggah atau unggah ulang e-berkas
func (h *BerkasHandler) UploadBerkas(c *fiber.Ctx) error {
	currentRole := c.Locals("role").(string)
	currentNIP := c.Locals("nip").(string)

	nip := c.FormValue("nip")
	if currentRole == "PEGAWAI" {
		nip = currentNIP // Paksa pakai NIP sendiri
	}
	if !validation.ValidNIP(nip) {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "NIP harus berjumlah 18 digit angka"})
	}

	dokumenIDStr := c.FormValue("dokumen_id")
	dokumenID, err := strconv.Atoi(dokumenIDStr)
	if err != nil || dokumenID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "dokumen_id tidak valid",
		})
	}

	// Cek batasan max file size dokumen dari DB
	var master models.MasterDokumen
	if err := h.DB.Get(&master, "SELECT id, nama_dokumen, max_size_mb FROM sikapj.master_dokumen WHERE id = $1", dokumenID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Jenis dokumen tidak ditemukan",
		})
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "File wajib disertakan",
		})
	}

	// Validasi Ukuran File (Bytes)
	maxSizeBytes := int64(master.MaxSizeMB) * 1024 * 1024
	if file.Size > maxSizeBytes {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Ukuran file melebihi batas maksimum %d MB untuk dokumen ini", master.MaxSizeMB),
		})
	}

	// Validasi Format Ekstensi File
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]string{
		".pdf":  "application/pdf",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
	}

	mimeType, allowed := allowedExts[ext]
	if !allowed {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format file tidak didukung. Harap unggah file PDF, JPG, atau PNG.",
		})
	}

	opened, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "File tidak dapat dibaca"})
	}
	detectedMime, _, typeErr := validation.DetectAllowedFileType(opened, ext)
	_ = opened.Close()
	if typeErr != nil || detectedMime != mimeType {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Isi file tidak sesuai dengan format PDF, JPG, atau PNG"})
	}

	// Buat direktori penyimpanan
	destDir := filepath.Join(h.Cfg.UploadDir, "berkas", nip)
	if err := os.MkdirAll(destDir, 0750); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat direktori penyimpanan berkas",
		})
	}

	// Nama file unik
	uniqueName := fmt.Sprintf("%d_%s%s", dokumenID, uuid.New().String()[:8], ext)
	savedPath := filepath.Join(destDir, uniqueName)

	if err := c.SaveFile(file, savedPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menyimpan file ke disk",
		})
	}

	// Simpan atau Perbarui ke tabel e_berkas (UPSERT)
	// Jika upload ulang, status kembali ke 'Menunggu Verifikasi'
	query := `
		INSERT INTO sikapj.e_berkas (nip, dokumen_id, file_path, file_name, file_size_bytes, mime_type, status_verifikasi, uploaded_at, verified_at, catatan_admin)
		VALUES ($1, $2, $3, $4, $5, $6, 'Menunggu Verifikasi', NOW(), NULL, NULL)
		ON CONFLICT (nip, dokumen_id) 
		DO UPDATE SET 
			file_path = EXCLUDED.file_path,
			file_name = EXCLUDED.file_name,
			file_size_bytes = EXCLUDED.file_size_bytes,
			mime_type = EXCLUDED.mime_type,
			status_verifikasi = 'Menunggu Verifikasi',
			catatan_admin = NULL,
			uploaded_at = NOW(),
			verified_at = NULL,
			verified_by = NULL
	`

	var oldPath string
	_ = h.DB.Get(&oldPath, "SELECT file_path FROM sikapj.e_berkas WHERE nip=$1 AND dokumen_id=$2", nip, dokumenID)
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
	if _, err := tx.Exec(query, nip, dokumenID, savedPath, filepath.Base(file.Filename), file.Size, mimeType); err != nil {
		_ = os.Remove(savedPath)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menyimpan data berkas ke database",
		})
	}
	actor := c.Locals("user_id").(int)
	if err = (audit.Service{DB: tx}).Log(&actor, "UPLOAD_BERKAS", map[string]any{"nip": nip, "dokumen_id": dokumenID}, c.IP()); err != nil {
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
		"message": "Berkas berhasil diunggah dan sedang menunggu verifikasi",
	})
}

// StreamPreviewFile menampilkan dokumen langsung di browser (PDF / Image) tanpa download
func (h *BerkasHandler) StreamPreviewFile(c *fiber.Ctx) error {
	berkasID := c.Params("id")

	var berkas models.EBerkas
	query := `SELECT id, nip, dokumen_id, file_path, file_name, mime_type FROM sikapj.e_berkas WHERE id = $1 LIMIT 1`
	if err := h.DB.Get(&berkas, query, berkasID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Berkas tidak ditemukan",
		})
	}

	// Validasi Hak Akses
	currentRole := c.Locals("role").(string)
	currentNIP := c.Locals("nip").(string)
	if currentRole == "PEGAWAI" && currentNIP != berkas.NIP {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Akses ditolak",
		})
	}

	// Set header agar ditampilkan INLINE di browser
	c.Set("Content-Type", berkas.MimeType)
	c.Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, berkas.FileName))

	return c.SendFile(berkas.FilePath)
}

// VerifikasiBerkas (Khusus ADMIN) - Setujui atau Tolak disertai Catatan
func (h *BerkasHandler) VerifikasiBerkas(c *fiber.Ctx) error {
	berkasID := c.Params("id")
	adminID := c.Locals("user_id").(int)

	type VerifyRequest struct {
		StatusVerifikasi string    `json:"status_verifikasi"` // "Disetujui" / "Ditolak"
		CatatanAdmin     *string   `json:"catatan_admin"`
		UploadedAt       time.Time `json:"uploaded_at"`
	}

	var req VerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Payload request tidak valid",
		})
	}

	if req.UploadedAt.IsZero() {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "uploaded_at wajib diisi sesuai versi berkas yang ditinjau"})
	}
	if req.StatusVerifikasi != "Disetujui" && req.StatusVerifikasi != "Ditolak" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Status verifikasi harus 'Disetujui' atau 'Ditolak'",
		})
	}

	if req.StatusVerifikasi == "Ditolak" && (req.CatatanAdmin == nil || strings.TrimSpace(*req.CatatanAdmin) == "") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Catatan admin wajib diisi apabila berkas ditolak",
		})
	}

	query := `
		UPDATE sikapj.e_berkas 
		SET status_verifikasi = $1, catatan_admin = $2, verified_at = NOW(), verified_by = $3
		WHERE id = $4 AND uploaded_at = $5
	`
	tx, err := h.DB.Beginx()
	if err != nil {
		return internalError(c)
	}
	defer tx.Rollback()
	result, err := tx.Exec(query, req.StatusVerifikasi, req.CatatanAdmin, adminID, berkasID, req.UploadedAt)
	if err != nil {
		return internalError(c)
	}
	rows, err := result.RowsAffected()
	if err != nil || rows != 1 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Berkas berubah atau tidak ditemukan; muat ulang dan tinjau versi terbaru"})
	}
	if _, err = tx.Exec(`INSERT INTO sikapj.audit_log(user_id,action,details,ip_address) VALUES($1,'VERIFIKASI_BERKAS',$2,$3)`, adminID, fmt.Sprintf(`{"berkas_id":%q,"status":%q}`, berkasID, req.StatusVerifikasi), c.IP()); err != nil {
		return internalError(c)
	}
	if err = tx.Commit(); err != nil {
		return internalError(c)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Status verifikasi berkas berhasil diperbarui",
	})
}
