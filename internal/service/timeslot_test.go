package service

import (
	"testing"
	"time"

	"github.com/qq900306ss/linda-salon-api/internal/model"
)

func testSettings() model.Settings {
	return model.Settings{
		OpenTime:            "10:00",
		CloseTime:           "13:00",
		SlotIntervalMinutes: 30,
		ClosedWeekdays:      []int{1}, // Monday closed
	}
}

func testStylist() model.Stylist {
	return model.Stylist{
		ID:       "sty-1",
		Name:     "Linda",
		IsActive: true,
		Schedule: model.Schedule{
			WorkDays:  []int{0, 2, 3, 4, 5, 6}, // off on Monday
			StartTime: "10:00",
			EndTime:   "13:00",
			DaysOff:   []string{},
		},
	}
}

// futureNow returns a fixed "now" long before the test dates.
func futureNow() time.Time {
	return time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
}

func slotByTime(t *testing.T, slots []TimeSlot, hm string) TimeSlot {
	t.Helper()
	for _, s := range slots {
		if s.Time == hm {
			return s
		}
	}
	t.Fatalf("slot %s not found", hm)
	return TimeSlot{}
}

func TestOverlaps(t *testing.T) {
	cases := []struct {
		name                       string
		aStart, aDur, bStart, bDur int
		want                       bool
	}{
		{"identical intervals", 600, 60, 600, 60, true},
		{"a ends exactly when b starts", 600, 60, 660, 60, false},
		{"b ends exactly when a starts", 660, 60, 600, 60, false},
		{"a inside b", 630, 30, 600, 120, true},
		{"b inside a", 600, 120, 630, 30, true},
		{"partial overlap front", 570, 60, 600, 60, true},
		{"partial overlap back", 630, 60, 600, 60, true},
		{"disjoint", 600, 30, 700, 30, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Overlaps(tc.aStart, tc.aDur, tc.bStart, tc.bDur); got != tc.want {
				t.Errorf("Overlaps(%d,%d,%d,%d) = %v, want %v",
					tc.aStart, tc.aDur, tc.bStart, tc.bDur, got, tc.want)
			}
		})
	}
}

func TestParseHM(t *testing.T) {
	if v, err := ParseHM("10:30"); err != nil || v != 630 {
		t.Errorf("ParseHM(10:30) = %d, %v; want 630, nil", v, err)
	}
	if _, err := ParseHM("25:00"); err == nil {
		t.Error("ParseHM(25:00) should fail")
	}
	if _, err := ParseHM("abc"); err == nil {
		t.Error("ParseHM(abc) should fail")
	}
	if FormatHM(630) != "10:30" {
		t.Errorf("FormatHM(630) = %s, want 10:30", FormatHM(630))
	}
}

func TestGenerateTimeSlots_AllAvailable(t *testing.T) {
	// 2026-01-07 is a Wednesday.
	slots, err := GenerateTimeSlots(SlotQuery{
		Settings:        testSettings(),
		Stylist:         testStylist(),
		Date:            "2026-01-07",
		DurationMinutes: 60,
		Now:             futureNow(),
	})
	if err != nil {
		t.Fatal(err)
	}
	// 10:00–13:00 with 30-minute interval → 6 slots.
	if len(slots) != 6 {
		t.Fatalf("expected 6 slots, got %d", len(slots))
	}
	// 10:00 .. 12:00 fit a 60-minute service; 12:30 does not (ends 13:30).
	for _, hm := range []string{"10:00", "10:30", "11:00", "11:30", "12:00"} {
		if s := slotByTime(t, slots, hm); !s.Available {
			t.Errorf("slot %s should be available, got reason %q", hm, s.Reason)
		}
	}
	if s := slotByTime(t, slots, "12:30"); s.Available || s.Reason != ReasonOutsideHours {
		t.Errorf("slot 12:30 should be outsideHours, got %+v", s)
	}
}

func TestGenerateTimeSlots_BookingOverlap(t *testing.T) {
	bookings := []model.Booking{
		{StylistID: "sty-1", Date: "2026-01-07", Time: "11:00", DurationMinutes: 60, Status: model.BookingStatusConfirmed},
	}
	slots, err := GenerateTimeSlots(SlotQuery{
		Settings:        testSettings(),
		Stylist:         testStylist(),
		Date:            "2026-01-07",
		DurationMinutes: 60,
		Bookings:        bookings,
		Now:             futureNow(),
	})
	if err != nil {
		t.Fatal(err)
	}
	// Booking occupies 11:00–12:00. A 60-minute service starting at
	// 10:30, 11:00 or 11:30 overlaps it; 10:00 and 12:00 do not.
	for _, hm := range []string{"10:30", "11:00", "11:30"} {
		if s := slotByTime(t, slots, hm); s.Available || s.Reason != ReasonBooked {
			t.Errorf("slot %s should be booked, got %+v", hm, s)
		}
	}
	for _, hm := range []string{"10:00", "12:00"} {
		if s := slotByTime(t, slots, hm); !s.Available {
			t.Errorf("slot %s should be available, got reason %q", hm, s.Reason)
		}
	}
}

