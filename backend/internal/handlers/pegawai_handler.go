package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/mail"
	"sikapj-backend/internal/audit"
	"sikapj-backend/internal/auth"
	"sikapj-backend/internal/config"
	"sikapj-backend/internal/models"
	"sikapj-backend/internal/services"
	"sikapj-backend/internal/validation"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type PegawaiHandler struct {
	DB  *sqlx.DB
	Cfg *config.Config
}

// GetAllPegawai mengambil seluruh daftar pegawai untuk Admin
func (h *PegawaiHandler) GetAllPegawai(c *fiber.Ctx) error {
	var list []models.Pegawai
	query := `SELECT * FROM sikapj.pegawai ORDER BY is_active DESC, nama_lengkap ASC`
	if err := h.DB.Select(&list, query); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data pegawai",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    list,
	})
}

// GetPegawaiDetail mengambil detail 1 pegawai
func (h *PegawaiHandler) GetPegawaiDetail(c *fiber.Ctx) error {
	nip := c.Params("nip")
	var p models.Pegawai
	if err := h.DB.Get(&p, "SELECT * FROM sikapj.pegawai WHERE nip = $1", nip); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Pegawai tidak ditemukan",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    p,
	})
}

// CreatePegawai menambah data pegawai baru beserta akun loginnya
func (h *PegawaiHandler) CreatePegawai(c *fiber.Ctx) error {
	type CreatePegawaiRequest struct {
		NIP                 string  `json:"nip"`
		NamaLengkap         string  `json:"nama_lengkap"`
		TempatLahir         *string `json:"tempat_lahir"`
		TanggalLahir        string  `json:"tanggal_lahir"` // YYYY-MM-DD
		JenisKelamin        *string `json:"jenis_kelamin"`
		Email               *string `json:"email"`
		NoHP                *string `json:"no_hp"`
		StatusPegawai       string  `json:"status_pegawai"` // PNS / PPPK
		Bidang              string  `json:"bidang"`
		JenisJabatan        string  `json:"jenis_jabatan"`
		NamaJabatan         string  `json:"nama_jabatan"`
		Pangkat             string  `json:"pangkat"`
		Golongan            string  `json:"golongan"`
		TMTPengangkatan     string  `json:"tmt_pengangkatan"`
		TMTPangkatTerakhir  string  `json:"tmt_pangkat_terakhir"`
		TMTKGBTerakhir      string  `json:"tmt_kgb_terakhir"`
		JenjangPendidikan   *string `json:"jenjang_pendidikan"`
		Jurusan             *string `json:"jurusan"`
		InstitusiPendidikan *string `json:"institusi_pendidikan"`
		Password            string  `json:"password"` // Legacy alias; never defaulted.
		InitialPassword     string  `json:"initial_password"`
	}

	var req CreatePegawaiRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Payload tidak valid"})
	}

	if req.InitialPassword != "" {
		if req.Password != "" && req.Password != req.InitialPassword {
			return c.Status(400).JSON(fiber.Map{"success": false, "message": "Password awal tidak konsisten"})
		}
		req.Password = req.InitialPassword
	}
	if utf8.RuneCountInString(req.Password) < 8 || len(req.Password) > 72 {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Password awal wajib 8 karakter atau lebih, maksimum 72 byte"})
	}
	if !validation.ValidNIP(req.NIP) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "NIP harus berjumlah 18 digit angka"})
	}

	if err := validateMasterFields(c.Body()); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	tmtPengangkatan, err := validation.ParseRequiredDate(req.TMTPengangkatan, "TMT pengangkatan")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	tmtPangkat, err := validation.ParseRequiredDate(req.TMTPangkatTerakhir, "TMT pangkat terakhir")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	tmtKGB, err := validation.ParseRequiredDate(req.TMTKGBTerakhir, "TMT KGB terakhir")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	var tglLahir *time.Time
	if req.TanggalLahir != "" {
		t, parseErr := time.Parse(time.DateOnly, req.TanggalLahir)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Tanggal lahir harus berformat YYYY-MM-DD"})
		}
		tglLahir = &t
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memulai transaksi"})
	}
	defer tx.Rollback()

	// 1. Buat user login jika belum ada
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal mengamankan kata sandi"})
	}

	var userID int
	err = tx.QueryRow(
		"INSERT INTO sikapj.users (username, password_hash, role, is_active, must_change_password) VALUES ($1, $2, 'PEGAWAI', TRUE, TRUE) RETURNING id",
		req.NIP, hash,
	).Scan(&userID)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Akun tidak dapat dibuat; NIP mungkin sudah digunakan"})
	}

	// 2. Buat profil pegawai
	query := `
		INSERT INTO sikapj.pegawai (
			nip, user_id, nama_lengkap, tempat_lahir, tanggal_lahir, jenis_kelamin,
			email, no_hp, status_pegawai, bidang, jenis_jabatan, nama_jabatan,
			pangkat, golongan, tmt_pengangkatan, tmt_pangkat_terakhir, tmt_kgb_terakhir,
			jenjang_pendidikan, jurusan, institusi_pendidikan, is_active
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, TRUE
		)
	`
	_, err = tx.Exec(query,
		req.NIP, userID, req.NamaLengkap, req.TempatLahir, tglLahir, req.JenisKelamin,
		req.Email, req.NoHP, req.StatusPegawai, req.Bidang, req.JenisJabatan, req.NamaJabatan,
		req.Pangkat, req.Golongan, tmtPengangkatan, tmtPangkat, tmtKGB,
		req.JenjangPendidikan, req.Jurusan, req.InstitusiPendidikan,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menambah data pegawai",
		})
	}

	actor := c.Locals("user_id").(int)
	if err := (audit.Service{DB: tx}).Log(&actor, "CREATE_PEGAWAI", map[string]string{"nip": req.NIP}, c.IP()); err != nil {
		return internalError(c)
	}
	if err := tx.Commit(); err != nil {
		return internalError(c)
	}
	h.syncAlertsAfterCommit()
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Pegawai berhasil ditambahkan",
	})
}

