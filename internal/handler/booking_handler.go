package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"linda-salon-api/internal/middleware"
	"linda-salon-api/internal/model"
	"linda-salon-api/internal/repository"
)

type BookingHandler struct {
	bookingRepo *repository.BookingRepository
	serviceRepo *repository.ServiceRepository
	stylistRepo *repository.StylistRepository
	userRepo    *repository.UserRepository
}

func NewBookingHandler(
	bookingRepo *repository.BookingRepository,
	serviceRepo *repository.ServiceRepository,
	stylistRepo *repository.StylistRepository,
	userRepo    *repository.UserRepository,
) *BookingHandler {
	return &BookingHandler{
		bookingRepo: bookingRepo,
		serviceRepo: serviceRepo,
		stylistRepo: stylistRepo,
		userRepo:    userRepo,
	}
}

type CreateBookingRequest struct {
	ServiceIDs    []uint `json:"service_ids" binding:"required,min=1"` // 支援多個服務
	StylistID     uint   `json:"stylist_id" binding:"required"`
	Date          string `json:"date" binding:"required"`     // YYYY-MM-DD
	StartTime     string `json:"start_time" binding:"required"` // HH:MM
	Notes         string `json:"notes"`
	CustomerName  string `json:"customer_name"`  // 可選：覆蓋用戶姓名
	CustomerPhone string `json:"customer_phone"` // 可選：覆蓋用戶電話
	CustomerEmail string `json:"customer_email"` // 可選：覆蓋用戶信箱
}

type UpdateBookingRequest struct {
	ServiceID *uint   `json:"service_id"`
	StylistID *uint   `json:"stylist_id"`
	Date      *string `json:"date"`
	StartTime *string `json:"start_time"`
	Status    *string `json:"status"`
	Notes     *string `json:"notes"`
}

// ListBookings godoc
// @Summary List bookings
// @Tags bookings
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by status"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} model.Booking
// @Router /bookings [get]
func (h *BookingHandler) ListBookings(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetUserRole(c)

	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var startDate, endDate *time.Time
	if sd := c.Query("start_date"); sd != "" {
		t, _ := time.Parse("2006-01-02", sd)
		startDate = &t
	}
	if ed := c.Query("end_date"); ed != "" {
		t, _ := time.Parse("2006-01-02", ed)
		endDate = &t
	}

	var userIDPtr *uint
	// Non-admin users can only see their own bookings
	if role != "admin" {
		userIDPtr = &userID
	}

	bookings, total, err := h.bookingRepo.List(userIDPtr, status, startDate, endDate, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bookings": bookings,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetBooking godoc
// @Summary Get booking by ID
// @Tags bookings
// @Security BearerAuth
// @Produce json
// @Param id path int true "Booking ID"
// @Success 200 {object} model.Booking
// @Router /bookings/{id} [get]
func (h *BookingHandler) GetBooking(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
		return
	}

	booking, err := h.bookingRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch booking"})
		return
	}
	if booking == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	// Check authorization
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetUserRole(c)
	if role != "admin" && booking.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, booking)
}

// CreateBooking godoc
// @Summary Create a new booking
// @Tags bookings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateBookingRequest true "Booking details"
// @Success 201 {object} model.Booking
// @Router /bookings [post]
func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var req CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := middleware.GetUserID(c)

	// Get user info
	user, err := h.userRepo.GetByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	// Get all services info and calculate total duration and price
	var services []model.BookingServiceItem
	var totalDuration int
	var totalPrice int

	for _, serviceID := range req.ServiceIDs {
		service, err := h.serviceRepo.GetByID(serviceID)
		if err != nil || service == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid service ID: %d", serviceID)})
			return
		}

		services = append(services, model.BookingServiceItem{
			ID:       service.ID,
			Name:     service.Name,
			Price:    service.Price,
			Duration: service.Duration,
		})

		totalDuration += service.Duration
		totalPrice += service.Price
	}

	// Get stylist info
	stylist, err := h.stylistRepo.GetByID(req.StylistID)
	if err != nil || stylist == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stylist"})
		return
	}

	// Parse booking date
	bookingDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	// Calculate end time based on total duration
	startHour, _ := strconv.Atoi(req.StartTime[:2])
	startMin, _ := strconv.Atoi(req.StartTime[3:5])
	endMin := startMin + totalDuration
	endHour := startHour + (endMin / 60)
	endMin = endMin % 60
	endTime := time.Date(0, 0, 0, endHour, endMin, 0, 0, time.UTC).Format("15:04")

	// Check stylist availability
	available, err := h.stylistRepo.IsAvailable(req.StylistID, bookingDate, req.StartTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check availability"})
		return
	}
	if !available {
		c.JSON(http.StatusConflict, gin.H{"error": "Stylist is not available at this time"})
		return
	}

	// 準備客戶資訊（優先使用前端傳來的，否則用資料庫的）
	customerName := req.CustomerName
	if customerName == "" {
		customerName = user.Name
	}

	customerPhone := req.CustomerPhone
	if customerPhone == "" && user.Phone != nil {
		customerPhone = *user.Phone
	}

	customerEmail := req.CustomerEmail
	if customerEmail == "" {
		customerEmail = user.Email
	}

	// Create booking
	booking := &model.Booking{
		UserID:        userID,
		StylistID:     req.StylistID,
		Services:      services,
		BookingDate:   bookingDate,
		StartTime:     req.StartTime,
		EndTime:       endTime,
		Duration:      totalDuration,
		Price:         totalPrice,
		Status:        model.BookingStatusPending,
		Notes:         req.Notes,
		CustomerName:  customerName,
		CustomerPhone: customerPhone,
		CustomerEmail: customerEmail,
	}

	if err := h.bookingRepo.Create(booking); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking"})
		return
	}

	// Fetch complete booking with relations
	booking, _ = h.bookingRepo.GetByID(booking.ID)

	c.JSON(http.StatusCreated, booking)
}

// UpdateBookingStatus godoc
// @Summary Update booking status (admin only)
// @Tags bookings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Booking ID"
// @Param status body map[string]string true "Status"
// @Success 200 {object} model.Booking
// @Router /bookings/{id}/status [patch]
func (h *BookingHandler) UpdateBookingStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		model.BookingStatusPending:   true,
		model.BookingStatusConfirmed: true,
		model.BookingStatusCompleted: true,
		model.BookingStatusCancelled: true,
	}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	if err := h.bookingRepo.UpdateStatus(uint(id), req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	booking, _ := h.bookingRepo.GetByID(uint(id))
	c.JSON(http.StatusOK, booking)
}

// CancelBooking godoc
// @Summary Cancel a booking
// @Tags bookings
// @Security BearerAuth
// @Param id path int true "Booking ID"
// @Success 200 {object} model.Booking
// @Router /bookings/{id}/cancel [post]
func (h *BookingHandler) CancelBooking(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
		return
	}

	booking, err := h.bookingRepo.GetByID(uint(id))
	if err != nil || booking == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	// Check authorization
	userID, _ := middleware.GetUserID(c)
	role, _ := middleware.GetUserRole(c)
	if role != "admin" && booking.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Check if cancellable
	if !booking.IsCancellable() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Booking cannot be cancelled"})
		return
	}

	if err := h.bookingRepo.UpdateStatus(uint(id), model.BookingStatusCancelled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel booking"})
		return
	}

	booking, _ = h.bookingRepo.GetByID(uint(id))
	c.JSON(http.StatusOK, booking)
}