func TestGenerateTimeSlots_CancelledBookingIgnored(t *testing.T) {
	bookings := []model.Booking{
		{StylistID: "sty-1", Date: "2026-01-07", Time: "11:00", DurationMinutes: 60, Status: model.BookingStatusCancelled},
	}
	slots, err := GenerateTimeSlots(SlotQuery{
		Settings:        testSettings(),
		Stylist:         testStylist(),
		Date:            "2026-01-07",
		DurationMinutes: 60,
		Bookings:        bookings,
		Now:             futureNow(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if s := slotByTime(t, slots, "11:00"); !s.Available {
		t.Errorf("slot 11:00 should be available (booking cancelled), got reason %q", s.Reason)
	}
}

func TestGenerateTimeSlots_SalonClosedWeekday(t *testing.T) {
	// 2026-01-05 is a Monday — salon closed.
	slots, err := GenerateTimeSlots(SlotQuery{
		Settings:        testSettings(),
		Stylist:         testStylist(),
		Date:            "2026-01-05",
		DurationMinutes: 60,
		Now:             futureNow(),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range slots {
		if s.Available || s.Reason != ReasonClosed {
			t.Errorf("slot %s should be closed, got %+v", s.Time, s)
		}
	}
}

func TestGenerateTimeSlots_StylistDayOff(t *testing.T) {
	stylist := testStylist()
	stylist.Schedule.DaysOff = []string{"2026-01-07"}
	slots, err := GenerateTimeSlots(SlotQuery{
		Settings:        testSettings(),
		Stylist:         stylist,
		Date:            "2026-01-07",
		DurationMinutes: 60,
		Now:             futureNow(),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range slots {
		if s.Available || s.Reason != ReasonDayOff {
			t.Errorf("slot %s should be dayOff, got %+v", s.Time, s)
		}
	}
}

func TestGenerateTimeSlots_StylistNonWorkDay(t *testing.T) {
	stylist := testStylist()
	stylist.Schedule.WorkDays = []int{2, 3} // Tue, Wed only
	// 2026-01-08 is a Thursday.
	slots, err := GenerateTimeSlots(SlotQuery{
		Settings:        testSettings(),
		Stylist:         stylist,
		Date:            "2026-01-08",
		DurationMinutes: 60,
		Now:             futureNow(),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range slots {
		if s.Available || s.Reason != ReasonDayOff {
			t.Errorf("slot %s should be dayOff, got %+v", s.Time, s)
		}
	}
}

func TestGenerateTimeSlots_PastSlots(t *testing.T) {
	// "now" is 11:10 on the queried date.
	now := time.Date(2026, 1, 7, 11, 10, 0, 0, time.UTC)
	slots, err := GenerateTimeSlots(SlotQuery{
		Settings:        testSettings(),
		Stylist:         testStylist(),
		Date:            "2026-01-07",
		DurationMinutes: 30,
		Now:             now,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, hm := range []string{"10:00", "10:30", "11:00"} {
		if s := slotByTime(t, slots, hm); s.Available || s.Reason != ReasonPast {
			t.Errorf("slot %s should be past, got %+v", hm, s)
		}
	}
	if s := slotByTime(t, slots, "11:30"); !s.Available {
		t.Errorf("slot 11:30 should be available, got reason %q", s.Reason)
	}
}

func TestGenerateTimeSlots_EntirePastDate(t *testing.T) {
	now := time.Date(2026, 1, 10, 9, 0, 0, 0, time.UTC)
	slots, err := GenerateTimeSlots(SlotQuery{
		Settings:        testSettings(),
		Stylist:         testStylist(),
		Date:            "2026-01-07",
		DurationMinutes: 30,
		Now:             now,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range slots {
		if s.Available || s.Reason != ReasonPast {
			t.Errorf("slot %s should be past, got %+v", s.Time, s)
		}
	}
}

func TestGenerateTimeSlots_StylistHoursClamp(t *testing.T) {
	stylist := testStylist()
	stylist.Schedule.StartTime = "11:00"
	stylist.Schedule.EndTime = "12:30"
	slots, err := GenerateTimeSlots(SlotQuery{
		Settings:        testSettings(),
		Stylist:         stylist,
		Date:            "2026-01-07",
		DurationMinutes: 30,
		Now:             futureNow(),
	})
	if err != nil {
		t.Fatal(err)
	}
	// Before 11:00 → outsideHours; 12:30 slot ends 13:00 > stylist end 12:30.
	for _, hm := range []string{"10:00", "10:30", "12:30"} {
		if s := slotByTime(t, slots, hm); s.Available || s.Reason != ReasonOutsideHours {
			t.Errorf("slot %s should be outsideHours, got %+v", hm, s)
		}
	}
	for _, hm := range []string{"11:00", "11:30", "12:00"} {
		if s := slotByTime(t, slots, hm); !s.Available {
			t.Errorf("slot %s should be available, got reason %q", hm, s.Reason)
		}
	}
}

func TestGenerateTimeSlots_InvalidDate(t *testing.T) {
	_, err := GenerateTimeSlots(SlotQuery{
		Settings: testSettings(),
		Stylist:  testStylist(),
		Date:     "not-a-date",
		Now:      futureNow(),
	})
	if err == nil {
		t.Error("expected error for invalid date")
	}
}
