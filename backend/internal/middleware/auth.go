package middleware

import (
	"sikapj-backend/internal/auth"
	"sikapj-backend/internal/config"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AuthRequired memvalidasi bearer token JWT
func AuthRequired(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		tokenStr := ""

		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenStr = parts[1]
			}
		} else {
			// Cek query parameter (untuk preview di iframe/img)
			tokenStr = c.Query("token")
		}

		if tokenStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Akses ditolak: Token autentikasi tidak ditemukan",
			})
		}

		claims, err := auth.ValidateJWTToken(tokenStr, cfg)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Sesi Anda tidak valid atau telah kedaluwarsa",
			})
		}

		// Simpan claims ke dalam konteks request
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("role", claims.Role)
		c.Locals("nip", claims.NIP)

		return c.Next()
	}
}

// RoleAdminOnly memastikan hanya pengguna dengan role ADMIN yang bisa mengakses
func RoleAdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || role != "ADMIN" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Akses terlarang: Fitur ini hanya dapat diakses oleh Admin Pengelola Kepegawaian",
			})
		}
		return c.Next()
	}
}
