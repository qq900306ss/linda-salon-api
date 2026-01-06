package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"linda-salon-api/internal/model"
)

type BookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) Create(booking *model.Booking) error {
	return r.db.Create(booking).Error
}

func (r *BookingRepository) GetByID(id uint) (*model.Booking, error) {
	var booking model.Booking
	err := r.db.Preload("User").Preload("Stylist").First(&booking, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &booking, nil
}

func (r *BookingRepository) Update(booking *model.Booking) error {
	return r.db.Save(booking).Error
}

func (r *BookingRepository) Delete(id uint) error {
	return r.db.Delete(&model.Booking{}, id).Error
}

func (r *BookingRepository) List(userID *uint, status string, startDate, endDate *time.Time, limit, offset int) ([]model.Booking, int64, error) {
	var bookings []model.Booking
	var total int64

	query := r.db.Model(&model.Booking{}).Preload("User").Preload("Stylist")

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if startDate != nil {
		query = query.Where("booking_date >= ?", *startDate)
	}

	if endDate != nil {
		query = query.Where("booking_date <= ?", *endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("booking_date DESC, start_time DESC").
		Limit(limit).Offset(offset).
		Find(&bookings).Error

	return bookings, total, err
}

func (r *BookingRepository) GetUserBookings(userID uint, upcoming bool) ([]model.Booking, error) {
	var bookings []model.Booking
	query := r.db.Preload("Stylist").
		Where("user_id = ?", userID)

	if upcoming {
		query = query.Where("booking_date >= ? AND status IN ?",
			time.Now().Format("2006-01-02"),
			[]string{model.BookingStatusPending, model.BookingStatusConfirmed})
	}

	err := query.Order("booking_date ASC, start_time ASC").Find(&bookings).Error
	return bookings, err
}

func (r *BookingRepository) GetByDate(date time.Time) ([]model.Booking, error) {
	var bookings []model.Booking
	err := r.db.Preload("User").Preload("Stylist").
		Where("booking_date = ?", date.Format("2006-01-02")).
		Order("start_time").
		Find(&bookings).Error
	return bookings, err
}

func (r *BookingRepository) GetByStylistAndDate(stylistID uint, date time.Time) ([]model.Booking, error) {
	var bookings []model.Booking
	err := r.db.Preload("User").
		Where("stylist_id = ? AND booking_date = ? AND status IN ?",
			stylistID, date.Format("2006-01-02"),
			[]string{model.BookingStatusPending, model.BookingStatusConfirmed}).
		Order("start_time").
		Find(&bookings).Error
	return bookings, err
}

func (r *BookingRepository) GetByStylistAndDateString(stylistID uint, dateStr string) ([]model.Booking, error) {
	var bookings []model.Booking
	err := r.db.Preload("User").
		Where("stylist_id = ? AND booking_date = ? AND status IN ?",
			stylistID, dateStr,
			[]string{model.BookingStatusPending, model.BookingStatusConfirmed}).
		Order("start_time").
		Find(&bookings).Error
	return bookings, err
}

func (r *BookingRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.Booking{}).Where("id = ?", id).Update("status", status).Error
}

// Statistics queries
func (r *BookingRepository) CountByDateRange(startDate, endDate time.Time, status string) (int64, error) {
	var count int64
	query := r.db.Model(&model.Booking{}).
		Where("booking_date BETWEEN ? AND ?", startDate, endDate)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&count).Error
	return count, err
}

func (r *BookingRepository) GetRevenueByDateRange(startDate, endDate time.Time) (int, error) {
	var result struct {
		TotalRevenue int
	}
	err := r.db.Model(&model.Booking{}).
		Select("COALESCE(SUM(price), 0) as total_revenue").
		Where("booking_date BETWEEN ? AND ? AND status = ?",
			startDate, endDate, model.BookingStatusCompleted).
		Scan(&result).Error

	return result.TotalRevenue, err
}

func (r *BookingRepository) GetRevenueByDay(startDate, endDate time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := r.db.Model(&model.Booking{}).
		Select("booking_date as date, COUNT(*) as bookings, SUM(price) as revenue").
		Where("booking_date BETWEEN ? AND ? AND status = ?",
			startDate, endDate, model.BookingStatusCompleted).
		Group("booking_date").
		Order("booking_date ASC").
		Find(&results).Error

	return results, err
}

func (r *BookingRepository) GetPopularServices(limit int, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	// Since services are now stored as JSONB array in bookings, we need to:
	// 1. Extract service items from the JSONB array
	// 2. Count each service across all bookings
	var results []map[string]interface{}

	query := `
		SELECT
			service->>'name' as name,
			COUNT(*) as count
		FROM bookings,
		jsonb_array_elements(services) as service
		WHERE booking_date BETWEEN ? AND ?
		GROUP BY service->>'name'
		ORDER BY count DESC
		LIMIT ?
	`

	err := r.db.Raw(query, startDate, endDate, limit).Scan(&results).Error
	return results, err
}
