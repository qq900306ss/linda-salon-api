package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
	"linda-salon-api/internal/database/migrations"
)

// Migration represents a database migration record
type Migration struct {
	ID        uint      `gorm:"primarykey"`
	Version   string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	Name      string    `gorm:"type:varchar(255);not null"`
	AppliedAt time.Time `gorm:"not null"`
}

// MigrationFunc is a function that performs a migration
type MigrationFunc func(*gorm.DB) error

// migrationList holds all migrations in order
var migrationList = []struct {
	version string
	name    string
	fn      MigrationFunc
}{
	{
		version: "v1",
		name:    "make_user_fields_nullable",
		fn:      migrations.V1MakeUserFieldsNullable,
	},
	// Add new migrations here in order
	// {
	//     version: "v2",
	//     name:    "add_new_table",
	//     fn:      migrations.V2AddNewTable,
	// },
}

// RunMigrations runs all pending migrations
func (d *Database) RunMigrations() error {
	// Create migrations table if not exists
	if err := d.DB.AutoMigrate(&Migration{}); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	log.Println("🔄 Checking for pending migrations...")

	// Get list of applied migrations
	var appliedMigrations []Migration
	if err := d.DB.Find(&appliedMigrations).Error; err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	appliedMap := make(map[string]bool)
	for _, m := range appliedMigrations {
		appliedMap[m.Version] = true
	}

	// Run pending migrations
	pendingCount := 0
	for _, migration := range migrationList {
		if appliedMap[migration.version] {
			log.Printf("⏭️  Skipping migration %s (%s) - already applied", migration.version, migration.name)
			continue
		}

		log.Printf("📝 Running migration %s: %s", migration.version, migration.name)

		// Run migration in transaction
		err := d.DB.Transaction(func(tx *gorm.DB) error {
			// Execute migration
			if err := migration.fn(tx); err != nil {
				return fmt.Errorf("migration failed: %w", err)
			}

			// Record migration as applied
			record := Migration{
				Version:   migration.version,
				Name:      migration.name,
				AppliedAt: time.Now().UTC(),
			}
			if err := tx.Create(&record).Error; err != nil {
				return fmt.Errorf("failed to record migration: %w", err)
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration.version, err)
		}

		log.Printf("✅ Migration %s completed: %s", migration.version, migration.name)
		pendingCount++
	}

	if pendingCount == 0 {
		log.Println("✅ No pending migrations")
	} else {
		log.Printf("✅ Applied %d migration(s)", pendingCount)
	}

	return nil
}
