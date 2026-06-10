// Package service contains the pure business logic of the booking system.
package service

import (
	"fmt"
	"time"

	"github.com/qq900306ss/linda-salon-api/internal/model"
)

// Slot unavailability reasons.
const (
	ReasonBooked       = "booked"
	ReasonClosed       = "closed"
	ReasonDayOff       = "dayOff"
	ReasonOutsideHours = "outsideHours"
	ReasonPast         = "past"
)

// TimeSlot is a single bookable slot in the timeslot response.
type TimeSlot struct {
	Time      string `json:"time"`
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}

// SlotQuery is the input to GenerateTimeSlots.
type SlotQuery struct {
	Settings        model.Settings
	Stylist         model.Stylist
	Date            string // "YYYY-MM-DD"
	DurationMinutes int    // duration of the requested service
	// Bookings should contain the stylist's non-cancelled bookings on Date.
	Bookings []model.Booking
	// Now is the current time used for past-slot detection (salon timezone).
	Now time.Time
}

// TaipeiLocation returns the Asia/Taipei location, falling back to a fixed
// UTC+8 zone if the tz database is unavailable.
func TaipeiLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		return time.FixedZone("UTC+8", 8*60*60)
	}
	return loc
}

// ParseHM converts "HH:MM" to minutes since midnight.
func ParseHM(s string) (int, error) {
	var h, m int
	if _, err := fmt.Sscanf(s, "%d:%d", &h, &m); err != nil {
		return 0, fmt.Errorf("invalid time %q", s)
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, fmt.Errorf("invalid time %q", s)
	}
	return h*60 + m, nil
}

// FormatHM converts minutes since midnight to "HH:MM".
func FormatHM(minutes int) string {
	return fmt.Sprintf("%02d:%02d", minutes/60, minutes%60)
}

// Overlaps reports whether two time intervals [aStart, aStart+aDur) and
// [bStart, bStart+bDur) (in minutes since midnight) overlap.
func Overlaps(aStart, aDur, bStart, bDur int) bool {
	return aStart < bStart+bDur && bStart < aStart+aDur
}

func containsInt(list []int, v int) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func containsString(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// GenerateTimeSlots produces the list of slots for a stylist on a date.
// A slot is unavailable when (in priority order): the salon is closed that
// weekday, the stylist does not work that day (or has a day off), the slot
// does not fit inside salon and stylist working hours, the slot start is in
// the past, or it overlaps a non-cancelled booking.
func GenerateTimeSlots(q SlotQuery) ([]TimeSlot, error) {
	date, err := time.ParseInLocation("2006-01-02", q.Date, q.Now.Location())
	if err != nil {
		return nil, fmt.Errorf("invalid date %q", q.Date)
	}

	openMin, err := ParseHM(q.Settings.OpenTime)
	if err != nil {
		return nil, err
	}
	closeMin, err := ParseHM(q.Settings.CloseTime)
	if err != nil {
		return nil, err
	}
	interval := q.Settings.SlotIntervalMinutes
	if interval <= 0 {
		interval = 30
	}
	duration := q.DurationMinutes
	if duration <= 0 {
		duration = interval
	}

	salonClosed := containsInt(q.Settings.ClosedWeekdays, int(date.Weekday()))

	stylistOff := !containsInt(q.Stylist.Schedule.WorkDays, int(date.Weekday())) ||
		containsString(q.Stylist.Schedule.DaysOff, q.Date)

	// Stylist working window, clamped to salon hours.
	workStart, workEnd := openMin, closeMin
	if q.Stylist.Schedule.StartTime != "" {
		if v, err := ParseHM(q.Stylist.Schedule.StartTime); err == nil && v > workStart {
			workStart = v
		}
	}
	if q.Stylist.Schedule.EndTime != "" {
		if v, err := ParseHM(q.Stylist.Schedule.EndTime); err == nil && v < workEnd {
			workEnd = v
		}
	}

	// Minutes since midnight of "now" if the queried date is today; -1 means
	// the whole day is in the future, a huge value means it is entirely past.
	nowMin := -1
	today := q.Now.Format("2006-01-02")
	if q.Date == today {
		nowMin = q.Now.Hour()*60 + q.Now.Minute()
	} else if q.Date < today {
		nowMin = 24 * 60
	}

	var slots []TimeSlot
	for t := openMin; t < closeMin; t += interval {
		slot := TimeSlot{Time: FormatHM(t)}
		switch {
		case salonClosed:
			slot.Reason = ReasonClosed
		case stylistOff:
			slot.Reason = ReasonDayOff
		case t < workStart || t+duration > workEnd:
			slot.Reason = ReasonOutsideHours
		case t <= nowMin:
			slot.Reason = ReasonPast
		default:
			for _, b := range q.Bookings {
				if b.Status == model.BookingStatusCancelled {
					continue
				}
				bStart, err := ParseHM(b.Time)
				if err != nil {
					continue
				}
				if Overlaps(t, duration, bStart, b.DurationMinutes) {
					slot.Reason = ReasonBooked
					break
				}
			}
		}
		slot.Available = slot.Reason == ""
		slots = append(slots, slot)
	}
	return slots, nil
}
