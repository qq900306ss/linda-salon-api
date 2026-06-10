package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/qq900306ss/linda-salon-api/internal/service"
)

// UploadHandler issues presigned S3 upload URLs.
type UploadHandler struct {
	s3Service *service.S3Service
}

// NewUploadHandler creates an UploadHandler.
func NewUploadHandler(s3Service *service.S3Service) *UploadHandler {
	return &UploadHandler{s3Service: s3Service}
}

type uploadRequest struct {
	FileName    string `json:"fileName" binding:"required"`
	ContentType string `json:"contentType" binding:"required"`
}

// CreateUpload handles POST /api/admin/uploads.
func (h *UploadHandler) CreateUpload(c *gin.Context) {
	var req uploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "INVALID_REQUEST", "fileName and contentType are required")
		return
	}
	uploadURL, publicURL, err := h.s3Service.PresignUpload(c.Request.Context(), req.FileName, req.ContentType)
	if err != nil {
		if errors.Is(err, service.ErrUploadNotConfigured) {
			Fail(c, http.StatusServiceUnavailable, "UPLOAD_NOT_CONFIGURED", "圖片上傳功能尚未設定 (缺少 S3_BUCKET)")
			return
		}
		Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create upload URL")
		return
	}
	OK(c, http.StatusOK, gin.H{"uploadUrl": uploadURL, "publicUrl": publicURL})
}
