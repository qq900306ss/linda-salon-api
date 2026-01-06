package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"linda-salon-api/internal/model"
)

type StylistRepository struct {
	db *gorm.DB
}

func NewStylistRepository(db *gorm.DB) *StylistRepository {
	return &StylistRepository{db: db}
}

func (r *StylistRepository) Create(stylist *model.Stylist) error {
	return r.db.Create(stylist).Error
}

func (r *StylistRepository) GetByID(id uint) (*model.Stylist, error) {
	var stylist model.Stylist
	err := r.db.Preload("Schedules").First(&stylist, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &stylist, nil
}

func (r *StylistRepository) Update(stylist *model.Stylist) error {
	return r.db.Save(stylist).Error
}

func (r *StylistRepository) Delete(id uint) error {
	return r.db.Delete(&model.Stylist{}, id).Error
}

func (r *StylistRepository) List(activeOnly bool) ([]model.Stylist, error) {
	var stylists []model.Stylist
	query := r.db.Preload("Schedules")

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	err := query.Order("name").Find(&stylists).Error
	return stylists, err
}

// Schedule management
func (r *StylistRepository) CreateSchedule(schedule *model.StylistSchedule) error {
	return r.db.Create(schedule).Error
}

func (r *StylistRepository) UpdateSchedule(schedule *model.StylistSchedule) error {
	return r.db.Save(schedule).Error
}

func (r *StylistRepository) DeleteSchedule(id uint) error {
	return r.db.Delete(&model.StylistSchedule{}, id).Error
}

func (r *StylistRepository) GetSchedulesByStylistID(stylistID uint) ([]model.StylistSchedule, error) {
	var schedules []model.StylistSchedule
	err := r.db.Where("stylist_id = ? AND is_active = ?", stylistID, true).
		Order("day_of_week, start_time").
		Find(&schedules).Error
	return schedules, err
}

// Check if stylist is available at given time
func (r *StylistRepository) IsAvailable(stylistID uint, date time.Time, startTime, endTime string) (bool, error) {
	dayOfWeek := int(date.Weekday())

	// Check if stylist has schedule for this day
	var schedule model.StylistSchedule
	err := r.db.Where("stylist_id = ? AND day_of_week = ? AND is_active = ?", stylistID, dayOfWeek, true).
		First(&schedule).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	// Check if requested time is within schedule
	if startTime < schedule.StartTime || endTime > schedule.EndTime {
		return false, nil
	}

	// Check for conflicting bookings
	var count int64
	err = r.db.Model(&model.Booking{}).
		Where("stylist_id = ? AND booking_date = ? AND status IN ?",
			stylistID, date.Format("2006-01-02"), []string{"pending", "confirmed"}).
		Where("NOT (end_time <= ? OR start_time >= ?)", startTime, endTime).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count == 0, nil
}

// Get top stylists by booking count
func (r *StylistRepository) GetTopStylists(limit int, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	err := r.db.Model(&model.Stylist{}).
		Select("stylists.id, stylists.name, COUNT(bookings.id) as booking_count, SUM(CASE WHEN bookings.status = ? THEN bookings.price ELSE 0 END) as revenue").
		Joins("LEFT JOIN bookings ON bookings.stylist_id = stylists.id AND bookings.status IN (?, ?, ?) AND bookings.booking_date BETWEEN ? AND ?",
			model.BookingStatusCompleted, model.BookingStatusPending, model.BookingStatusConfirmed, model.BookingStatusCompleted, startDate, endDate).
		Where("stylists.is_active = ?", true).
		Group("stylists.id, stylists.name").
		Order("booking_count DESC").
		Limit(limit).
		Find(&results).Error

	return results, err
}
