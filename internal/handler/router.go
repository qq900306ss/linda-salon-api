package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/qq900306ss/linda-salon-api/internal/auth"
	"github.com/qq900306ss/linda-salon-api/internal/middleware"
)

// Handlers bundles all route handlers for router setup.
type Handlers struct {
	Auth       *AuthHandler
	Service    *ServiceHandler
	Stylist    *StylistHandler
	Booking    *BookingHandler
	Settings   *SettingsHandler
	Statistics *StatisticsHandler
	Upload     *UploadHandler
}

// NewRouter builds the Gin engine with all routes and middleware.
func NewRouter(h Handlers, jwtManager *auth.JWTManager, allowedOrigins string) *gin.Engine {
	router := gin.New()
	router.Use(middleware.Logger())
	router.Use(middleware.CORS(allowedOrigins))
	router.Use(gin.Recovery())

	// Health check.
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now().UTC().Format(time.RFC3339)})
	})

	api := router.Group("/api")
	{
		// Public routes.
		api.GET("/services", h.Service.ListServices)
		api.GET("/services/:id", h.Service.GetService)

		api.GET("/stylists", h.Stylist.ListStylists)
		api.GET("/stylists/:id", h.Stylist.GetStylist)

		api.GET("/timeslots", h.Booking.GetTimeSlots)

		api.POST("/bookings", h.Booking.CreateBooking)
		api.GET("/bookings/lookup", h.Booking.LookupBookings)

		api.POST("/auth/admin/login", h.Auth.AdminLogin)

		// Admin routes.
		admin := api.Group("/admin")
		admin.Use(middleware.AdminRequired(jwtManager))
		{
			admin.GET("/bookings", h.Booking.ListBookingsAdmin)
			admin.PATCH("/bookings/:id/status", h.Booking.UpdateBookingStatus)
			admin.DELETE("/bookings/:id", h.Booking.CancelBooking)

			admin.GET("/services", h.Service.ListServicesAdmin)
			admin.POST("/services", h.Service.CreateService)
			admin.PUT("/services/:id", h.Service.UpdateService)
			admin.DELETE("/services/:id", h.Service.DeleteService)

			admin.GET("/stylists", h.Stylist.ListStylistsAdmin)
			admin.POST("/stylists", h.Stylist.CreateStylist)
			admin.PUT("/stylists/:id", h.Stylist.UpdateStylist)
			admin.DELETE("/stylists/:id", h.Stylist.DeleteStylist)
			admin.GET("/stylists/:id/schedule", h.Stylist.GetSchedule)
			admin.PUT("/stylists/:id/schedule", h.Stylist.UpdateSchedule)

			admin.GET("/customers", h.Statistics.ListCustomers)
			admin.GET("/statistics/dashboard", h.Statistics.GetDashboard)
			admin.GET("/statistics/revenue", h.Statistics.GetRevenue)

			admin.POST("/uploads", h.Upload.CreateUpload)

			admin.GET("/settings", h.Settings.GetSettings)
			admin.PUT("/settings", h.Settings.UpdateSettings)
		}
	}

	router.NoRoute(func(c *gin.Context) {
		Fail(c, http.StatusNotFound, "NOT_FOUND", "Route not found")
	})

	return router
}
