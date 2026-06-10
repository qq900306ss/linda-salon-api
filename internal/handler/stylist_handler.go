package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/qq900306ss/linda-salon-api/internal/model"
	"github.com/qq900306ss/linda-salon-api/internal/repository"
	"github.com/qq900306ss/linda-salon-api/internal/service"
)

// StylistHandler handles stylist routes.
type StylistHandler struct {
	stylistRepo *repository.StylistRepository
}

// NewStylistHandler creates a StylistHandler.
func NewStylistHandler(stylistRepo *repository.StylistRepository) *StylistHandler {
	return &StylistHandler{stylistRepo: stylistRepo}
}

type stylistRequest struct {
	Name            string          `json:"name" binding:"required"`
	Title           string          `json:"title"`
	Bio             string          `json:"bio"`
	Specialties     []string        `json:"specialties"`
	ImageURL        string          `json:"imageUrl"`
	YearsExperience int             `json:"yearsExperience"`
	Rating          float64         `json:"rating"`
	IsActive        *bool           `json:"isActive"`
	Schedule        *model.Schedule `json:"schedule"`
}

func defaultSchedule() model.Schedule {
	return model.Schedule{
		WorkDays:  []int{0, 2, 3, 4, 5, 6},
		StartTime: "10:00",
		EndTime:   "19:00",
		DaysOff:   []string{},
	}
}

func validateSchedule(s *model.Schedule) string {
	for _, d := range s.WorkDays {
		if d < 0 || d > 6 {
			return "workDays must contain integers 0-6"
		}
	}
	if _, err := service.ParseHM(s.StartTime); err != nil {
		return "startTime must be in HH:MM format"
	}
	if _, err := service.ParseHM(s.EndTime); err != nil {
		return "endTime must be in HH:MM format"
	}
	return ""
}

// ListStylists handles GET /api/stylists (active only).
func (h *StylistHandler) ListStylists(c *gin.Context) {
	stylists, err := h.stylistRepo.List(c.Request.Context(), true)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch stylists")
		return
	}
	OK(c, http.StatusOK, stylists)
}

// ListStylistsAdmin handles GET /api/admin/stylists (includes inactive).
func (h *StylistHandler) ListStylistsAdmin(c *gin.Context) {
	stylists, err := h.stylistRepo.List(c.Request.Context(), false)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch stylists")
		return
	}
	OK(c, http.StatusOK, stylists)
}

// GetStylist handles GET /api/stylists/:id.
func (h *StylistHandler) GetStylist(c *gin.Context) {
	stylist, err := h.stylistRepo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch stylist")
		return
	}
	if stylist == nil {
		Fail(c, http.StatusNotFound, "NOT_FOUND", "Stylist not found")
		return
	}
	OK(c, http.StatusOK, stylist)
}

// CreateStylist handles POST /api/admin/stylists.
func (h *StylistHandler) CreateStylist(c *gin.Context) {
	var req stylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	schedule := defaultSchedule()
	if req.Schedule != nil {
		schedule = *req.Schedule
		if msg := validateSchedule(&schedule); msg != "" {
			Fail(c, http.StatusBadRequest, "INVALID_REQUEST", msg)
			return
		}
	}
	stylist := model.Stylist{
		ID:              uuid.NewString(),
		Name:            req.Name,
		Title:           req.Title,
		Bio:             req.Bio,
		Specialties:     req.Specialties,
		ImageURL:        req.ImageURL,
		YearsExperience: req.YearsExperience,
		Rating:          req.Rating,
		IsActive:        isActive,
		Schedule:        schedule,
	}
	stylist.Normalize()
	if err := h.stylistRepo.Put(c.Request.Context(), &stylist); err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create stylist")
		return
	}
	OK(c, http.StatusCreated, stylist)
}

// UpdateStylist handles PUT /api/admin/stylists/:id.
func (h *StylistHandler) UpdateStylist(c *gin.Context) {
	id := c.Param("id")
	existing, err := h.stylistRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch stylist")
		return
	}
	if existing == nil {
		Fail(c, http.StatusNotFound, "NOT_FOUND", "Stylist not found")
		return
	}

	var req stylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	isActive := existing.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	schedule := existing.Schedule
	if req.Schedule != nil {
		schedule = *req.Schedule
		if msg := validateSchedule(&schedule); msg != "" {
			Fail(c, http.StatusBadRequest, "INVALID_REQUEST", msg)
			return
		}
	}
	stylist := model.Stylist{
		ID:              id,
		Name:            req.Name,
		Title:           req.Title,
		Bio:             req.Bio,
		Specialties:     req.Specialties,
		ImageURL:        req.ImageURL,
		YearsExperience: req.YearsExperience,
		Rating:          req.Rating,
		IsActive:        isActive,
		Schedule:        schedule,
	}
	stylist.Normalize()
	if err := h.stylistRepo.Put(c.Request.Context(), &stylist); err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update stylist")
		return
	}
	OK(c, http.StatusOK, stylist)
}

// DeleteStylist handles DELETE /api/admin/stylists/:id.
func (h *StylistHandler) DeleteStylist(c *gin.Context) {
	id := c.Param("id")
	existing, err := h.stylistRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch stylist")
		return
	}
	if existing == nil {
		Fail(c, http.StatusNotFound, "NOT_FOUND", "Stylist not found")
		return
	}
	if err := h.stylistRepo.Delete(c.Request.Context(), id); err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete stylist")
		return
	}
	OK(c, http.StatusOK, gin.H{"deleted": true})
}

// GetSchedule handles GET /api/admin/stylists/:id/schedule.
func (h *StylistHandler) GetSchedule(c *gin.Context) {
	stylist, err := h.stylistRepo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch stylist")
		return
	}
	if stylist == nil {
		Fail(c, http.StatusNotFound, "NOT_FOUND", "Stylist not found")
		return
	}
	OK(c, http.StatusOK, stylist.Schedule)
}

// UpdateSchedule handles PUT /api/admin/stylists/:id/schedule.
func (h *StylistHandler) UpdateSchedule(c *gin.Context) {
	stylist, err := h.stylistRepo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch stylist")
		return
	}
	if stylist == nil {
		Fail(c, http.StatusNotFound, "NOT_FOUND", "Stylist not found")
		return
	}

	var schedule model.Schedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		Fail(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	if msg := validateSchedule(&schedule); msg != "" {
		Fail(c, http.StatusBadRequest, "INVALID_REQUEST", msg)
		return
	}
	schedule.Normalize()
	stylist.Schedule = schedule
	if err := h.stylistRepo.Put(c.Request.Context(), stylist); err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update schedule")
		return
	}
	OK(c, http.StatusOK, stylist.Schedule)
}
