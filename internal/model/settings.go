package model

import (
	"time"
)

// Settings 系統設定
type Settings struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Key       string    `json:"key" gorm:"uniqueIndex;not null"` // 設定鍵值
	Value     string    `json:"value" gorm:"type:text"`          // 設定值 (JSON string)
	Category  string    `json:"category"`                        // 設定分類 (pwa, branding, general)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PWA Icons Configuration
type PWAIconConfig struct {
	Icon72  string `json:"icon_72"`
	Icon96  string `json:"icon_96"`
	Icon128 string `json:"icon_128"`
	Icon144 string `json:"icon_144"`
	Icon152 string `json:"icon_152"`
	Icon192 string `json:"icon_192"`
	Icon384 string `json:"icon_384"`
	Icon512 string `json:"icon_512"`
}

// Branding Configuration
type BrandingConfig struct {
	Logo            string `json:"logo"`              // 主要 Logo URL
	LogoDark        string `json:"logo_dark"`         // 深色模式 Logo URL
	Favicon         string `json:"favicon"`           // Favicon URL
	Name            string `json:"name"`              // 品牌名稱
	ShortName       string `json:"short_name"`        // 簡短名稱
	Description     string `json:"description"`       // 品牌描述
	ThemeColor      string `json:"theme_color"`       // 主題顏色
	BackgroundColor string `json:"background_color"`  // 背景顏色
}

// PWA Configuration
type PWAConfig struct {
	Icons       PWAIconConfig `json:"icons"`
	Screenshots []string      `json:"screenshots"`
}

// 預設設定鍵值
const (
	SettingsKeyPWAIcons   = "pwa.icons"
	SettingsKeyBranding   = "branding"
	SettingsKeyScreenshots = "pwa.screenshots"
)
