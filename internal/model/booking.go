package model

// Booking statuses.
const (
	BookingStatusPending   = "pending"
	BookingStatusConfirmed = "confirmed"
	BookingStatusCompleted = "completed"
	BookingStatusCancelled = "cancelled"
)

// ValidBookingStatus reports whether s is a known booking status.
func ValidBookingStatus(s string) bool {
	switch s {
	case BookingStatusPending, BookingStatusConfirmed, BookingStatusCompleted, BookingStatusCancelled:
		return true
	}
	return false
}

// Customer holds the customer contact info embedded in a booking.
type Customer struct {
	Name  string `json:"name" dynamodbav:"name"`
	Phone string `json:"phone" dynamodbav:"phone"`
	Email string `json:"email" dynamodbav:"email"`
	Notes string `json:"notes" dynamodbav:"notes"`
}

// Booking represents a customer booking. Service/stylist info is denormalized
// at creation time so historical records survive later edits.
type Booking struct {
	ID              string   `json:"id" dynamodbav:"id"`
	ServiceID       string   `json:"serviceId" dynamodbav:"serviceId"`
	ServiceName     string   `json:"serviceName" dynamodbav:"serviceName"`
	StylistID       string   `json:"stylistId" dynamodbav:"stylistId"`
	StylistName     string   `json:"stylistName" dynamodbav:"stylistName"`
	Date            string   `json:"date" dynamodbav:"date"` // "YYYY-MM-DD"
	Time            string   `json:"time" dynamodbav:"time"` // "HH:MM"
	DurationMinutes int      `json:"durationMinutes" dynamodbav:"durationMinutes"`
	Price           int      `json:"price" dynamodbav:"price"`
	Status          string   `json:"status" dynamodbav:"status"`
	Customer        Customer `json:"customer" dynamodbav:"customer"`
	// Phone duplicates Customer.Phone as a top-level attribute for the phone-index GSI.
	Phone     string `json:"-" dynamodbav:"phone"`
	CreatedAt string `json:"createdAt" dynamodbav:"createdAt"` // RFC3339
	UpdatedAt string `json:"updatedAt" dynamodbav:"updatedAt"` // RFC3339
}

// CountsAsRevenue reports whether this booking should be counted as revenue.
func (b *Booking) CountsAsRevenue() bool {
	return b.Status == BookingStatusConfirmed || b.Status == BookingStatusCompleted
}
