package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"linda-salon-api/internal/repository"
)

type StatisticsHandler struct {
	bookingRepo *repository.BookingRepository
	stylistRepo *repository.StylistRepository
}

func NewStatisticsHandler(bookingRepo *repository.BookingRepository, stylistRepo *repository.StylistRepository) *StatisticsHandler {
	return &StatisticsHandler{
		bookingRepo: bookingRepo,
		stylistRepo: stylistRepo,
	}
}

type DashboardStats struct {
	TodayBookings  int64                    `json:"today_bookings"`
	WeekBookings   int64                    `json:"week_bookings"`
	MonthBookings  int64                    `json:"month_bookings"`
	TodayRevenue   int                      `json:"today_revenue"`
	MonthRevenue   int                      `json:"month_revenue"`
	RevenueByDay   []map[string]interface{} `json:"revenue_by_day"`
	PopularServices []map[string]interface{} `json:"popular_services"`
	TopStylists    []map[string]interface{} `json:"top_stylists"`
}

// GetDashboardStats godoc
// @Summary Get dashboard statistics (admin only)
// @Tags statistics
// @Security BearerAuth
// @Produce json
// @Success 200 {object} DashboardStats
// @Router /statistics/dashboard [get]
func (h *StatisticsHandler) GetDashboardStats(c *gin.Context) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Start of week (Monday)
	weekStart := today.AddDate(0, 0, -int(today.Weekday())+1)
	if today.Weekday() == time.Sunday {
		weekStart = weekStart.AddDate(0, 0, -7)
	}

	// Start of month
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// Today's bookings count
	todayBookings, err := h.bookingRepo.CountByDateRange(today, today, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch today's bookings"})
		return
	}

	// Week's bookings count
	weekEnd := weekStart.AddDate(0, 0, 6)
	weekBookings, err := h.bookingRepo.CountByDateRange(weekStart, weekEnd, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch week's bookings"})
		return
	}

	// Month's bookings count
	monthEnd := monthStart.AddDate(0, 1, -1)
	monthBookings, err := h.bookingRepo.CountByDateRange(monthStart, monthEnd, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch month's bookings"})
		return
	}

	// Today's revenue
	todayRevenue, err := h.bookingRepo.GetRevenueByDateRange(today, today)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch today's revenue"})
		return
	}

	// Month's revenue
	monthRevenue, err := h.bookingRepo.GetRevenueByDateRange(monthStart, monthEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch month's revenue"})
		return
	}

	// Revenue by day (last 30 days)
	thirtyDaysAgo := today.AddDate(0, 0, -29)
	revenueByDay, err := h.bookingRepo.GetRevenueByDay(thirtyDaysAgo, today)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch revenue by day"})
		return
	}

	// Popular services (this month)
	popularServices, err := h.bookingRepo.GetPopularServices(5, monthStart, monthEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch popular services"})
		return
	}

	// Top stylists (this month)
	topStylists, err := h.stylistRepo.GetTopStylists(5, monthStart, monthEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch top stylists"})
		return
	}

	stats := DashboardStats{
		TodayBookings:   todayBookings,
		WeekBookings:    weekBookings,
		MonthBookings:   monthBookings,
		TodayRevenue:    todayRevenue,
		MonthRevenue:    monthRevenue,
		RevenueByDay:    revenueByDay,
		PopularServices: popularServices,
		TopStylists:     topStylists,
	}

	c.JSON(http.StatusOK, stats)
}

// GetRevenueReport godoc
// @Summary Get revenue report (admin only)
// @Tags statistics
// @Security BearerAuth
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{}
// @Router /statistics/revenue [get]
func (h *StatisticsHandler) GetRevenueReport(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format"})
		return
	}

	// Total revenue
	totalRevenue, err := h.bookingRepo.GetRevenueByDateRange(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch total revenue"})
		return
	}

	// Revenue by day
	revenueByDay, err := h.bookingRepo.GetRevenueByDay(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch revenue by day"})
		return
	}

	// Booking count
	bookingCount, err := h.bookingRepo.CountByDateRange(startDate, endDate, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch booking count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"start_date":     startDateStr,
		"end_date":       endDateStr,
		"total_revenue":  totalRevenue,
		"booking_count":  bookingCount,
		"revenue_by_day": revenueByDay,
	})
}
