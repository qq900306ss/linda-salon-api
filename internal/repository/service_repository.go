package repository

import (
	"errors"

	"gorm.io/gorm"
	"linda-salon-api/internal/model"
)

type ServiceRepository struct {
	db *gorm.DB
}

func NewServiceRepository(db *gorm.DB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

func (r *ServiceRepository) Create(service *model.Service) error {
	return r.db.Create(service).Error
}

func (r *ServiceRepository) GetByID(id uint) (*model.Service, error) {
	var service model.Service
	err := r.db.First(&service, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &service, nil
}

func (r *ServiceRepository) Update(service *model.Service) error {
	return r.db.Save(service).Error
}

func (r *ServiceRepository) Delete(id uint) error {
	return r.db.Delete(&model.Service{}, id).Error
}

func (r *ServiceRepository) List(category string, activeOnly bool) ([]model.Service, error) {
	var services []model.Service
	query := r.db.Model(&model.Service{})

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	err := query.Order("category, name").Find(&services).Error
	return services, err
}

func (r *ServiceRepository) GetByCategory(category string) ([]model.Service, error) {
	var services []model.Service
	err := r.db.Where("category = ? AND is_active = ?", category, true).
		Order("name").
		Find(&services).Error
	return services, err
}

func (r *ServiceRepository) GetPopular(limit int) ([]model.Service, error) {
	var services []model.Service
	err := r.db.
		Select("services.*, COUNT(bookings.id) as booking_count").
		Joins("LEFT JOIN bookings ON bookings.service_id = services.id").
		Where("services.is_active = ?", true).
		Group("services.id").
		Order("booking_count DESC").
		Limit(limit).
		Find(&services).Error
	return services, err
}
