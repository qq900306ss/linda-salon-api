package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/qq900306ss/linda-salon-api/internal/repository"
	"github.com/qq900306ss/linda-salon-api/internal/service"
)

// StatisticsHandler handles admin statistics and customer aggregation routes.
type StatisticsHandler struct {
	bookingRepo *repository.BookingRepository
}

// NewStatisticsHandler creates a StatisticsHandler.
func NewStatisticsHandler(bookingRepo *repository.BookingRepository) *StatisticsHandler {
	return &StatisticsHandler{bookingRepo: bookingRepo}
}

// GetDashboard handles GET /api/admin/statistics/dashboard.
func (h *StatisticsHandler) GetDashboard(c *gin.Context) {
	bookings, err := h.bookingRepo.ScanAll(c.Request.Context())
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch bookings")
		return
	}
	now := time.Now().In(service.TaipeiLocation())
	OK(c, http.StatusOK, service.BuildDashboardStats(bookings, now))
}

// GetRevenue handles GET /api/admin/statistics/revenue?from=&to=.
func (h *StatisticsHandler) GetRevenue(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	now := time.Now().In(service.TaipeiLocation())
	if from == "" || to == "" {
		// Default: last 30 days.
		to = now.Format("2006-01-02")
		from = now.AddDate(0, 0, -29).Format("2006-01-02")
	}
	if !dateRe.MatchString(from) || !dateRe.MatchString(to) {
		Fail(c, http.StatusBadRequest, "INVALID_REQUEST", "from/to must be in YYYY-MM-DD format")
		return
	}

	bookings, err := h.bookingRepo.ListByDateRange(c.Request.Context(), from, to)
	if err != nil {
		Fail(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	OK(c, http.StatusOK, service.BuildRevenueReport(bookings, from, to))
}

// ListCustomers handles GET /api/admin/customers.
func (h *StatisticsHandler) ListCustomers(c *gin.Context) {
	bookings, err := h.bookingRepo.ScanAll(c.Request.Context())
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch bookings")
		return
	}
	OK(c, http.StatusOK, service.BuildCustomerSummaries(bookings))
}
