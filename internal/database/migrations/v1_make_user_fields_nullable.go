package migrations

import (
	"log"

	"gorm.io/gorm"
)

// V1MakeUserFieldsNullable makes phone, google_id, and line_id nullable
func V1MakeUserFieldsNullable(tx *gorm.DB) error {
	log.Println("  [V1] Making user fields nullable...")

	// Check and modify phone column
	var phoneNullable string
	tx.Raw("SELECT is_nullable FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'phone'").Scan(&phoneNullable)
	if phoneNullable == "NO" {
		log.Println("    - Making phone column nullable")
		if err := tx.Exec("ALTER TABLE users ALTER COLUMN phone DROP NOT NULL").Error; err != nil {
			return err
		}
	}

	// Check and modify google_id column
	var googleIDNullable string
	tx.Raw("SELECT is_nullable FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'google_id'").Scan(&googleIDNullable)
	if googleIDNullable == "NO" {
		log.Println("    - Making google_id column nullable")
		if err := tx.Exec("ALTER TABLE users ALTER COLUMN google_id DROP NOT NULL").Error; err != nil {
			return err
		}
	}

	// Check and modify line_id column
	var lineIDNullable string
	tx.Raw("SELECT is_nullable FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'line_id'").Scan(&lineIDNullable)
	if lineIDNullable == "NO" {
		log.Println("    - Making line_id column nullable")
		if err := tx.Exec("ALTER TABLE users ALTER COLUMN line_id DROP NOT NULL").Error; err != nil {
			return err
		}
	}

	return nil
}
