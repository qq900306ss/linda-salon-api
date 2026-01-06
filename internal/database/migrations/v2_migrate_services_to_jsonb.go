package migrations

import (
	"encoding/json"
	"log"

	"gorm.io/gorm"
)

// V2MigrateServicesToJSONB migrates service_id to services JSONB array
func V2MigrateServicesToJSONB(tx *gorm.DB) error {
	log.Println("  [V2] Migrating bookings.service_id to services JSONB...")

	// Step 1: Add services JSONB column
	log.Println("    - Adding services JSONB column")
	if err := tx.Exec("ALTER TABLE bookings ADD COLUMN IF NOT EXISTS services JSONB").Error; err != nil {
		return err
	}

	// Step 2: Migrate existing data from service_id to services JSONB
	log.Println("    - Migrating existing service_id data to services array")

	// Get all bookings with their service info
	var bookings []struct {
		ID        uint
		ServiceID uint
	}
	if err := tx.Raw("SELECT id, service_id FROM bookings WHERE service_id IS NOT NULL").Scan(&bookings).Error; err != nil {
		return err
	}

	log.Printf("    - Found %d bookings to migrate", len(bookings))

	// For each booking, fetch service details and create JSONB array
	for _, booking := range bookings {
		var service struct {
			ID       uint
			Name     string
			Price    int
			Duration int
		}

		// Get service details
		err := tx.Raw("SELECT id, name, price, duration FROM services WHERE id = ?", booking.ServiceID).Scan(&service).Error
		if err != nil || service.ID == 0 {
			// If service not found, create a placeholder with booking data
			log.Printf("    - Warning: Could not find service with id %d for booking %d, using placeholder", booking.ServiceID, booking.ID)

			// Get price and duration from booking record
			var bookingData struct {
				Price    int
				Duration int
			}
			tx.Raw("SELECT price, duration FROM bookings WHERE id = ?", booking.ID).Scan(&bookingData)

			service = struct {
				ID       uint
				Name     string
				Price    int
				Duration int
			}{
				ID:       booking.ServiceID,
				Name:     "未知服務",
				Price:    bookingData.Price,
				Duration: bookingData.Duration,
			}
		}

		// Create JSONB array with single service
		servicesJSON, err := json.Marshal([]map[string]interface{}{
			{
				"id":       service.ID,
				"name":     service.Name,
				"price":    service.Price,
				"duration": service.Duration,
			},
		})
		if err != nil {
			return err
		}

		// Update booking with services JSONB
		if err := tx.Exec("UPDATE bookings SET services = ? WHERE id = ?", servicesJSON, booking.ID).Error; err != nil {
			return err
		}
	}

	// Step 3: Make services column NOT NULL (now that all data is migrated)
	log.Println("    - Making services column NOT NULL")
	if err := tx.Exec("ALTER TABLE bookings ALTER COLUMN services SET NOT NULL").Error; err != nil {
		return err
	}

	// Step 4: Drop service_id column
	log.Println("    - Dropping service_id column")
	if err := tx.Exec("ALTER TABLE bookings DROP COLUMN IF EXISTS service_id").Error; err != nil {
		return err
	}

	log.Println("    - Migration completed successfully")
	return nil
}