// UpdatePegawai memperbarui data master pegawai (termasuk update TMT yang me-reset alert)
func (h *PegawaiHandler) UpdatePegawai(c *fiber.Ctx) error {
	nip := c.Params("nip")
	type UpdateRequest struct {
		NamaLengkap         string  `json:"nama_lengkap"`
		TempatLahir         *string `json:"tempat_lahir"`
		TanggalLahir        string  `json:"tanggal_lahir"`
		JenisKelamin        *string `json:"jenis_kelamin"`
		Email               *string `json:"email"`
		NoHP                *string `json:"no_hp"`
		StatusPegawai       string  `json:"status_pegawai"`
		Bidang              string  `json:"bidang"`
		JenisJabatan        string  `json:"jenis_jabatan"`
		NamaJabatan         string  `json:"nama_jabatan"`
		Pangkat             string  `json:"pangkat"`
		Golongan            string  `json:"golongan"`
		TMTPengangkatan     string  `json:"tmt_pengangkatan"`
		TMTPangkatTerakhir  string  `json:"tmt_pangkat_terakhir"`
		TMTKGBTerakhir      string  `json:"tmt_kgb_terakhir"`
		JenjangPendidikan   *string `json:"jenjang_pendidikan"`
		Jurusan             *string `json:"jurusan"`
		InstitusiPendidikan *string `json:"institusi_pendidikan"`
		IsActive            *bool   `json:"is_active"`
	}

	var req UpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Payload tidak valid"})
	}

	if err := validateMasterFields(c.Body()); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	tmtPengangkatan, err := validation.ParseRequiredDate(req.TMTPengangkatan, "TMT pengangkatan")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	tmtPangkat, err := validation.ParseRequiredDate(req.TMTPangkatTerakhir, "TMT pangkat terakhir")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	tmtKGB, err := validation.ParseRequiredDate(req.TMTKGBTerakhir, "TMT KGB terakhir")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	var tglLahir *time.Time
	if req.TanggalLahir != "" {
		t, parseErr := time.Parse(time.DateOnly, req.TanggalLahir)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Tanggal lahir harus berformat YYYY-MM-DD"})
		}
		tglLahir = &t
	}
	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memulai transaksi"})
	}
	defer tx.Rollback()

	query := `
		UPDATE sikapj.pegawai SET
			nama_lengkap = $1, tempat_lahir = $2, tanggal_lahir = $3, jenis_kelamin = $4,
			email = $5, no_hp = $6, status_pegawai = $7, bidang = $8,
			jenis_jabatan = $9, nama_jabatan = $10, pangkat = $11, golongan = $12,
			tmt_pengangkatan = $13, tmt_pangkat_terakhir = $14, tmt_kgb_terakhir = $15,
			jenjang_pendidikan = $16, jurusan = $17, institusi_pendidikan = $18,
			is_active = COALESCE($19, is_active), updated_at = NOW()
		WHERE nip = $20
		RETURNING is_active
	`
	var effectiveActive bool
	err = tx.QueryRow(query,
		req.NamaLengkap, req.TempatLahir, tglLahir, req.JenisKelamin,
		req.Email, req.NoHP, req.StatusPegawai, req.Bidang,
		req.JenisJabatan, req.NamaJabatan, req.Pangkat, req.Golongan,
		tmtPengangkatan, tmtPangkat, tmtKGB,
		req.JenjangPendidikan, req.Jurusan, req.InstitusiPendidikan,
		req.IsActive, nip,
	).Scan(&effectiveActive)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Pegawai tidak ditemukan atau gagal diperbarui"})
	}
	if _, err := tx.Exec("UPDATE sikapj.users SET is_active = $1, updated_at = NOW() WHERE username = $2", effectiveActive, nip); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal menyinkronkan status akun"})
	}
	actor := c.Locals("user_id").(int)
	if err := (audit.Service{DB: tx}).Log(&actor, "UPDATE_PEGAWAI", map[string]string{"nip": nip}, c.IP()); err != nil {
		return internalError(c)
	}
	if err := tx.Commit(); err != nil {
		return internalError(c)
	}
	h.syncAlertsAfterCommit()

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Data pegawai berhasil diperbarui",
	})
}

