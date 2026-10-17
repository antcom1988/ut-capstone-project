package handlers

import (
	"database/sql"
	"sikapj-backend/internal/audit"
	"sikapj-backend/internal/auth"
	"sikapj-backend/internal/config"
	"sikapj-backend/internal/models"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type AuthHandler struct {
	DB            *sqlx.DB
	Cfg           *config.Config
	loginMu       sync.Mutex
	loginAttempts map[string]loginWindow
}

type loginWindow struct {
	Count   int
	Expires time.Time
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// reserveLogin bounds both concurrent attempts and memory; trusted peer IP only.
func (h *AuthHandler) reserveLogin(ip string) bool {
	h.loginMu.Lock()
	defer h.loginMu.Unlock()
	now := time.Now()
	if h.loginAttempts == nil {
		h.loginAttempts = make(map[string]loginWindow)
	}
	for key, w := range h.loginAttempts {
		if !now.Before(w.Expires) {
			delete(h.loginAttempts, key)
		}
	}
	w, exists := h.loginAttempts[ip]
	if !exists {
		if len(h.loginAttempts) >= 10000 {
			return false
		}
		w.Expires = now.Add(15 * time.Minute)
	}
	if w.Count >= 10 {
		return false
	}
	w.Count++
	h.loginAttempts[ip] = w
	return true
}
func (h *AuthHandler) releaseLogin(ip string) {
	h.loginMu.Lock()
	defer h.loginMu.Unlock()
	w, ok := h.loginAttempts[ip]
	if ok && w.Count > 0 {
		w.Count--
		if w.Count == 0 {
			delete(h.loginAttempts, ip)
		} else {
			h.loginAttempts[ip] = w
		}
	}
}
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	ip := c.IP()
	if !h.reserveLogin(ip) {
		c.Set("Retry-After", "900")
		return c.Status(429).JSON(fiber.Map{"success": false, "message": "Terlalu banyak percobaan login. Coba lagi dalam 15 menit."})
	}
	defer func() {
		status := c.Response().StatusCode()
		if status != 401 && status != 403 {
			h.releaseLogin(ip)
		}
	}()
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
	query := `SELECT id, username, password_hash, role, is_active, must_change_password, session_version, last_login, created_at, updated_at
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
	token, err := auth.GenerateJWTToken(user.ID, user.Username, user.Role, nip, h.Cfg, user.SessionVersion)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengenerate token autentikasi",
		})
	}

	// Login metadata and audit must succeed together before issuing a session.
	tx, err := h.DB.Beginx()
	if err != nil {
		return internalError(c)
	}
	defer tx.Rollback()
	if _, err = tx.Exec("UPDATE sikapj.users SET last_login = $1 WHERE id = $2", time.Now(), user.ID); err != nil {
		return internalError(c)
	}
	if err = (audit.Service{DB: tx}).Log(&user.ID, "LOGIN", map[string]string{"result": "success"}, c.IP()); err != nil {
		return internalError(c)
	}
	if err = tx.Commit(); err != nil {
		return internalError(c)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Login berhasil",
		"data": fiber.Map{
			"token": token,
			"user": fiber.Map{
				"id":                   user.ID,
				"username":             user.Username,
				"role":                 user.Role,
				"nip":                  nip,
				"must_change_password": user.MustChangePassword,
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

func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	type request struct {
		Current string `json:"current_password"`
		New     string `json:"new_password"`
	}
	var req request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Payload tidak valid"})
	}
	if len(req.New) < 10 {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Kata sandi baru minimal 10 karakter"})
	}
	userID := c.Locals("user_id").(int)
	var hash string
	if err := h.DB.Get(&hash, `SELECT password_hash FROM sikapj.users WHERE id=$1 AND is_active=TRUE`, userID); err != nil || !auth.CheckPasswordHash(req.Current, hash) {
		return c.Status(401).JSON(fiber.Map{"success": false, "message": "Kata sandi saat ini tidak sesuai"})
	}
	newHash, err := auth.HashPassword(req.New)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Terjadi kesalahan internal"})
	}
	tx, err := h.DB.Beginx()
	if err != nil {
		return internalError(c)
	}
	defer tx.Rollback()
	result, err := tx.Exec(`UPDATE sikapj.users SET password_hash=$1,must_change_password=FALSE,session_version=session_version+1,updated_at=NOW() WHERE id=$2 AND password_hash=$3`, newHash, userID, hash)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Terjadi kesalahan internal"})
	}
	n, _ := result.RowsAffected()
	if n != 1 {
		return c.Status(404).JSON(fiber.Map{"success": false, "message": "Pengguna tidak ditemukan"})
	}
	if err = (audit.Service{DB: tx}).Log(&userID, "CHANGE_PASSWORD", map[string]string{"result": "success"}, c.IP()); err != nil {
		return internalError(c)
	}
	if err = tx.Commit(); err != nil {
		return internalError(c)
	}
	return c.JSON(fiber.Map{"success": true, "message": "Kata sandi berhasil diubah"})
}
