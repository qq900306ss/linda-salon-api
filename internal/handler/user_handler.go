package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"linda-salon-api/internal/repository"
)

type UserHandler struct {
	userRepo    *repository.UserRepository
	bookingRepo *repository.BookingRepository
}

func NewUserHandler(userRepo *repository.UserRepository, bookingRepo *repository.BookingRepository) *UserHandler {
	return &UserHandler{
		userRepo:    userRepo,
		bookingRepo: bookingRepo,
	}
}

// ListUsers godoc
// @Summary List all users (admin only)
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Router /admin/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	users, total, err := h.userRepo.List(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users":  users,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetUser godoc
// @Summary Get user by ID (admin only)
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} model.User
// @Router /admin/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userRepo.GetByID(uint(id))
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetUserBookings godoc
// @Summary Get user's booking history (admin only)
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {array} model.Booking
// @Router /admin/users/{id}/bookings [get]
func (h *UserHandler) GetUserBookings(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	bookings, err := h.bookingRepo.GetUserBookings(uint(id), false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user bookings"})
		return
	}

	c.JSON(http.StatusOK, bookings)
}
