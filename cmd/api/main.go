package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"

	"linda-salon-api/config"
	"linda-salon-api/internal/auth"
	"linda-salon-api/internal/database"
	"linda-salon-api/internal/handler"
	"linda-salon-api/internal/middleware"
	"linda-salon-api/internal/repository"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Initialize database
	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("❌ Failed to run migrations: %v", err)
	}

	// Initialize AWS S3 client
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.AWS.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AWS.AccessKeyID,
			cfg.AWS.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		log.Fatalf("❌ Failed to load AWS config: %v", err)
	}
	s3Client := s3.NewFromConfig(awsCfg)

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(&cfg.JWT)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.DB)
	serviceRepo := repository.NewServiceRepository(db.DB)
	stylistRepo := repository.NewStylistRepository(db.DB)
	bookingRepo := repository.NewBookingRepository(db.DB)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userRepo, jwtManager)
	serviceHandler := handler.NewServiceHandler(serviceRepo)
	stylistHandler := handler.NewStylistHandler(stylistRepo)
	bookingHandler := handler.NewBookingHandler(bookingRepo, serviceRepo, stylistRepo, userRepo)
	statsHandler := handler.NewStatisticsHandler(bookingRepo, stylistRepo)
	uploadHandler := handler.NewUploadHandler(s3Client, &cfg.AWS)

	// Setup router
	router := setupRouter(cfg, jwtManager, authHandler, serviceHandler, stylistHandler, bookingHandler, statsHandler, uploadHandler)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("🚀 Server starting on %s", addr)
	log.Printf("📝 Environment: %s", cfg.Server.GinMode)
	log.Printf("🗄️  Database: %s@%s:%s/%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)

	// Graceful shutdown
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		if err := router.Run(addr); err != nil {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited")
}

func setupRouter(
	cfg *config.Config,
	jwtManager *auth.JWTManager,
	authHandler *handler.AuthHandler,
	serviceHandler *handler.ServiceHandler,
	stylistHandler *handler.StylistHandler,
	bookingHandler *handler.BookingHandler,
	statsHandler *handler.StatisticsHandler,
	uploadHandler *handler.UploadHandler,
) *gin.Engine {
	router := gin.New()

	// Middleware
	router.Use(middleware.Logger())
	router.Use(middleware.CORS(&cfg.CORS))
	router.Use(gin.Recovery())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Public routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Public service routes
		services := v1.Group("/services")
		{
			services.GET("", serviceHandler.ListServices)
			services.GET("/:id", serviceHandler.GetService)
		}

		// Public stylist routes
		stylists := v1.Group("/stylists")
		{
			stylists.GET("", stylistHandler.ListStylists)
			stylists.GET("/:id", stylistHandler.GetStylist)
			stylists.GET("/:id/schedules", stylistHandler.GetSchedules)
		}

		// Protected routes (require authentication)
		protected := v1.Group("")
		protected.Use(middleware.AuthRequired(jwtManager))
		{
			// User profile
			protected.GET("/auth/profile", authHandler.GetProfile)

			// Bookings
			bookings := protected.Group("/bookings")
			{
				bookings.GET("", bookingHandler.ListBookings)
				bookings.GET("/:id", bookingHandler.GetBooking)
				bookings.POST("", bookingHandler.CreateBooking)
				bookings.POST("/:id/cancel", bookingHandler.CancelBooking)
			}

			// Upload
			upload := protected.Group("/upload")
			{
				upload.POST("/image", uploadHandler.UploadImage)
			}
		}

		// Admin routes (require admin role)
		admin := v1.Group("/admin")
		admin.Use(middleware.AdminRequired(jwtManager))
		{
			// Service management
			admin.POST("/services", serviceHandler.CreateService)
			admin.PUT("/services/:id", serviceHandler.UpdateService)
			admin.DELETE("/services/:id", serviceHandler.DeleteService)

			// Stylist management
			admin.POST("/stylists", stylistHandler.CreateStylist)
			admin.PUT("/stylists/:id", stylistHandler.UpdateStylist)
			admin.DELETE("/stylists/:id", stylistHandler.DeleteStylist)
			admin.POST("/stylists/:id/schedules", stylistHandler.CreateSchedule)

			// Booking management
			admin.PATCH("/bookings/:id/status", bookingHandler.UpdateBookingStatus)

			// Statistics
			admin.GET("/statistics/dashboard", statsHandler.GetDashboardStats)
			admin.GET("/statistics/revenue", statsHandler.GetRevenueReport)

			// Upload management
			admin.DELETE("/upload/image", uploadHandler.DeleteImage)
		}
	}

	return router
}
