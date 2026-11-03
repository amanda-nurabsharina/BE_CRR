package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPass      string
	DBName      string
	DBSSLMode   string
	JWTSecret   string
	StorageDir  string
	Environment string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		Port:        getEnv("PORT", "8080"),
		DBHost:      getEnv("DBHost", getEnv("DB_HOST", "localhost")),
		DBPort:      getEnv("DBPort", getEnv("DB_PORT", "5432")),
		DBUser:      getEnv("DBUser", getEnv("DB_USER", "postgre")),
		DBPass:      getEnv("DBPass", getEnv("DB_PASS", "admin12345")),
		DBName:      getEnv("DBName", getEnv("DB_NAME", "court_recording")),
		DBSSLMode:   getEnv("DBSSLMode", getEnv("DB_SSLMODE", "disable")),
		JWTSecret:   getEnv("JWT_SECRET", "court_recording_super_secure_jwt_secret_key_2026_indonesia"),
		StorageDir:  getEnv("STORAGE_DIR", "./storage_data"),
		Environment: getEnv("APP_ENV", "development"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

