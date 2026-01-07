package repository

import (
	"gorm.io/gorm"
	"linda-salon-api/internal/model"
)

type SettingsRepository struct {
	db *gorm.DB
}

func NewSettingsRepository(db *gorm.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

// Get 取得設定
func (r *SettingsRepository) Get(key string) (*model.Settings, error) {
	var settings model.Settings
	err := r.db.Where("key = ?", key).First(&settings).Error
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

// GetAll 取得所有設定
func (r *SettingsRepository) GetAll() ([]model.Settings, error) {
	var settings []model.Settings
	err := r.db.Find(&settings).Error
	return settings, err
}

// GetByCategory 依分類取得設定
func (r *SettingsRepository) GetByCategory(category string) ([]model.Settings, error) {
	var settings []model.Settings
	err := r.db.Where("category = ?", category).Find(&settings).Error
	return settings, err
}

// Upsert 建立或更新設定
func (r *SettingsRepository) Upsert(settings *model.Settings) error {
	var existing model.Settings
	err := r.db.Where("key = ?", settings.Key).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// 建立新設定
		return r.db.Create(settings).Error
	}

	if err != nil {
		return err
	}

	// 更新現有設定
	settings.ID = existing.ID
	return r.db.Save(settings).Error
}

// Delete 刪除設定
func (r *SettingsRepository) Delete(key string) error {
	return r.db.Where("key = ?", key).Delete(&model.Settings{}).Error
}
