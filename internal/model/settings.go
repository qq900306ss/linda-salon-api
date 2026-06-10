package model

// SettingsID is the fixed primary key of the single settings item.
const SettingsID = "global"

// Settings holds salon-wide configuration (single item, id="global").
type Settings struct {
	ID                  string `json:"-" dynamodbav:"id"`
	SalonName           string `json:"salonName" dynamodbav:"salonName"`
	Phone               string `json:"phone" dynamodbav:"phone"`
	Address             string `json:"address" dynamodbav:"address"`
	OpenTime            string `json:"openTime" dynamodbav:"openTime"`   // "10:00"
	CloseTime           string `json:"closeTime" dynamodbav:"closeTime"` // "19:00"
	SlotIntervalMinutes int    `json:"slotIntervalMinutes" dynamodbav:"slotIntervalMinutes"`
	// ClosedWeekdays uses 0=Sunday ... 6=Saturday.
	ClosedWeekdays []int `json:"closedWeekdays" dynamodbav:"closedWeekdays"`
}

// Normalize replaces nil slices with empty slices so JSON output stays as arrays.
func (s *Settings) Normalize() {
	if s.ClosedWeekdays == nil {
		s.ClosedWeekdays = []int{}
	}
}
