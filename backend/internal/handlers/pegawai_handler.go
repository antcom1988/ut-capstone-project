package handlers

import (
	"sikapj-backend/internal/auth"
	"sikapj-backend/internal/config"
	"sikapj-backend/internal/models"
	"time"

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
			"message": "Gagal mengambil data pegawai: " + err.Error(),
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
		Password            string  `json:"password"` // Default password jika diisi
	}

	var req CreatePegawaiRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Payload tidak valid"})
	}

	if len(req.NIP) != 18 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "NIP harus berjumlah 18 digit angka"})
	}

	tmtPengangkatan, _ := time.Parse("2006-01-02", req.TMTPengangkatan)
	tmtPangkat, _ := time.Parse("2006-01-02", req.TMTPangkatTerakhir)
	tmtKGB, _ := time.Parse("2006-01-02", req.TMTKGBTerakhir)
	var tglLahir *time.Time
	if req.TanggalLahir != "" {
		if t, err := time.Parse("2006-01-02", req.TanggalLahir); err == nil {
			tglLahir = &t
		}
	}

	// 1. Buat user login jika belum ada
	defaultPassword := "Pegawai123!"
	if req.Password != "" {
		defaultPassword = req.Password
	}
	hash, _ := auth.HashPassword(defaultPassword)

	var userID int
	err := h.DB.QueryRow(
		"INSERT INTO sikapj.users (username, password_hash, role, is_active) VALUES ($1, $2, 'PEGAWAI', TRUE) RETURNING id",
		req.NIP, hash,
	).Scan(&userID)
	if err != nil {
		// Jika username sudah ada, ambil id nya
		_ = h.DB.Get(&userID, "SELECT id FROM sikapj.users WHERE username = $1", req.NIP)
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
	_, err = h.DB.Exec(query,
		req.NIP, userID, req.NamaLengkap, req.TempatLahir, tglLahir, req.JenisKelamin,
		req.Email, req.NoHP, req.StatusPegawai, req.Bidang, req.JenisJabatan, req.NamaJabatan,
		req.Pangkat, req.Golongan, tmtPengangkatan, tmtPangkat, tmtKGB,
		req.JenjangPendidikan, req.Jurusan, req.InstitusiPendidikan,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menambah data pegawai: " + err.Error(),
		})
	}

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

	tmtPengangkatan, _ := time.Parse("2006-01-02", req.TMTPengangkatan)
	tmtPangkat, _ := time.Parse("2006-01-02", req.TMTPangkatTerakhir)
	tmtKGB, _ := time.Parse("2006-01-02", req.TMTKGBTerakhir)
	var tglLahir *time.Time
	if req.TanggalLahir != "" {
		if t, err := time.Parse("2006-01-02", req.TanggalLahir); err == nil {
			tglLahir = &t
		}
	}

	query := `
		UPDATE sikapj.pegawai SET
			nama_lengkap = $1, tempat_lahir = $2, tanggal_lahir = $3, jenis_kelamin = $4,
			email = $5, no_hp = $6, status_pegawai = $7, bidang = $8,
			jenis_jabatan = $9, nama_jabatan = $10, pangkat = $11, golongan = $12,
			tmt_pengangkatan = $13, tmt_pangkat_terakhir = $14, tmt_kgb_terakhir = $15,
			jenjang_pendidikan = $16, jurusan = $17, institusi_pendidikan = $18,
			updated_at = NOW()
		WHERE nip = $19
	`
	_, err := h.DB.Exec(query,
		req.NamaLengkap, req.TempatLahir, tglLahir, req.JenisKelamin,
		req.Email, req.NoHP, req.StatusPegawai, req.Bidang,
		req.JenisJabatan, req.NamaJabatan, req.Pangkat, req.Golongan,
		tmtPengangkatan, tmtPangkat, tmtKGB,
		req.JenjangPendidikan, req.Jurusan, req.InstitusiPendidikan,
		nip,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal memperbarui data pegawai: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Data pegawai berhasil diperbarui",
	})
}

// ToggleStatusPegawai mengaktifkan / menonaktifkan pegawai
func (h *PegawaiHandler) ToggleStatusPegawai(c *fiber.Ctx) error {
	nip := c.Params("nip")
	var isActive bool
	err := h.DB.QueryRow("UPDATE sikapj.pegawai SET is_active = NOT is_active, updated_at = NOW() WHERE nip = $1 RETURNING is_active", nip).Scan(&isActive)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal mengubah status"})
	}

	// Sinkronkan ke akun user
	_, _ = h.DB.Exec("UPDATE sikapj.users SET is_active = $1 WHERE username = $2", isActive, nip)

	return c.JSON(fiber.Map{
		"success":   true,
		"message":   "Status keaktifan pegawai berhasil diubah",
		"is_active": isActive,
	})
}
