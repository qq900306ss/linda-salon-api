package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"linda-salon-api/internal/model"
	"linda-salon-api/internal/repository"
)

type SettingsHandler struct {
	settingsRepo *repository.SettingsRepository
}

func NewSettingsHandler(settingsRepo *repository.SettingsRepository) *SettingsHandler {
	return &SettingsHandler{
		settingsRepo: settingsRepo,
	}
}

// GetPWAIcons 取得 PWA 圖標設定
// GET /api/v1/settings/pwa/icons
func (h *SettingsHandler) GetPWAIcons(c *gin.Context) {
	settings, err := h.settingsRepo.Get(model.SettingsKeyPWAIcons)
	if err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get PWA icons"})
		return
	}

	if err == gorm.ErrRecordNotFound {
		// 返回預設值
		c.JSON(http.StatusOK, model.PWAIconConfig{})
		return
	}

	var config model.PWAIconConfig
	if err := json.Unmarshal([]byte(settings.Value), &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse PWA icons"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdatePWAIcons 更新 PWA 圖標設定 (Admin only)
// PUT /api/v1/admin/settings/pwa/icons
func (h *SettingsHandler) UpdatePWAIcons(c *gin.Context) {
	var config model.PWAIconConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	value, err := json.Marshal(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize config"})
		return
	}

	settings := &model.Settings{
		Key:      model.SettingsKeyPWAIcons,
		Value:    string(value),
		Category: "pwa",
	}

	if err := h.settingsRepo.Upsert(settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save PWA icons"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// GetBranding 取得品牌設定
// GET /api/v1/settings/branding
func (h *SettingsHandler) GetBranding(c *gin.Context) {
	settings, err := h.settingsRepo.Get(model.SettingsKeyBranding)
	if err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get branding"})
		return
	}

	if err == gorm.ErrRecordNotFound {
		// 返回預設值
		c.JSON(http.StatusOK, model.BrandingConfig{
			Name:            "Linda 髮廊",
			ShortName:       "Linda",
			Description:     "專業美髮服務，打造您的完美造型",
			ThemeColor:      "#8B5CF6",
			BackgroundColor: "#FFFFFF",
		})
		return
	}

	var config model.BrandingConfig
	if err := json.Unmarshal([]byte(settings.Value), &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse branding"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateBranding 更新品牌設定 (Admin only)
// PUT /api/v1/admin/settings/branding
func (h *SettingsHandler) UpdateBranding(c *gin.Context) {
	var config model.BrandingConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	value, err := json.Marshal(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize config"})
		return
	}

	settings := &model.Settings{
		Key:      model.SettingsKeyBranding,
		Value:    string(value),
		Category: "branding",
	}

	if err := h.settingsRepo.Upsert(settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save branding"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// GetManifest 取得 PWA manifest.json
// GET /api/v1/manifest.json
func (h *SettingsHandler) GetManifest(c *gin.Context) {
	// 取得品牌設定
	branding, err := h.settingsRepo.Get(model.SettingsKeyBranding)
	var brandingConfig model.BrandingConfig
	if err == nil {
		json.Unmarshal([]byte(branding.Value), &brandingConfig)
	} else {
		// 使用預設值
		brandingConfig = model.BrandingConfig{
			Name:            "Linda 髮廊",
			ShortName:       "Linda",
			Description:     "專業美髮服務，打造您的完美造型",
			ThemeColor:      "#8B5CF6",
			BackgroundColor: "#FFFFFF",
		}
	}

	// 取得圖標設定
	icons, err := h.settingsRepo.Get(model.SettingsKeyPWAIcons)
	var iconConfig model.PWAIconConfig
	if err == nil {
		json.Unmarshal([]byte(icons.Value), &iconConfig)
	}

	// 構建 manifest
	manifest := map[string]interface{}{
		"name":             brandingConfig.Name,
		"short_name":       brandingConfig.ShortName,
		"description":      brandingConfig.Description,
		"start_url":        "/",
		"display":          "standalone",
		"background_color": brandingConfig.BackgroundColor,
		"theme_color":      brandingConfig.ThemeColor,
		"orientation":      "portrait-primary",
		"categories":       []string{"lifestyle", "business"},
		"lang":             "zh-TW",
		"dir":              "ltr",
	}

	// 添加圖標
	manifestIcons := []map[string]interface{}{}
	if iconConfig.Icon72 != "" {
		manifestIcons = append(manifestIcons, map[string]interface{}{
			"src":     iconConfig.Icon72,
			"sizes":   "72x72",
			"type":    "image/png",
			"purpose": "any maskable",
		})
	}
	if iconConfig.Icon96 != "" {
		manifestIcons = append(manifestIcons, map[string]interface{}{
			"src":     iconConfig.Icon96,
			"sizes":   "96x96",
			"type":    "image/png",
			"purpose": "any maskable",
		})
	}
	if iconConfig.Icon128 != "" {
		manifestIcons = append(manifestIcons, map[string]interface{}{
			"src":     iconConfig.Icon128,
			"sizes":   "128x128",
			"type":    "image/png",
			"purpose": "any maskable",
		})
	}
	if iconConfig.Icon144 != "" {
		manifestIcons = append(manifestIcons, map[string]interface{}{
			"src":     iconConfig.Icon144,
			"sizes":   "144x144",
			"type":    "image/png",
			"purpose": "any maskable",
		})
	}
	if iconConfig.Icon152 != "" {
		manifestIcons = append(manifestIcons, map[string]interface{}{
			"src":     iconConfig.Icon152,
			"sizes":   "152x152",
			"type":    "image/png",
			"purpose": "any maskable",
		})
	}
	if iconConfig.Icon192 != "" {
		manifestIcons = append(manifestIcons, map[string]interface{}{
			"src":     iconConfig.Icon192,
			"sizes":   "192x192",
			"type":    "image/png",
			"purpose": "any maskable",
		})
	}
	if iconConfig.Icon384 != "" {
		manifestIcons = append(manifestIcons, map[string]interface{}{
			"src":     iconConfig.Icon384,
			"sizes":   "384x384",
			"type":    "image/png",
			"purpose": "any maskable",
		})
	}
	if iconConfig.Icon512 != "" {
		manifestIcons = append(manifestIcons, map[string]interface{}{
			"src":     iconConfig.Icon512,
			"sizes":   "512x512",
			"type":    "image/png",
			"purpose": "any maskable",
		})
	}

	if len(manifestIcons) > 0 {
		manifest["icons"] = manifestIcons
	}

	c.Header("Content-Type", "application/manifest+json")
	c.JSON(http.StatusOK, manifest)
}
