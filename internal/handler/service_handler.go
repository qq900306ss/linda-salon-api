package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/qq900306ss/linda-salon-api/internal/model"
	"github.com/qq900306ss/linda-salon-api/internal/repository"
)

// ServiceHandler handles service routes.
type ServiceHandler struct {
	serviceRepo *repository.ServiceRepository
}

// NewServiceHandler creates a ServiceHandler.
func NewServiceHandler(serviceRepo *repository.ServiceRepository) *ServiceHandler {
	return &ServiceHandler{serviceRepo: serviceRepo}
}

type serviceRequest struct {
	Name            string `json:"name" binding:"required"`
	Description     string `json:"description"`
	Category        string `json:"category"`
	DurationMinutes int    `json:"durationMinutes" binding:"required,min=1"`
	Price           int    `json:"price" binding:"min=0"`
	ImageURL        string `json:"imageUrl"`
	IsActive        *bool  `json:"isActive"`
	SortOrder       int    `json:"sortOrder"`
}

// ListServices handles GET /api/services (active only).
func (h *ServiceHandler) ListServices(c *gin.Context) {
	services, err := h.serviceRepo.List(c.Request.Context(), true)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch services")
		return
	}
	OK(c, http.StatusOK, services)
}

// ListServicesAdmin handles GET /api/admin/services (includes inactive).
func (h *ServiceHandler) ListServicesAdmin(c *gin.Context) {
	services, err := h.serviceRepo.List(c.Request.Context(), false)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch services")
		return
	}
	OK(c, http.StatusOK, services)
}

// GetService handles GET /api/services/:id.
func (h *ServiceHandler) GetService(c *gin.Context) {
	service, err := h.serviceRepo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch service")
		return
	}
	if service == nil {
		Fail(c, http.StatusNotFound, "NOT_FOUND", "Service not found")
		return
	}
	OK(c, http.StatusOK, service)
}

// CreateService handles POST /api/admin/services.
func (h *ServiceHandler) CreateService(c *gin.Context) {
	var req serviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	service := model.Service{
		ID:              uuid.NewString(),
		Name:            req.Name,
		Description:     req.Description,
		Category:        req.Category,
		DurationMinutes: req.DurationMinutes,
		Price:           req.Price,
		ImageURL:        req.ImageURL,
		IsActive:        isActive,
		SortOrder:       req.SortOrder,
	}
	if err := h.serviceRepo.Put(c.Request.Context(), &service); err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create service")
		return
	}
	OK(c, http.StatusCreated, service)
}

// UpdateService handles PUT /api/admin/services/:id.
func (h *ServiceHandler) UpdateService(c *gin.Context) {
	id := c.Param("id")
	existing, err := h.serviceRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch service")
		return
	}
	if existing == nil {
		Fail(c, http.StatusNotFound, "NOT_FOUND", "Service not found")
		return
	}

	var req serviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	isActive := existing.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	service := model.Service{
		ID:              id,
		Name:            req.Name,
		Description:     req.Description,
		Category:        req.Category,
		DurationMinutes: req.DurationMinutes,
		Price:           req.Price,
		ImageURL:        req.ImageURL,
		IsActive:        isActive,
		SortOrder:       req.SortOrder,
	}
	if err := h.serviceRepo.Put(c.Request.Context(), &service); err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update service")
		return
	}
	OK(c, http.StatusOK, service)
}

// DeleteService handles DELETE /api/admin/services/:id.
func (h *ServiceHandler) DeleteService(c *gin.Context) {
	id := c.Param("id")
	existing, err := h.serviceRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch service")
		return
	}
	if existing == nil {
		Fail(c, http.StatusNotFound, "NOT_FOUND", "Service not found")
		return
	}
	if err := h.serviceRepo.Delete(c.Request.Context(), id); err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete service")
		return
	}
	OK(c, http.StatusOK, gin.H{"deleted": true})
}
