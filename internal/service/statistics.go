package service

import (
	"sort"
	"time"

	"github.com/qq900306ss/linda-salon-api/internal/model"
)

// PeriodStats summarizes bookings and revenue for a period.
type PeriodStats struct {
	Bookings int `json:"bookings"`
	Revenue  int `json:"revenue"`
}

// ServiceStat aggregates bookings per service.
type ServiceStat struct {
	ServiceID   string `json:"serviceId"`
	ServiceName string `json:"serviceName"`
	Count       int    `json:"count"`
	Revenue     int    `json:"revenue"`
}

// StylistStat aggregates bookings per stylist.
type StylistStat struct {
	StylistID   string `json:"stylistId"`
	StylistName string `json:"stylistName"`
	Count       int    `json:"count"`
	Revenue     int    `json:"revenue"`
}

// DailyRevenue is one day's revenue summary.
type DailyRevenue struct {
	Date     string `json:"date"`
	Revenue  int    `json:"revenue"`
	Bookings int    `json:"bookings"`
}

// DashboardStats is the admin dashboard payload.
type DashboardStats struct {
	Today           PeriodStats     `json:"today"`
	Week            PeriodStats     `json:"week"`
	Month           PeriodStats     `json:"month"`
	PendingCount    int             `json:"pendingCount"`
	PopularServices []ServiceStat   `json:"popularServices"`
	TopStylists     []StylistStat   `json:"topStylists"`
	RecentBookings  []model.Booking `json:"recentBookings"`
	DailyRevenue    []DailyRevenue  `json:"dailyRevenue"`
}

