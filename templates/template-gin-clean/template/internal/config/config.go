package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration.
type Config struct {
	AppName  string
	AppAddr  string
	DBDriver string
	DBDSN    string
	JWTSecret string
}

// Load reads environment variables and returns a Config.
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system env")
	}
	return &Config{
		AppName:   getEnv("APP_NAME", "{{PROJECT_NAME}}"),
		AppAddr:   ":" + getEnv("APP_PORT", "8080"),
		DBDriver:  getEnv("DB_DRIVER", "postgres"),
		DBDSN:     getEnv("DB_DSN", ""),
		JWTSecret: getEnv("JWT_SECRET", "change-me"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
