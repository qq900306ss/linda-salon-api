package model

// Schedule describes a stylist's recurring working schedule.
type Schedule struct {
	// WorkDays uses Go's time.Weekday convention: 0=Sunday ... 6=Saturday.
	WorkDays  []int    `json:"workDays" dynamodbav:"workDays"`
	StartTime string   `json:"startTime" dynamodbav:"startTime"` // "10:00"
	EndTime   string   `json:"endTime" dynamodbav:"endTime"`     // "19:00"
	DaysOff   []string `json:"daysOff" dynamodbav:"daysOff"`     // ["YYYY-MM-DD"]
}

// Normalize replaces nil slices with empty slices so JSON output stays as arrays.
func (s *Schedule) Normalize() {
	if s.WorkDays == nil {
		s.WorkDays = []int{}
	}
	if s.DaysOff == nil {
		s.DaysOff = []string{}
	}
}

// Stylist represents a hair stylist.
type Stylist struct {
	ID              string   `json:"id" dynamodbav:"id"`
	Name            string   `json:"name" dynamodbav:"name"`
	Title           string   `json:"title" dynamodbav:"title"`
	Bio             string   `json:"bio" dynamodbav:"bio"`
	Specialties     []string `json:"specialties" dynamodbav:"specialties"`
	ImageURL        string   `json:"imageUrl" dynamodbav:"imageUrl"`
	YearsExperience int      `json:"yearsExperience" dynamodbav:"yearsExperience"`
	Rating          float64  `json:"rating" dynamodbav:"rating"`
	IsActive        bool     `json:"isActive" dynamodbav:"isActive"`
	Schedule        Schedule `json:"schedule" dynamodbav:"schedule"`
}

// Normalize replaces nil slices with empty slices so JSON output stays as arrays.
func (s *Stylist) Normalize() {
	if s.Specialties == nil {
		s.Specialties = []string{}
	}
	s.Schedule.Normalize()
}
