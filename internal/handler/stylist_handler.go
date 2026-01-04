package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"linda-salon-api/internal/model"
	"linda-salon-api/internal/repository"
)

type StylistHandler struct {
	stylistRepo *repository.StylistRepository
}

func NewStylistHandler(stylistRepo *repository.StylistRepository) *StylistHandler {
	return &StylistHandler{stylistRepo: stylistRepo}
}

type CreateStylistRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Specialty   string `json:"specialty"`
	Experience  int    `json:"experience" binding:"omitempty,min=0"`
	Avatar      string `json:"avatar"`
}

type UpdateStylistRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Specialty   string `json:"specialty"`
	Experience  int    `json:"experience" binding:"omitempty,min=0"`
	Avatar      string `json:"avatar"`
	IsActive    *bool  `json:"is_active"`
}

type CreateScheduleRequest struct {
	DayOfWeek int    `json:"day_of_week" binding:"required,min=0,max=6"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
}

// ListStylists godoc
// @Summary List all stylists
// @Tags stylists
// @Produce json
// @Param active_only query bool false "Show only active stylists" default(true)
// @Success 200 {array} model.Stylist
// @Router /stylists [get]
func (h *StylistHandler) ListStylists(c *gin.Context) {
	activeOnly := c.DefaultQuery("active_only", "true") == "true"

	stylists, err := h.stylistRepo.List(activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stylists"})
		return
	}

	c.JSON(http.StatusOK, stylists)
}

// GetStylist godoc
// @Summary Get stylist by ID
// @Tags stylists
// @Produce json
// @Param id path int true "Stylist ID"
// @Success 200 {object} model.Stylist
// @Router /stylists/{id} [get]
func (h *StylistHandler) GetStylist(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stylist ID"})
		return
	}

	stylist, err := h.stylistRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stylist"})
		return
	}
	if stylist == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stylist not found"})
		return
	}

	c.JSON(http.StatusOK, stylist)
}

// CreateStylist godoc
// @Summary Create a new stylist (admin only)
// @Tags stylists
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateStylistRequest true "Stylist details"
// @Success 201 {object} model.Stylist
// @Router /stylists [post]
func (h *StylistHandler) CreateStylist(c *gin.Context) {
	var req CreateStylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stylist := &model.Stylist{
		Name:        req.Name,
		Description: req.Description,
		Specialty:   req.Specialty,
		Experience:  req.Experience,
		Avatar:      req.Avatar,
		IsActive:    true,
	}

	if err := h.stylistRepo.Create(stylist); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create stylist"})
		return
	}

	c.JSON(http.StatusCreated, stylist)
}

// UpdateStylist godoc
// @Summary Update stylist (admin only)
// @Tags stylists
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Stylist ID"
// @Param request body UpdateStylistRequest true "Stylist details"
// @Success 200 {object} model.Stylist
// @Router /stylists/{id} [put]
func (h *StylistHandler) UpdateStylist(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stylist ID"})
		return
	}

	stylist, err := h.stylistRepo.GetByID(uint(id))
	if err != nil || stylist == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stylist not found"})
		return
	}

	var req UpdateStylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		stylist.Name = req.Name
	}
	if req.Description != "" {
		stylist.Description = req.Description
	}
	if req.Specialty != "" {
		stylist.Specialty = req.Specialty
	}
	if req.Experience > 0 {
		stylist.Experience = req.Experience
	}
	if req.Avatar != "" {
		stylist.Avatar = req.Avatar
	}
	if req.IsActive != nil {
		stylist.IsActive = *req.IsActive
	}

	if err := h.stylistRepo.Update(stylist); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stylist"})
		return
	}

	c.JSON(http.StatusOK, stylist)
}

// DeleteStylist godoc
// @Summary Delete stylist (admin only)
// @Tags stylists
// @Security BearerAuth
// @Param id path int true "Stylist ID"
// @Success 204
// @Router /stylists/{id} [delete]
func (h *StylistHandler) DeleteStylist(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stylist ID"})
		return
	}

	if err := h.stylistRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete stylist"})
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateSchedule godoc
// @Summary Create stylist schedule (admin only)
// @Tags stylists
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Stylist ID"
// @Param request body CreateScheduleRequest true "Schedule details"
// @Success 201 {object} model.StylistSchedule
// @Router /stylists/{id}/schedules [post]
func (h *StylistHandler) CreateSchedule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stylist ID"})
		return
	}

	var req CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schedule := &model.StylistSchedule{
		StylistID: uint(id),
		DayOfWeek: req.DayOfWeek,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		IsActive:  true,
	}

	if err := h.stylistRepo.CreateSchedule(schedule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create schedule"})
		return
	}

	c.JSON(http.StatusCreated, schedule)
}

// GetSchedules godoc
// @Summary Get stylist schedules
// @Tags stylists
// @Produce json
// @Param id path int true "Stylist ID"
// @Success 200 {array} model.StylistSchedule
// @Router /stylists/{id}/schedules [get]
func (h *StylistHandler) GetSchedules(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stylist ID"})
		return
	}

	schedules, err := h.stylistRepo.GetSchedulesByStylistID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch schedules"})
		return
	}

	c.JSON(http.StatusOK, schedules)
}

// DeleteSchedule godoc
// @Summary Delete stylist schedule (admin only)
// @Tags stylists
// @Security BearerAuth
// @Param id path int true "Schedule ID"
// @Success 204
// @Router /stylists/schedules/{id} [delete]
func (h *StylistHandler) DeleteSchedule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schedule ID"})
		return
	}

	if err := h.stylistRepo.DeleteSchedule(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete schedule"})
		return
	}

	c.Status(http.StatusNoContent)
}
