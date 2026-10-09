package handlers

import (
	"database/sql"
	"sikapj-backend/internal/auth"
	"sikapj-backend/internal/config"
	"sikapj-backend/internal/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type AuthHandler struct {
	DB  *sqlx.DB
	Cfg *config.Config
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format payload request tidak valid",
		})
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Username/NIP dan kata sandi wajib diisi",
		})
	}

	var user models.User
	query := `SELECT id, username, password_hash, role, is_active, last_login, created_at, updated_at 
	          FROM sikapj.users WHERE username = $1 LIMIT 1`
	err := h.DB.Get(&user, query, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "NIP/Username atau kata sandi tidak sesuai",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Terjadi kesalahan pada database",
		})
	}

	if !user.IsActive {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Akun Anda berstatus nonaktif. Silakan hubungi Subag Umpeg.",
		})
	}

	// Cek kesesuaian password
	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "NIP/Username atau kata sandi tidak sesuai",
		})
	}

	// Dapatkan NIP jika role PEGAWAI
	nip := ""
	if user.Role == "PEGAWAI" {
		_ = h.DB.Get(&nip, "SELECT nip FROM sikapj.pegawai WHERE user_id = $1 LIMIT 1", user.ID)
	}

	// Buat token JWT
	token, err := auth.GenerateJWTToken(user.ID, user.Username, user.Role, nip, h.Cfg)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengenerate token autentikasi",
		})
	}

	// Update last_login
	now := time.Now()
	_, _ = h.DB.Exec("UPDATE sikapj.users SET last_login = $1 WHERE id = $2", now, user.ID)

	// Catat audit log
	_, _ = h.DB.Exec(
		"INSERT INTO sikapj.audit_log (user_id, action, details, ip_address) VALUES ($1, $2, $3, $4)",
		user.ID, "LOGIN", "Pengguna berhasil login ke sistem", c.IP(),
	)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Login berhasil",
		"data": fiber.Map{
			"token": token,
			"user": fiber.Map{
				"id":       user.ID,
				"username": user.Username,
				"role":     user.Role,
				"nip":      nip,
			},
		},
	})
}

func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	role := c.Locals("role").(string)

	if role == "PEGAWAI" {
		var p models.Pegawai
		err := h.DB.Get(&p, "SELECT * FROM sikapj.pegawai WHERE user_id = $1 LIMIT 1", userID)
		if err == nil {
			return c.JSON(fiber.Map{
				"success": true,
				"data": fiber.Map{
					"role":    role,
					"pegawai": p,
				},
			})
		}
	}

	var u models.User
	_ = h.DB.Get(&u, "SELECT id, username, role, last_login FROM sikapj.users WHERE id = $1", userID)
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"role": role,
			"user": u,
		},
	})
}
