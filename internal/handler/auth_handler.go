package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"linda-salon-api/internal/auth"
	"linda-salon-api/internal/model"
	"linda-salon-api/internal/repository"
)

type AuthHandler struct {
	userRepo   *repository.UserRepository
	jwtManager *auth.JWTManager
}

func NewAuthHandler(userRepo *repository.UserRepository, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type GoogleLoginRequest struct {
	GoogleID string `json:"google_id" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required"`
	Picture  string `json:"picture"`
	Phone    string `json:"phone"` // Optional, can be filled later
}

// Register godoc
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register request"
// @Success 201 {object} auth.TokenPair
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if email already exists
	existingUser, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check email"})
		return
	}
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// Check if phone already exists
	existingUser, err = h.userRepo.GetByPhone(req.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check phone"})
		return
	}
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Phone number already registered"})
		return
	}

	// Create user
	user := &model.User{
		Name:  req.Name,
		Email: req.Email,
		Phone: req.Phone,
		Role:  "customer",
	}

	if err := user.HashPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	if err := h.userRepo.Create(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate tokens
	tokens, err := h.jwtManager.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":   user,
		"tokens": tokens,
	})
}

// Login godoc
// @Summary Login user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} auth.TokenPair
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user
	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password
	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate tokens
	tokens, err := h.jwtManager.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":   user,
		"tokens": tokens,
	})
}

// RefreshToken godoc
// @Summary Refresh access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} map[string]string
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, err := h.jwtManager.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
	})
}

// GetProfile godoc
// @Summary Get current user profile
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} model.User
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	user, err := h.userRepo.GetByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GoogleLoginURL godoc
// @Summary Get Google OAuth login URL
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string
// @Router /auth/google/login [get]
func (h *AuthHandler) GoogleLoginURL(c *gin.Context) {
	// Generate random state for CSRF protection
	b := make([]byte, 32)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	// Store state in cookie
	c.SetCookie("oauth_state", state, 600, "/", "", false, true)

	// Build Google OAuth URL manually
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	redirectURI := os.Getenv("GOOGLE_REDIRECT_URL")

	params := url.Values{}
	params.Add("client_id", clientID)
	params.Add("redirect_uri", redirectURI)
	params.Add("response_type", "code")
	params.Add("scope", "https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile")
	params.Add("state", state)
	params.Add("access_type", "offline")

	authURL := fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?%s", params.Encode())

	c.JSON(http.StatusOK, gin.H{
		"url": authURL,
	})
}

// GoogleCallback godoc
// @Summary Handle Google OAuth callback
// @Tags auth
// @Produce json
// @Param state query string true "OAuth state"
// @Param code query string true "OAuth code"
// @Success 302 {string} string "Redirect to frontend"
// @Router /auth/google/callback [get]
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	// Verify state for CSRF protection
	state := c.Query("state")
	storedState, err := c.Cookie("oauth_state")
	if err != nil || state != storedState {
		c.Redirect(http.StatusTemporaryRedirect, os.Getenv("FRONTEND_URL")+"/login?error=invalid_state")
		return
	}

	// Exchange code for token
	code := c.Query("code")
	accessToken, err := h.exchangeCodeForToken(code)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, os.Getenv("FRONTEND_URL")+"/login?error=token_exchange_failed")
		return
	}

	// Get user info from Google
	googleUser, err := h.getGoogleUserInfo(accessToken)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, os.Getenv("FRONTEND_URL")+"/login?error=userinfo_failed")
		return
	}

	// Check if user already exists by Google ID
	user, err := h.userRepo.GetByGoogleID(googleUser.ID)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, os.Getenv("FRONTEND_URL")+"/login?error=db_error")
		return
	}

	// If user doesn't exist, check by email
	if user == nil {
		user, err = h.userRepo.GetByEmail(googleUser.Email)
		if err != nil {
			c.Redirect(http.StatusTemporaryRedirect, os.Getenv("FRONTEND_URL")+"/login?error=db_error")
			return
		}

		// If user exists with same email but no Google ID, link the account
		if user != nil {
			user.GoogleID = googleUser.ID
			user.Avatar = googleUser.Picture
			if err := h.userRepo.Update(user); err != nil {
				c.Redirect(http.StatusTemporaryRedirect, os.Getenv("FRONTEND_URL")+"/login?error=update_failed")
				return
			}
		}
	}

	// If user still doesn't exist, create new user
	if user == nil {
		user = &model.User{
			Name:     googleUser.Name,
			Email:    googleUser.Email,
			Phone:    "google_" + googleUser.ID, // Temporary phone
			GoogleID: googleUser.ID,
			Avatar:   googleUser.Picture,
			Role:     "customer",
		}

		// For OAuth users, set a random unguessable password hash
		if err := user.HashPassword("oauth_" + googleUser.ID + "_" + googleUser.Email); err != nil {
			c.Redirect(http.StatusTemporaryRedirect, os.Getenv("FRONTEND_URL")+"/login?error=hash_failed")
			return
		}

		if err := h.userRepo.Create(user); err != nil {
			c.Redirect(http.StatusTemporaryRedirect, os.Getenv("FRONTEND_URL")+"/login?error=create_failed")
			return
		}
	}

	// Generate JWT tokens
	tokens, err := h.jwtManager.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, os.Getenv("FRONTEND_URL")+"/login?error=token_failed")
		return
	}

	// Set JWT tokens in HTTP-only cookies with SameSite=None for cross-origin
	// Note: SameSite=None requires Secure=true (HTTPS only)
	c.Writer.Header().Add("Set-Cookie", fmt.Sprintf("access_token=%s; Path=/; Max-Age=3600; HttpOnly; Secure; SameSite=None", tokens.AccessToken))
	c.Writer.Header().Add("Set-Cookie", fmt.Sprintf("refresh_token=%s; Path=/; Max-Age=%d; HttpOnly; Secure; SameSite=None", tokens.RefreshToken, 86400*7))

	// Redirect to frontend
	c.Redirect(http.StatusTemporaryRedirect, os.Getenv("FRONTEND_URL")+"/?login=success")
}

// Helper function to exchange authorization code for access token
func (h *AuthHandler) exchangeCodeForToken(code string) (string, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURI := os.Getenv("GOOGLE_REDIRECT_URL")

	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", "authorization_code")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

// Helper function to get user info from Google
func (h *AuthHandler) getGoogleUserInfo(accessToken string) (*GoogleUserInfo, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}
