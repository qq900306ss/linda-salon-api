package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"linda-salon-api/internal/model"
	"linda-salon-api/internal/repository"
)

type ServiceHandler struct {
	serviceRepo *repository.ServiceRepository
}

func NewServiceHandler(serviceRepo *repository.ServiceRepository) *ServiceHandler {
	return &ServiceHandler{serviceRepo: serviceRepo}
}

type CreateServiceRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Category    string `json:"category" binding:"required"`
	Price       int    `json:"price" binding:"required,min=0"`
	Duration    int    `json:"duration" binding:"required,min=1"`
	ImageURL    string `json:"image_url"`
}

type UpdateServiceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Price       int    `json:"price" binding:"omitempty,min=0"`
	Duration    int    `json:"duration" binding:"omitempty,min=1"`
	ImageURL    string `json:"image_url"`
	IsActive    *bool  `json:"is_active"`
}

// ListServices godoc
// @Summary List all services
// @Tags services
// @Produce json
// @Param category query string false "Filter by category"
// @Param active_only query bool false "Show only active services"
// @Success 200 {array} model.Service
// @Router /services [get]
func (h *ServiceHandler) ListServices(c *gin.Context) {
	category := c.Query("category")
	activeOnly := c.DefaultQuery("active_only", "true") == "true"

	services, err := h.serviceRepo.List(category, activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch services"})
		return
	}

	c.JSON(http.StatusOK, services)
}

// GetService godoc
// @Summary Get service by ID
// @Tags services
// @Produce json
// @Param id path int true "Service ID"
// @Success 200 {object} model.Service
// @Router /services/{id} [get]
func (h *ServiceHandler) GetService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	service, err := h.serviceRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch service"})
		return
	}
	if service == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	c.JSON(http.StatusOK, service)
}

// CreateService godoc
// @Summary Create a new service (admin only)
// @Tags services
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateServiceRequest true "Service details"
// @Success 201 {object} model.Service
// @Router /services [post]
func (h *ServiceHandler) CreateService(c *gin.Context) {
	var req CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service := &model.Service{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Price:       req.Price,
		Duration:    req.Duration,
		ImageURL:    req.ImageURL,
		IsActive:    true,
	}

	if err := h.serviceRepo.Create(service); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create service"})
		return
	}

	c.JSON(http.StatusCreated, service)
}

// UpdateService godoc
// @Summary Update service (admin only)
// @Tags services
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Service ID"
// @Param request body UpdateServiceRequest true "Service details"
// @Success 200 {object} model.Service
// @Router /services/{id} [put]
func (h *ServiceHandler) UpdateService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	service, err := h.serviceRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch service"})
		return
	}
	if service == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	var req UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields
	if req.Name != "" {
		service.Name = req.Name
	}
	if req.Description != "" {
		service.Description = req.Description
	}
	if req.Category != "" {
		service.Category = req.Category
	}
	if req.Price > 0 {
		service.Price = req.Price
	}
	if req.Duration > 0 {
		service.Duration = req.Duration
	}
	if req.ImageURL != "" {
		service.ImageURL = req.ImageURL
	}
	if req.IsActive != nil {
		service.IsActive = *req.IsActive
	}

	if err := h.serviceRepo.Update(service); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update service"})
		return
	}

	c.JSON(http.StatusOK, service)
}

// DeleteService godoc
// @Summary Delete service (admin only)
// @Tags services
// @Security BearerAuth
// @Param id path int true "Service ID"
// @Success 204
// @Router /services/{id} [delete]
func (h *ServiceHandler) DeleteService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	if err := h.serviceRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete service"})
		return
	}

	c.Status(http.StatusNoContent)
}