// CustomerSummary aggregates a customer's bookings by phone.
type CustomerSummary struct {
	Phone         string `json:"phone"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	TotalBookings int    `json:"totalBookings"`
	TotalSpent    int    `json:"totalSpent"`
	LastVisit     string `json:"lastVisit"`
	FirstVisit    string `json:"firstVisit"`
}

// BuildDashboardStats computes dashboard statistics from all bookings.
// now must be in the salon timezone. Revenue counts confirmed + completed.
func BuildDashboardStats(bookings []model.Booking, now time.Time) DashboardStats {
	today := now.Format("2006-01-02")
	weekStart := now.AddDate(0, 0, -6).Format("2006-01-02")
	monthStart := now.AddDate(0, 0, -29).Format("2006-01-02")

	stats := DashboardStats{
		PopularServices: []ServiceStat{},
		TopStylists:     []StylistStat{},
		RecentBookings:  []model.Booking{},
		DailyRevenue:    []DailyRevenue{},
	}

	serviceAgg := map[string]*ServiceStat{}
	stylistAgg := map[string]*StylistStat{}
	dailyAgg := map[string]*DailyRevenue{}

	for i := range bookings {
		b := &bookings[i]
		if b.Status == model.BookingStatusPending {
			stats.PendingCount++
		}
		revenue := 0
		if b.CountsAsRevenue() {
			revenue = b.Price
		}

		if b.Date == today {
			stats.Today.Bookings++
			stats.Today.Revenue += revenue
		}
		if b.Date >= weekStart && b.Date <= today {
			stats.Week.Bookings++
			stats.Week.Revenue += revenue
		}
		if b.Date >= monthStart && b.Date <= today {
			stats.Month.Bookings++
			stats.Month.Revenue += revenue

			ss := serviceAgg[b.ServiceID]
			if ss == nil {
				ss = &ServiceStat{ServiceID: b.ServiceID, ServiceName: b.ServiceName}
				serviceAgg[b.ServiceID] = ss
			}
			ss.Count++
			ss.Revenue += revenue

			ts := stylistAgg[b.StylistID]
			if ts == nil {
				ts = &StylistStat{StylistID: b.StylistID, StylistName: b.StylistName}
				stylistAgg[b.StylistID] = ts
			}
			ts.Count++
			ts.Revenue += revenue

			dr := dailyAgg[b.Date]
			if dr == nil {
				dr = &DailyRevenue{Date: b.Date}
				dailyAgg[b.Date] = dr
			}
			dr.Bookings++
			dr.Revenue += revenue
		}
	}

	for _, s := range serviceAgg {
		stats.PopularServices = append(stats.PopularServices, *s)
	}
	sort.Slice(stats.PopularServices, func(i, j int) bool {
		return stats.PopularServices[i].Count > stats.PopularServices[j].Count
	})
	if len(stats.PopularServices) > 5 {
		stats.PopularServices = stats.PopularServices[:5]
	}

	for _, s := range stylistAgg {
		stats.TopStylists = append(stats.TopStylists, *s)
	}
	sort.Slice(stats.TopStylists, func(i, j int) bool {
		return stats.TopStylists[i].Count > stats.TopStylists[j].Count
	})
	if len(stats.TopStylists) > 5 {
		stats.TopStylists = stats.TopStylists[:5]
	}

	// Last 30 days, every day present even with zero bookings.
	for d := 0; d < 30; d++ {
		date := now.AddDate(0, 0, -(29 - d)).Format("2006-01-02")
		if dr := dailyAgg[date]; dr != nil {
			stats.DailyRevenue = append(stats.DailyRevenue, *dr)
		} else {
			stats.DailyRevenue = append(stats.DailyRevenue, DailyRevenue{Date: date})
		}
	}

	// Last 10 bookings by createdAt (RFC3339 sorts lexicographically).
	recent := make([]model.Booking, len(bookings))
	copy(recent, bookings)
	sort.Slice(recent, func(i, j int) bool { return recent[i].CreatedAt > recent[j].CreatedAt })
	if len(recent) > 10 {
		recent = recent[:10]
	}
	stats.RecentBookings = recent

	return stats
}

// BuildRevenueReport aggregates revenue per day within [from, to].
func BuildRevenueReport(bookings []model.Booking, from, to string) []DailyRevenue {
	dailyAgg := map[string]*DailyRevenue{}
	for i := range bookings {
		b := &bookings[i]
		if b.Date < from || b.Date > to {
			continue
		}
		dr := dailyAgg[b.Date]
		if dr == nil {
			dr = &DailyRevenue{Date: b.Date}
			dailyAgg[b.Date] = dr
		}
		dr.Bookings++
		if b.CountsAsRevenue() {
			dr.Revenue += b.Price
		}
	}
	report := []DailyRevenue{}
	start, err1 := time.Parse("2006-01-02", from)
	end, err2 := time.Parse("2006-01-02", to)
	if err1 != nil || err2 != nil || end.Before(start) {
		return report
	}
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		date := d.Format("2006-01-02")
		if dr := dailyAgg[date]; dr != nil {
			report = append(report, *dr)
		} else {
			report = append(report, DailyRevenue{Date: date})
		}
	}
	return report
}

// BuildCustomerSummaries groups bookings by customer phone.
// TotalSpent counts confirmed + completed bookings only.
func BuildCustomerSummaries(bookings []model.Booking) []CustomerSummary {
	agg := map[string]*CustomerSummary{}
	latestCreated := map[string]string{}

	for i := range bookings {
		b := &bookings[i]
		phone := b.Customer.Phone
		if phone == "" {
			phone = b.Phone
		}
		if phone == "" {
			continue
		}
		cs := agg[phone]
		if cs == nil {
			cs = &CustomerSummary{Phone: phone, FirstVisit: b.Date, LastVisit: b.Date}
			agg[phone] = cs
		}
		cs.TotalBookings++
		if b.CountsAsRevenue() {
			cs.TotalSpent += b.Price
		}
		if b.Date < cs.FirstVisit {
			cs.FirstVisit = b.Date
		}
		if b.Date > cs.LastVisit {
			cs.LastVisit = b.Date
		}
		// Use the most recently created booking for name/email.
		if b.CreatedAt >= latestCreated[phone] {
			latestCreated[phone] = b.CreatedAt
			cs.Name = b.Customer.Name
			cs.Email = b.Customer.Email
		}
	}

	customers := make([]CustomerSummary, 0, len(agg))
	for _, cs := range agg {
		customers = append(customers, *cs)
	}
	sort.Slice(customers, func(i, j int) bool { return customers[i].LastVisit > customers[j].LastVisit })
	return customers
}