// ToggleStatusPegawai mengaktifkan / menonaktifkan pegawai
func (h *PegawaiHandler) ToggleStatusPegawai(c *fiber.Ctx) error {
	nip := c.Params("nip")
	tx, err := h.DB.Beginx()
	if err != nil {
		return internalError(c)
	}
	defer tx.Rollback()
	var isActive bool
	err = tx.QueryRow("UPDATE sikapj.pegawai SET is_active = NOT is_active, updated_at = NOW() WHERE nip = $1 RETURNING is_active", nip).Scan(&isActive)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Pegawai tidak ditemukan"})
	}
	result, err := tx.Exec("UPDATE sikapj.users SET is_active = $1, updated_at=NOW() WHERE username = $2", isActive, nip)
	if err != nil {
		return internalError(c)
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Akun pegawai tidak ditemukan"})
	}
	adminID := c.Locals("user_id").(int)
	_, err = tx.Exec(`INSERT INTO sikapj.audit_log(user_id,action,details,ip_address) VALUES($1,'TOGGLE_PEGAWAI',$2,$3)`, adminID, mapDetails(nip, isActive), c.IP())
	if err != nil {
		return internalError(c)
	}
	if err = tx.Commit(); err != nil {
		return internalError(c)
	}
	h.syncAlertsAfterCommit()

	return c.JSON(fiber.Map{
		"success":   true,
		"message":   "Status keaktifan pegawai berhasil diubah",
		"is_active": isActive,
	})
}

