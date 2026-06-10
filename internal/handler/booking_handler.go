package handler

import (
	"net/http"
	"regexp"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/qq900306ss/linda-salon-api/internal/model"
	"github.com/qq900306ss/linda-salon-api/internal/repository"
	"github.com/qq900306ss/linda-salon-api/internal/service"
)

var (
	dateRe  = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	timeRe  = regexp.MustCompile(`^\d{2}:\d{2}$`)
	phoneRe = regexp.MustCompile(`^[0-9+\-\s()]{8,20}$`)
)

// BookingHandler handles booking and timeslot routes.
type BookingHandler struct {
	bookingRepo  *repository.BookingRepository
	serviceRepo  *repository.ServiceRepository
	stylistRepo  *repository.StylistRepository
	settingsRepo *repository.SettingsRepository
}

// NewBookingHandler creates a BookingHandler.
func NewBookingHandler(
	bookingRepo *repository.BookingRepository,
	serviceRepo *repository.ServiceRepository,
	stylistRepo *repository.StylistRepository,
	settingsRepo *repository.SettingsRepository,
) *BookingHandler {
	return &BookingHandler{
		bookingRepo:  bookingRepo,
		serviceRepo:  serviceRepo,
		stylistRepo:  stylistRepo,
		settingsRepo: settingsRepo,
	}
}

type createBookingRequest struct {
	ServiceID string `json:"serviceId" binding:"required"`
	StylistID string `json:"stylistId" binding:"required"`
	Date      string `json:"date" binding:"required"`
	Time      string `json:"time" binding:"required"`
	Customer  struct {
		Name  string `json:"name" binding:"required"`
		Phone string `json:"phone" binding:"required"`
		Email string `json:"email"`
		Notes string `json:"notes"`
	} `json:"customer" binding:"required"`
}

// slotContext loads everything needed to compute slots for a stylist/date.
type slotContext struct {
	settings *model.Settings
	stylist  *model.Stylist
	bookings []model.Booking // stylist's non-cancelled bookings on the date
}

func (h *BookingHandler) loadSlotContext(c *gin.Context, stylistID, date string) (*slotContext, bool) {
	ctx := c.Request.Context()

	settings, err := h.settingsRepo.Get(ctx)
	if err != nil || settings == nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load settings")
		return nil, false
	}

	stylist, err := h.stylistRepo.GetByID(ctx, stylistID)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load stylist")
		return nil, false
	}
	if stylist == nil || !stylist.IsActive {
		Fail(c, http.StatusNotFound, "NOT_FOUND", "Stylist not found")
		return nil, false
	}

	dayBookings, err := h.bookingRepo.ListByDate(ctx, date)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load bookings")
		return nil, false
	}
	var stylistBookings []model.Booking
	for _, b := range dayBookings {
		if b.StylistID == stylistID && b.Status != model.BookingStatusCancelled {
			stylistBookings = append(stylistBookings, b)
		}
	}

	return &slotContext{settings: settings, stylist: stylist, bookings: stylistBookings}, true
}

