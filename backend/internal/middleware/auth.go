package middleware

import (
	"sikapj-backend/internal/auth"
	"sikapj-backend/internal/config"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/gofiber/fiber/v2"
)

// AuthRequired memvalidasi bearer token JWT
func AuthRequired(cfg *config.Config, db *sqlx.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		tokenStr := ""
		if authHeader != "" { parts := strings.Fields(authHeader); if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") { tokenStr = parts[1] } }
		if tokenStr == "" { return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Akses ditolak: Token autentikasi tidak ditemukan"}) }
		claims, err := auth.ValidateJWTToken(tokenStr, cfg)
		if err != nil { return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Sesi Anda tidak valid atau telah kedaluwarsa"}) }
		var state struct { Active bool `db:"is_active"`; MustChange bool `db:"must_change_password"`; Version int `db:"session_version"` }
		if err := db.Get(&state, "SELECT is_active, must_change_password, session_version FROM sikapj.users WHERE id = $1", claims.UserID); err != nil || !state.Active || state.Version != claims.SessionVersion { return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Akun tidak aktif atau tidak ditemukan"}) }
		allowed := (c.Method() == "GET" && c.Path() == "/api/auth/profile") || (c.Method() == "PUT" && c.Path() == "/api/auth/password")
		if state.MustChange && !allowed { return c.Status(403).JSON(fiber.Map{"success": false, "code": "PASSWORD_CHANGE_REQUIRED", "message": "Kata sandi wajib diubah terlebih dahulu"}) }
		c.Locals("must_change_password", state.MustChange); c.Locals("user_id", claims.UserID); c.Locals("username", claims.Username); c.Locals("role", claims.Role); c.Locals("nip", claims.NIP)
		return c.Next()
	}
}

// RolePegawaiOnly memastikan hanya pegawai yang bisa mengunggah atau mengganti dokumen pribadi
func RolePegawaiOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || role != "PEGAWAI" { return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": "Admin hanya dapat memverifikasi dokumen, bukan mengunggah atau mengganti dokumen pegawai"}) }
		return c.Next()
	}
}

// RoleAdminOnly memastikan hanya pengguna dengan role ADMIN yang bisa mengakses
func RoleAdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || role != "ADMIN" { return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": "Akses terlarang: Fitur ini hanya dapat diakses oleh Admin Pengelola Kepegawaian"}) }
		return c.Next()
	}
}