// UpdateMyContact memperbarui email dan nomor HP milik pegawai yang sedang login.
func (h *PegawaiHandler) UpdateMyContact(c *fiber.Ctx) error {
	nip := c.Locals("nip").(string)
	type request struct {
		Email *string `json:"email"`
		NoHP  *string `json:"no_hp"`
	}
	var req request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Payload tidak valid"})
	}
	if req.Email == nil && req.NoHP == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Email atau nomor HP wajib diisi"})
	}
	if err := validateContact(req.Email, req.NoHP); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	tx, err := h.DB.Beginx()
	if err != nil {
		return internalError(c)
	}
	defer tx.Rollback()
	result, err := tx.Exec(`UPDATE sikapj.pegawai SET email = COALESCE($1, email), no_hp = COALESCE($2, no_hp), updated_at = NOW() WHERE nip = $3`, req.Email, req.NoHP, nip)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memperbarui kontak"})
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Pegawai tidak ditemukan"})
	}
	actor := c.Locals("user_id").(int)
	if err = (audit.Service{DB: tx}).Log(&actor, "UPDATE_CONTACT", map[string]string{"nip": nip}, c.IP()); err != nil {
		return internalError(c)
	}
	if err = tx.Commit(); err != nil {
		return internalError(c)
	}
	return c.JSON(fiber.Map{"success": true, "message": "Kontak berhasil diperbarui"})
}

func (h *PegawaiHandler) syncAlertsAfterCommit() {
	threshold := 3
	if h.Cfg != nil {
		threshold = h.Cfg.AlertThresholdMonths
	}
	if _, err := services.SyncAlertLogs(h.DB, threshold); err != nil {
		log.Printf("[WARN] Master pegawai committed; alert synchronization failed: %v", err)
	}
}

func validateContact(email, phone *string) error {
	if email != nil && *email != "" {
		a, err := mail.ParseAddress(*email)
		if err != nil || a.Address != *email || utf8.RuneCountInString(*email) > 254 {
			return fmt.Errorf("Email tidak valid (maksimal 254 karakter)")
		}
	}
	if phone != nil {
		if utf8.RuneCountInString(*phone) > 20 {
			return fmt.Errorf("Nomor HP maksimal 20 karakter")
		}
		for _, r := range *phone {
			if !strings.ContainsRune("0123456789+ -()", r) {
				return fmt.Errorf("Nomor HP tidak valid")
			}
		}
	}
	return nil
}
func validateMasterFields(body []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err != nil {
		return fmt.Errorf("Payload tidak valid")
	}
	limits := map[string]int{"nama_lengkap": 150, "tempat_lahir": 100, "bidang": 100, "nama_jabatan": 150, "pangkat": 50, "golongan": 10, "jenjang_pendidikan": 20, "jurusan": 100, "institusi_pendidikan": 150}
	required := map[string]bool{"nama_lengkap": true, "bidang": true, "nama_jabatan": true, "pangkat": true, "golongan": true}
	for field, max := range limits {
		var value string
		if raw, ok := fields[field]; ok {
			if err := json.Unmarshal(raw, &value); err != nil {
				return fmt.Errorf("%s tidak valid", field)
			}
		}
		if (required[field] && strings.TrimSpace(value) == "") || utf8.RuneCountInString(value) > max {
			return fmt.Errorf("%s wajib valid dan maksimal %d karakter", field, max)
		}
	}
	for field, allowed := range map[string][]string{"status_pegawai": {"PNS", "PPPK"}, "jenis_jabatan": {"Struktural", "Fungsional", "Pelaksana"}, "jenis_kelamin": {"", "L", "P"}} {
		var value string
		if raw, ok := fields[field]; ok {
			if err := json.Unmarshal(raw, &value); err != nil {
				return fmt.Errorf("%s tidak valid", field)
			}
		}
		found := false
		for _, v := range allowed {
			if value == v {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("%s tidak valid", field)
		}
	}
	var contact struct {
		Email *string `json:"email"`
		Phone *string `json:"no_hp"`
	}
	if err := json.Unmarshal(body, &contact); err != nil {
		return fmt.Errorf("Kontak tidak valid")
	}
	return validateContact(contact.Email, contact.Phone)
}
func mapDetails(nip string, active bool) string {
	return fmt.Sprintf(`{"nip":%q,"is_active":%t}`, nip, active)
}
