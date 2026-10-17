package main

import (
	"log"
	"os"
	"sikapj-backend/internal/config"
	"sikapj-backend/internal/database"
	"sikapj-backend/internal/handlers"
	"sikapj-backend/internal/middleware"
	"sikapj-backend/internal/services"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// 1. Muat Konfigurasi
	cfg := config.LoadConfig()

	// 2. Buat folder upload lokal
	_ = os.MkdirAll(cfg.UploadDir, 0750)

	// 3. Hubungkan ke PostgreSQL (Schema: sikapj)
	db := database.ConnectDB(cfg)
	defer db.Close()

	// 4. Inisialisasi Fiber App
	app := fiber.New(fiber.Config{
		AppName:   "SIKAP-J Backend API v1.0",
		BodyLimit: 20 * 1024 * 1024, // 20 MB max file stream
	})

	// 5. Global Middlewares
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: false,
	}))

	// 6. Inisialisasi Handlers
	authHandler := &handlers.AuthHandler{DB: db, Cfg: cfg}
	pegawaiHandler := &handlers.PegawaiHandler{DB: db, Cfg: cfg}
	berkasHandler := &handlers.BerkasHandler{DB: db, Cfg: cfg}
	kinerjaHandler := &handlers.KinerjaDiklatHandler{DB: db, Cfg: cfg}
	dashboardHandler := &handlers.DashboardAlertHandler{DB: db, Cfg: cfg}
	adminHandler := &handlers.AdminHandler{DB: db}

	// 7. Routing API
	api := app.Group("/api")

	// --- Endpoint Publik ---
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "app": "SIKAP-J API", "schema": cfg.DBSchema})
	})
	api.Post("/auth/login", authHandler.Login)

	// --- Endpoint Terproteksi (Wajib Login) ---
	authGroup := api.Group("", middleware.AuthRequired(cfg, db))
	authGroup.Get("/auth/profile", authHandler.GetProfile)
	authGroup.Put("/auth/password", authHandler.ChangePassword)

	// Master Dokumen (Dapat diakses Pegawai & Admin)
	authGroup.Get("/dokumen/master", berkasHandler.GetMasterDokumen)

	// E-Berkas & In-Browser Preview (Inline Stream tanpa download)
	authGroup.Get("/berkas/:nip", berkasHandler.GetBerkasByNIP)
	authGroup.Post("/berkas/upload", middleware.RolePegawaiOnly(), berkasHandler.UploadBerkas)
	authGroup.Get("/berkas/preview/:id", berkasHandler.StreamPreviewFile) // << Stream file inline

	// Kinerja & Diklat
	authGroup.Get("/kinerja/:nip", kinerjaHandler.GetArsipKinerjaByNIP)
	authGroup.Post("/kinerja/upload", middleware.RolePegawaiOnly(), kinerjaHandler.UploadArsipKinerja)
	authGroup.Get("/kinerja/preview/:id", kinerjaHandler.StreamPreviewKinerja)

	authGroup.Get("/diklat/:nip", kinerjaHandler.GetRiwayatDiklatByNIP)
	authGroup.Post("/diklat", kinerjaHandler.CreateRiwayatDiklat)
	authGroup.Get("/diklat/preview/:id", kinerjaHandler.StreamPreviewSertifikat)

	// Portal Mandiri Pegawai (ESS)
	authGroup.Get("/pegawai/me/dashboard", dashboardHandler.GetPegawaiSelfDashboard)
	authGroup.Put("/pegawai/me/contact", pegawaiHandler.UpdateMyContact)

	// --- Endpoint Khusus ADMIN (Subag Umpeg) ---
	adminGroup := authGroup.Group("/admin", middleware.RoleAdminOnly())

	// Dashboard & Alerts
	adminGroup.Get("/dashboard", dashboardHandler.GetDashboardSummary)
	adminGroup.Get("/alerts", dashboardHandler.GetAlertList)
	adminGroup.Put("/settings/alert-threshold", dashboardHandler.UpdateAlertThreshold)

	// Verifikasi Berkas
	adminGroup.Put("/berkas/verify/:id", berkasHandler.VerifikasiBerkas)
	adminGroup.Get("/berkas/pending", adminHandler.GetPendingVerification)
	adminGroup.Get("/audit-logs", adminHandler.GetAuditLogs)
	adminGroup.Get("/reports/employees/export", adminHandler.ExportEmployees)
	adminGroup.Post("/alerts/sync", func(c *fiber.Ctx) error {
		count, err := services.SyncAlertLogs(db, cfg.AlertThresholdMonths)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"success": false, "message": "Terjadi kesalahan internal"})
		}
		return c.JSON(fiber.Map{"success": true, "synced": count})
	})

	// Master Data Pegawai
	adminGroup.Get("/pegawai", pegawaiHandler.GetAllPegawai)
	adminGroup.Get("/pegawai/:nip", pegawaiHandler.GetPegawaiDetail)
	adminGroup.Post("/pegawai", pegawaiHandler.CreatePegawai)
	adminGroup.Put("/pegawai/:nip", pegawaiHandler.UpdatePegawai)
	adminGroup.Put("/pegawai/:nip/toggle-status", pegawaiHandler.ToggleStatusPegawai)

	// Sinkronisasi alert saat startup dan setiap 24 jam.
	go func() {
		if _, err := services.SyncAlertLogs(db, cfg.AlertThresholdMonths); err != nil {
			log.Printf("[ERROR] Startup alert synchronization failed: %v", err)
		}
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if _, err := services.SyncAlertLogs(db, cfg.AlertThresholdMonths); err != nil {
				log.Printf("[ERROR] Scheduled alert synchronization failed: %v", err)
			}
		}
	}()

	// 8. Menjalankan Server
	log.Printf("[READY] Server SIKAP-J Backend berjalan di port :%s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("[FATAL] Gagal menjalankan server Fiber: %v", err)
	}
}
