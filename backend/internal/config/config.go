package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                 string
	AppEnv               string
	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	DBSchema             string
	DBSSLMode            string
	JWTSecret            string
	JWTExpirationHours   int
	UploadDir            string
	MaxFileSizeMB        int64
	AlertThresholdMonths int
	CORSAllowOrigins     string
}

var AppConfig *Config

func LoadConfig() *Config {
	// Load .env jika ada (ignore error jika sudah di set di environment OS)
	if err := godotenv.Load(); err != nil {
		log.Println("[INFO] File .env tidak ditemukan, menggunakan environment system")
	}

	jwtExp, _ := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))
	maxSize, _ := strconv.ParseInt(getEnv("MAX_FILE_SIZE_MB", "2"), 10, 64)
	alertMonths, _ := strconv.Atoi(getEnv("ALERT_THRESHOLD_MONTHS", "3"))

	appEnv := getEnv("APP_ENV", "development")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	jwtSecret := getEnv("JWT_SECRET", "super_secret_sikapj_jwt_key_2026_bpkad_jember")
	if appEnv == "production" {
		if dbPassword == "postgres" || jwtSecret == "super_secret_sikapj_jwt_key_2026_bpkad_jember" {
			log.Fatal("[FATAL] DB_PASSWORD dan JWT_SECRET wajib dikonfigurasi dengan nilai aman pada production")
		}
	}

	AppConfig = &Config{
		Port:                 getEnv("PORT", "8080"),
		AppEnv:               appEnv,
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5432"),
		DBUser:               getEnv("DB_USER", "postgres"),
		DBPassword:           dbPassword,
		DBName:               getEnv("DB_NAME", "sikap_j"),
		DBSchema:             getEnv("DB_SCHEMA", "sikapj"),
		DBSSLMode:            getEnv("DB_SSLMODE", "disable"),
		JWTSecret:            jwtSecret,
		JWTExpirationHours:   jwtExp,
		UploadDir:            getEnv("UPLOAD_DIR", "storage/uploads"),
		MaxFileSizeMB:        maxSize,
		AlertThresholdMonths: alertMonths,
		CORSAllowOrigins:     getEnv("CORS_ALLOW_ORIGINS", "http://localhost:5173"),
	}

	return AppConfig
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
