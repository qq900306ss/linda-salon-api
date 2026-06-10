// Package app wires configuration, database, repositories and the router.
// It is shared by the Lambda and the local server entrypoints.
package app

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/qq900306ss/linda-salon-api/internal/auth"
	"github.com/qq900306ss/linda-salon-api/internal/database"
	"github.com/qq900306ss/linda-salon-api/internal/handler"
	"github.com/qq900306ss/linda-salon-api/internal/repository"
	"github.com/qq900306ss/linda-salon-api/internal/service"
)

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Initialize sets up the database (tables + seed data) and returns the
// fully wired Gin router. Called once per cold start.
func Initialize(ctx context.Context) (*gin.Engine, error) {
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Database.
	db, err := database.New(ctx, envOrDefault("TABLE_PREFIX", "linda"))
	if err != nil {
		return nil, err
	}
	if err := db.EnsureTables(ctx); err != nil {
		return nil, err
	}

	adminUsername := envOrDefault("ADMIN_USERNAME", "admin")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	isDefaultPassword := adminPassword == ""
	if isDefaultPassword {
		adminPassword = "linda2024"
	}
	if err := db.Seed(ctx, adminUsername, adminPassword, isDefaultPassword); err != nil {
		return nil, err
	}

	// JWT.
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "linda-salon-dev-secret-do-not-use-in-production"
		log.Println("WARNING: JWT_SECRET not set, using the insecure development secret!")
	}
	jwtManager := auth.NewJWTManager(jwtSecret)

	// S3 (optional).
	s3Service, err := service.NewS3Service(ctx, os.Getenv("S3_BUCKET"))
	if err != nil {
		return nil, err
	}

	// Repositories.
	userRepo := repository.NewUserRepository(db)
	serviceRepo := repository.NewServiceRepository(db)
	stylistRepo := repository.NewStylistRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)

	// Handlers.
	handlers := handler.Handlers{
		Auth:       handler.NewAuthHandler(userRepo, jwtManager),
		Service:    handler.NewServiceHandler(serviceRepo),
		Stylist:    handler.NewStylistHandler(stylistRepo),
		Booking:    handler.NewBookingHandler(bookingRepo, serviceRepo, stylistRepo, settingsRepo),
		Settings:   handler.NewSettingsHandler(settingsRepo),
		Statistics: handler.NewStatisticsHandler(bookingRepo),
		Upload:     handler.NewUploadHandler(s3Service),
	}

	return handler.NewRouter(handlers, jwtManager, envOrDefault("ALLOWED_ORIGINS", "*")), nil
}
