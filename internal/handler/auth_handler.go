package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/qq900306ss/linda-salon-api/internal/auth"
	"github.com/qq900306ss/linda-salon-api/internal/repository"
)

// AuthHandler handles admin authentication.
type AuthHandler struct {
	userRepo   *repository.UserRepository
	jwtManager *auth.JWTManager
}

// NewAuthHandler creates an AuthHandler.
func NewAuthHandler(userRepo *repository.UserRepository, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, jwtManager: jwtManager}
}

type adminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AdminLogin handles POST /api/auth/admin/login.
func (h *AuthHandler) AdminLogin(c *gin.Context) {
	var req adminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "INVALID_REQUEST", "username and password are required")
		return
	}

	user, err := h.userRepo.GetByUsername(c.Request.Context(), req.Username)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to verify credentials")
		return
	}
	if user == nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		Fail(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "帳號或密碼錯誤")
		return
	}

	token, err := h.jwtManager.GenerateToken(user.Username, user.Role)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate token")
		return
	}

	OK(c, http.StatusOK, gin.H{
		"token":     token,
		"expiresIn": int64(auth.TokenExpiry.Seconds()),
		"user": gin.H{
			"username": user.Username,
			"role":     user.Role,
		},
	})
}
