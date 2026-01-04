package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	AWS      AWSConfig
	CORS     CORSConfig
}

type ServerConfig struct {
	Port    string
	GinMode string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	Secret                string
	Expiration            time.Duration
	RefreshTokenExpiration time.Duration
}

type AWSConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	S3Bucket        string
}

type CORSConfig struct {
	AllowedOrigins []string
}

func Load() (*Config, error) {
	// Load .env file if exists (for local development)
	godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port:    getEnv("PORT", "8080"),
			GinMode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "linda_salon"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:                getEnv("JWT_SECRET", "change-this-secret-key"),
			Expiration:            parseDuration(getEnv("JWT_EXPIRATION", "24h")),
			RefreshTokenExpiration: parseDuration(getEnv("REFRESH_TOKEN_EXPIRATION", "168h")),
		},
		AWS: AWSConfig{
			Region:          getEnv("AWS_REGION", "ap-northeast-1"),
			AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
			S3Bucket:        getEnv("S3_BUCKET", "linda-salon-uploads"),
		},
	}

	// Parse allowed origins
	originsStr := getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:3001")
	cfg.CORS.AllowedOrigins = parseCSV(originsStr)

	return cfg, nil
}

func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}

func parseCSV(s string) []string {
	var result []string
	for i := 0; i < len(s); {
		j := i
		for j < len(s) && s[j] != ',' {
			j++
		}
		if i < j {
			result = append(result, s[i:j])
		}
		i = j + 1
	}
	return result
}
