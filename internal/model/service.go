package model

import (
	"time"

	"gorm.io/gorm"
)

type Service struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string `gorm:"type:varchar(100);not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Category    string `gorm:"type:varchar(50);not null" json:"category"` // haircut, coloring, treatment, styling, perm
	Price       int    `gorm:"not null" json:"price"`
	Duration    int    `gorm:"not null" json:"duration"` // in minutes
	ImageURL    string `gorm:"type:varchar(500)" json:"image_url"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`

	// Relationships
	// Note: Removed Bookings relationship since services are now stored as JSONB in bookings table
	// Bookings []Booking `gorm:"foreignKey:ServiceID" json:"bookings,omitempty"`
}
