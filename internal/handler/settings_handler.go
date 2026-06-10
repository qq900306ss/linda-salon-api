package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/qq900306ss/linda-salon-api/internal/model"
	"github.com/qq900306ss/linda-salon-api/internal/repository"
	"github.com/qq900306ss/linda-salon-api/internal/service"
)

// SettingsHandler handles salon settings routes.
type SettingsHandler struct {
	settingsRepo *repository.SettingsRepository
}

// NewSettingsHandler creates a SettingsHandler.
func NewSettingsHandler(settingsRepo *repository.SettingsRepository) *SettingsHandler {
	return &SettingsHandler{settingsRepo: settingsRepo}
}

// GetSettings handles GET /api/admin/settings.
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	settings, err := h.settingsRepo.Get(c.Request.Context())
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch settings")
		return
	}
	if settings == nil {
		Fail(c, http.StatusNotFound, "NOT_FOUND", "Settings not found")
		return
	}
	OK(c, http.StatusOK, settings)
}

// UpdateSettings handles PUT /api/admin/settings.
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	var settings model.Settings
	if err := c.ShouldBindJSON(&settings); err != nil {
		Fail(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	if _, err := service.ParseHM(settings.OpenTime); err != nil {
		Fail(c, http.StatusBadRequest, "INVALID_REQUEST", "openTime must be in HH:MM format")
		return
	}
	if _, err := service.ParseHM(settings.CloseTime); err != nil {
		Fail(c, http.StatusBadRequest, "INVALID_REQUEST", "closeTime must be in HH:MM format")
		return
	}
	if settings.SlotIntervalMinutes <= 0 {
		Fail(c, http.StatusBadRequest, "INVALID_REQUEST", "slotIntervalMinutes must be positive")
		return
	}
	for _, d := range settings.ClosedWeekdays {
		if d < 0 || d > 6 {
			Fail(c, http.StatusBadRequest, "INVALID_REQUEST", "closedWeekdays must contain integers 0-6")
			return
		}
	}
	settings.Normalize()
	if err := h.settingsRepo.Put(c.Request.Context(), &settings); err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update settings")
		return
	}
	OK(c, http.StatusOK, settings)
}
