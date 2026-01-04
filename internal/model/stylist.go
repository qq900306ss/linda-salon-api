package model

import (
	"time"

	"gorm.io/gorm"
)

type Stylist struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string `gorm:"type:varchar(100);not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Specialty   string `gorm:"type:varchar(100)" json:"specialty"`
	Experience  int    `gorm:"default:0" json:"experience"` // years of experience
	Avatar      string `gorm:"type:varchar(500)" json:"avatar"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`

	// Relationships
	Schedules []StylistSchedule `gorm:"foreignKey:StylistID" json:"schedules,omitempty"`
	Bookings  []Booking         `gorm:"foreignKey:StylistID" json:"bookings,omitempty"`
}

type StylistSchedule struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	StylistID uint   `gorm:"not null;index" json:"stylist_id"`
	Stylist   Stylist `gorm:"foreignKey:StylistID" json:"stylist,omitempty"`

	DayOfWeek int    `gorm:"not null" json:"day_of_week"` // 0=Sunday, 1=Monday, ..., 6=Saturday
	StartTime string `gorm:"type:varchar(5);not null" json:"start_time"` // HH:MM format
	EndTime   string `gorm:"type:varchar(5);not null" json:"end_time"`   // HH:MM format
	IsActive  bool   `gorm:"default:true" json:"is_active"`
}
