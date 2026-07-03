package config

import (
	"bufio"
	"log"
	"os"
	"strings"
)

func init() {
	loadEnvFile(".env")
	loadEnvFile("../.env")
}

func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("error reading .env: %v", err)
	}
}

type Config struct {
	Port               string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPass             string
	DBName             string
	Frontend           string
	SecretKey          string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	SMTPHost           string
	SMTPPort           string
	SMTPUser           string
	SMTPPass           string
	SMTPFrom           string
}

func Load() *Config {
	return &Config{
		Port:               getEnv("PORT", "8080"),
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBUser:             getEnv("DB_USER", "postgres"),
		DBPass:             getEnv("DB_PASS", "0619"),
		DBName:             getEnv("DB_NAME", "latrode"),
		Frontend:           getEnv("FRONTEND_PATH", "../frontend"),
		SecretKey:          getEnv("SECRET_KEY", "latrode-secret-key-change-in-production"),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/v1/auth/google/callback"),
		SMTPHost:           getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:           getEnv("SMTP_PORT", "587"),
		SMTPUser:           getEnv("SMTP_USER", ""),
		SMTPPass:           getEnv("SMTP_PASS", ""),
		SMTPFrom:           getEnv("SMTP_FROM", ""),
	}
}

func (c *Config) DatabaseURL() string {
	return "postgres://" + c.DBUser + ":" + c.DBPass + "@" + c.DBHost + ":" + c.DBPort + "/" + c.DBName + "?sslmode=disable"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
