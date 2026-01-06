package model

import (
	"time"

	"gorm.io/gorm"
)

// BookingServiceItem represents a service item in a booking (stored in JSONB)
type BookingServiceItem struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Duration int    `json:"duration"`
}

type Booking struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Foreign Keys
	UserID    uint `gorm:"not null;index" json:"user_id"`
	StylistID uint `gorm:"not null;index" json:"stylist_id"`

	// Relationships
	User    User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Stylist Stylist `gorm:"foreignKey:StylistID" json:"stylist,omitempty"`

	// Multiple Services (JSONB) - replaces service_id
	Services []BookingServiceItem `gorm:"type:jsonb;serializer:json;not null" json:"services"`

	// Booking Details
	BookingDate time.Time `gorm:"not null;index" json:"booking_date"`
	StartTime   string    `gorm:"type:varchar(5);not null" json:"start_time"` // HH:MM
	EndTime     string    `gorm:"type:varchar(5);not null" json:"end_time"`   // HH:MM
	Duration    int       `gorm:"not null" json:"duration"` // minutes
	Price       int       `gorm:"not null" json:"price"`
	Status      string    `gorm:"type:varchar(20);not null;default:'pending'" json:"status"` // pending, confirmed, completed, cancelled
	Notes       string    `gorm:"type:text" json:"notes"`

	// Customer Info (denormalized for easier queries)
	CustomerName  string `gorm:"type:varchar(100);not null" json:"customer_name"`
	CustomerPhone string `gorm:"type:varchar(20);not null" json:"customer_phone"`
	CustomerEmail string `gorm:"type:varchar(255)" json:"customer_email"`
}

// BookingStatus constants
const (
	BookingStatusPending   = "pending"
	BookingStatusConfirmed = "confirmed"
	BookingStatusCompleted = "completed"
	BookingStatusCancelled = "cancelled"
)

// IsCancellable checks if booking can be cancelled
func (b *Booking) IsCancellable() bool {
	return b.Status == BookingStatusPending || b.Status == BookingStatusConfirmed
}

// IsUpcoming checks if booking is in the future
func (b *Booking) IsUpcoming() bool {
	return b.BookingDate.After(time.Now()) &&
		(b.Status == BookingStatusPending || b.Status == BookingStatusConfirmed)
}
