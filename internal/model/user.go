package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name         string `gorm:"type:varchar(100);not null" json:"name"`
	Email        string `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Phone        string `gorm:"type:varchar(20);uniqueIndex;not null" json:"phone"`
	PasswordHash string `gorm:"type:varchar(255);not null" json:"-"`
	Role         string `gorm:"type:varchar(20);not null;default:'customer'" json:"role"` // customer, admin
	Avatar       string `gorm:"type:varchar(500)" json:"avatar,omitempty"`

	// OAuth fields
	GoogleID string `gorm:"type:varchar(255);uniqueIndex" json:"google_id,omitempty"`
	LineID   string `gorm:"type:varchar(255);uniqueIndex" json:"line_id,omitempty"`

	// Relationships
	Bookings []Booking `gorm:"foreignKey:UserID" json:"bookings,omitempty"`
}

// HashPassword hashes the user's password
func (u *User) HashPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword verifies the password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// IsAdmin checks if user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}
