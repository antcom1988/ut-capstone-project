package main

import (
	"log"
	"os"
	"sikapj-backend/internal/config"
	"sikapj-backend/internal/database"
	"sikapj-backend/internal/handlers"
	"sikapj-backend/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// 1. Muat Konfigurasi
	cfg := config.LoadConfig()

	// 2. Buat folder upload lokal
	_ = os.MkdirAll(cfg.UploadDir, os.ModePerm)

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
		AllowOrigins:     "*",
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

	// 7. Routing API
	api := app.Group("/api")

	// --- Endpoint Publik ---
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "app": "SIKAP-J API", "schema": cfg.DBSchema})
	})
	api.Post("/auth/login", authHandler.Login)

	// --- Endpoint Terproteksi (Wajib Login) ---
	authGroup := api.Group("", middleware.AuthRequired(cfg))
	authGroup.Get("/auth/profile", authHandler.GetProfile)

	// Master Dokumen (Dapat diakses Pegawai & Admin)
	authGroup.Get("/dokumen/master", berkasHandler.GetMasterDokumen)

	// E-Berkas & In-Browser Preview (Inline Stream tanpa download)
	authGroup.Get("/berkas/:nip", berkasHandler.GetBerkasByNIP)
	authGroup.Post("/berkas/upload", berkasHandler.UploadBerkas)
	authGroup.Get("/berkas/preview/:id", berkasHandler.StreamPreviewFile) // << Stream file inline

	// Kinerja & Diklat
	authGroup.Get("/kinerja/:nip", kinerjaHandler.GetArsipKinerjaByNIP)
	authGroup.Post("/kinerja/upload", kinerjaHandler.UploadArsipKinerja)
	authGroup.Get("/kinerja/preview/:id", kinerjaHandler.StreamPreviewKinerja)

	authGroup.Get("/diklat/:nip", kinerjaHandler.GetRiwayatDiklatByNIP)
	authGroup.Post("/diklat", kinerjaHandler.CreateRiwayatDiklat)
	authGroup.Get("/diklat/preview/:id", kinerjaHandler.StreamPreviewSertifikat)

	// Portal Mandiri Pegawai (ESS)
	authGroup.Get("/pegawai/me/dashboard", dashboardHandler.GetPegawaiSelfDashboard)

	// --- Endpoint Khusus ADMIN (Subag Umpeg) ---
	adminGroup := authGroup.Group("/admin", middleware.RoleAdminOnly())

	// Dashboard & Alerts
	adminGroup.Get("/dashboard", dashboardHandler.GetDashboardSummary)
	adminGroup.Get("/alerts", dashboardHandler.GetAlertList)
	adminGroup.Put("/settings/alert-threshold", dashboardHandler.UpdateAlertThreshold)

	// Verifikasi Berkas
	adminGroup.Put("/berkas/verify/:id", berkasHandler.VerifikasiBerkas)

	// Master Data Pegawai
	adminGroup.Get("/pegawai", pegawaiHandler.GetAllPegawai)
	adminGroup.Get("/pegawai/:nip", pegawaiHandler.GetPegawaiDetail)
	adminGroup.Post("/pegawai", pegawaiHandler.CreatePegawai)
	adminGroup.Put("/pegawai/:nip", pegawaiHandler.UpdatePegawai)
	adminGroup.Put("/pegawai/:nip/toggle-status", pegawaiHandler.ToggleStatusPegawai)

	// 8. Menjalankan Server
	log.Printf("[READY] Server SIKAP-J Backend berjalan di port :%s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("[FATAL] Gagal menjalankan server Fiber: %v", err)
	}
}