// GetTimeSlots handles GET /api/timeslots?stylistId=&date=&serviceId=.
func (h *BookingHandler) GetTimeSlots(c *gin.Context) {
	stylistID := c.Query("stylistId")
	date := c.Query("date")
	serviceID := c.Query("serviceId")

	if stylistID == "" || !dateRe.MatchString(date) {
		Fail(c, http.StatusBadRequest, "INVALID_REQUEST", "stylistId and date (YYYY-MM-DD) are required")
		return
	}

	duration := 0
	if serviceID != "" {
		svc, err := h.serviceRepo.GetByID(c.Request.Context(), serviceID)
		if err != nil {
			Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load service")
			return
		}
		if svc == nil {
			Fail(c, http.StatusNotFound, "NOT_FOUND", "Service not found")
			return
		}
		duration = svc.DurationMinutes
	}

	sc, ok := h.loadSlotContext(c, stylistID, date)
	if !ok {
		return
	}

	slots, err := service.GenerateTimeSlots(service.SlotQuery{
		Settings:        *sc.settings,
		Stylist:         *sc.stylist,
		Date:            date,
		DurationMinutes: duration,
		Bookings:        sc.bookings,
		Now:             time.Now().In(service.TaipeiLocation()),
	})
	if err != nil {
		Fail(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	OK(c, http.StatusOK, gin.H{"date": date, "slots": slots})
}

// CreateBooking handles POST /api/bookings.
func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var req createBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	if !dateRe.MatchString(req.Date) {
		Fail(c, http.StatusBadRequest, "INVALID_REQUEST", "date must be in YYYY-MM-DD format")
		return
	}
	if !timeRe.MatchString(req.Time) {
		Fail(c, http.StatusBadRequest, "INVALID_REQUEST", "time must be in HH:MM format")
		return
	}
	if !phoneRe.MatchString(req.Customer.Phone) {
		Fail(c, http.StatusBadRequest, "INVALID_REQUEST", "phone format is invalid")
		return
	}

	ctx := c.Request.Context()
	svc, err := h.serviceRepo.GetByID(ctx, req.ServiceID)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load service")
		return
	}
	if svc == nil || !svc.IsActive {
		Fail(c, http.StatusNotFound, "NOT_FOUND", "Service not found")
		return
	}

	sc, ok := h.loadSlotContext(c, req.StylistID, req.Date)
	if !ok {
		return
	}

	slots, err := service.GenerateTimeSlots(service.SlotQuery{
		Settings:        *sc.settings,
		Stylist:         *sc.stylist,
		Date:            req.Date,
		DurationMinutes: svc.DurationMinutes,
		Bookings:        sc.bookings,
		Now:             time.Now().In(service.TaipeiLocation()),
	})
	if err != nil {
		Fail(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	var requested *service.TimeSlot
	for i := range slots {
		if slots[i].Time == req.Time {
			requested = &slots[i]
			break
		}
	}
	if requested == nil {
		Fail(c, http.StatusBadRequest, "INVALID_TIME", "Requested time is not a valid slot")
		return
	}
	if !requested.Available {
		Fail(c, http.StatusConflict, "SLOT_UNAVAILABLE", "該時段已無法預約，請選擇其他時間")
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	booking := model.Booking{
		ID:              uuid.NewString(),
		ServiceID:       svc.ID,
		ServiceName:     svc.Name,
		StylistID:       sc.stylist.ID,
		StylistName:     sc.stylist.Name,
		Date:            req.Date,
		Time:            req.Time,
		DurationMinutes: svc.DurationMinutes,
		Price:           svc.Price,
		Status:          model.BookingStatusPending,
		Customer: model.Customer{
			Name:  req.Customer.Name,
			Phone: req.Customer.Phone,
			Email: req.Customer.Email,
			Notes: req.Customer.Notes,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := h.bookingRepo.Create(ctx, &booking); err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create booking")
		return
	}
	OK(c, http.StatusCreated, booking)
}

// LookupBookings handles GET /api/bookings/lookup?phone=.
func (h *BookingHandler) LookupBookings(c *gin.Context) {
	phone := c.Query("phone")
	if !phoneRe.MatchString(phone) {
		Fail(c, http.StatusBadRequest, "INVALID_REQUEST", "phone query parameter is required")
		return
	}
	bookings, err := h.bookingRepo.ListByPhone(c.Request.Context(), phone, 20)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch bookings")
		return
	}
	if bookings == nil {
		bookings = []model.Booking{}
	}
	// Newest first: by date desc, then time desc.
	sort.Slice(bookings, func(i, j int) bool {
		if bookings[i].Date != bookings[j].Date {
			return bookings[i].Date > bookings[j].Date
		}
		return bookings[i].Time > bookings[j].Time
	})
	OK(c, http.StatusOK, bookings)
}

// ListBookingsAdmin handles GET /api/admin/bookings?date=&from=&to=&status=&stylistId=.
func (h *BookingHandler) ListBookingsAdmin(c *gin.Context) {
	ctx := c.Request.Context()
	date := c.Query("date")
	from := c.Query("from")
	to := c.Query("to")
	status := c.Query("status")
	stylistID := c.Query("stylistId")

	var bookings []model.Booking
	var err error
	switch {
	case date != "":
		if !dateRe.MatchString(date) {
			Fail(c, http.StatusBadRequest, "INVALID_REQUEST", "date must be in YYYY-MM-DD format")
			return
		}
		bookings, err = h.bookingRepo.ListByDate(ctx, date)
	case from != "" && to != "":
		if !dateRe.MatchString(from) || !dateRe.MatchString(to) {
			Fail(c, http.StatusBadRequest, "INVALID_REQUEST", "from/to must be in YYYY-MM-DD format")
			return
		}
		bookings, err = h.bookingRepo.ListByDateRange(ctx, from, to)
	case from != "" || to != "":
		Fail(c, http.StatusBadRequest, "INVALID_REQUEST", "both from and to are required for a range query")
		return
	default:
		today := time.Now().In(service.TaipeiLocation()).Format("2006-01-02")
		bookings, err = h.bookingRepo.ListByDate(ctx, today)
	}
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch bookings")
		return
	}

	filtered := []model.Booking{}
	for _, b := range bookings {
		if status != "" && b.Status != status {
			continue
		}
		if stylistID != "" && b.StylistID != stylistID {
			continue
		}
		filtered = append(filtered, b)
	}
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Date != filtered[j].Date {
			return filtered[i].Date < filtered[j].Date
		}
		return filtered[i].Time < filtered[j].Time
	})
	OK(c, http.StatusOK, filtered)
}

type updateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateBookingStatus handles PATCH /api/admin/bookings/:id/status.
func (h *BookingHandler) UpdateBookingStatus(c *gin.Context) {
	id := c.Param("id")
	var req updateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "INVALID_REQUEST", "status is required")
		return
	}
	if !model.ValidBookingStatus(req.Status) {
		Fail(c, http.StatusBadRequest, "INVALID_REQUEST", "status must be pending, confirmed, completed or cancelled")
		return
	}

	ctx := c.Request.Context()
	booking, err := h.bookingRepo.GetByID(ctx, id)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch booking")
		return
	}
	if booking == nil {
		Fail(c, http.StatusNotFound, "NOT_FOUND", "Booking not found")
		return
	}
	if err := h.bookingRepo.UpdateStatus(ctx, id, req.Status); err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update booking")
		return
	}
	booking.Status = req.Status
	booking.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	OK(c, http.StatusOK, booking)
}

// CancelBooking handles DELETE /api/admin/bookings/:id (sets status cancelled).
func (h *BookingHandler) CancelBooking(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()
	booking, err := h.bookingRepo.GetByID(ctx, id)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch booking")
		return
	}
	if booking == nil {
		Fail(c, http.StatusNotFound, "NOT_FOUND", "Booking not found")
		return
	}
	if err := h.bookingRepo.UpdateStatus(ctx, id, model.BookingStatusCancelled); err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to cancel booking")
		return
	}
	booking.Status = model.BookingStatusCancelled
	booking.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	OK(c, http.StatusOK, booking)
}
