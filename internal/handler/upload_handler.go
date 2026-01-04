package handler

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"linda-salon-api/config"
)

type UploadHandler struct {
	s3Client *s3.Client
	cfg      *config.AWSConfig
}

func NewUploadHandler(s3Client *s3.Client, cfg *config.AWSConfig) *UploadHandler {
	return &UploadHandler{
		s3Client: s3Client,
		cfg:      cfg,
	}
}

// UploadImage godoc
// @Summary Upload an image to S3
// @Tags upload
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Image file"
// @Param folder query string false "Folder name (services, stylists, avatars)"
// @Success 200 {object} map[string]string
// @Router /upload/image [post]
func (h *UploadHandler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Validate file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only JPG, PNG, and WEBP are allowed"})
		return
	}

	// Validate file size (max 5MB)
	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 5MB limit"})
		return
	}

	// Get folder from query or default to 'uploads'
	folder := c.DefaultQuery("folder", "uploads")
	validFolders := map[string]bool{
		"services": true,
		"stylists": true,
		"avatars":  true,
		"uploads":  true,
	}
	if !validFolders[folder] {
		folder = "uploads"
	}

	// Generate unique filename
	uniqueID := uuid.New().String()
	filename := fmt.Sprintf("%s/%s%s", folder, uniqueID, ext)

	// Open file
	fileContent, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer fileContent.Close()

	// Determine content type
	contentType := "image/jpeg"
	if ext == ".png" {
		contentType = "image/png"
	} else if ext == ".webp" {
		contentType = "image/webp"
	}

	// Upload to S3
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = h.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(h.cfg.S3Bucket),
		Key:         aws.String(filename),
		Body:        fileContent,
		ContentType: aws.String(contentType),
		ACL:         "public-read",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to upload to S3",
			"details": err.Error(),
		})
		return
	}

	// Generate URL
	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
		h.cfg.S3Bucket,
		h.cfg.Region,
		filename)

	c.JSON(http.StatusOK, gin.H{
		"url":      url,
		"filename": filename,
		"folder":   folder,
	})
}

// DeleteImage godoc
// @Summary Delete an image from S3 (admin only)
// @Tags upload
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Image filename"
// @Success 200 {object} map[string]string
// @Router /upload/image [delete]
func (h *UploadHandler) DeleteImage(c *gin.Context) {
	var req struct {
		Filename string `json:"filename" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := h.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(h.cfg.S3Bucket),
		Key:    aws.String(req.Filename),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete from S3",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Image deleted successfully",
	})
}
