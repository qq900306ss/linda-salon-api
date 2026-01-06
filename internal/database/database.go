package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"linda-salon-api/config"
	"linda-salon-api/internal/model"
)

type Database struct {
	DB *gorm.DB
}

func New(cfg *config.DatabaseConfig) (*Database, error) {
	dsn := cfg.GetDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("✅ Database connected successfully")

	return &Database{DB: db}, nil
}

func (d *Database) AutoMigrate() error {
	log.Println("🔄 Running database migrations...")

	err := d.DB.AutoMigrate(
		&model.User{},
		&model.Service{},
		&model.Stylist{},
		&model.StylistSchedule{},
		&model.Booking{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Manual migration: Make phone, google_id, and line_id nullable
	log.Println("🔄 Running manual migrations for nullable fields...")

	// Check if phone column needs to be made nullable
	var phoneNullable string
	d.DB.Raw("SELECT is_nullable FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'phone'").Scan(&phoneNullable)
	if phoneNullable == "NO" {
		log.Println("  - Making phone column nullable")
		if err := d.DB.Exec("ALTER TABLE users ALTER COLUMN phone DROP NOT NULL").Error; err != nil {
			log.Printf("⚠️  Warning: Failed to make phone nullable: %v", err)
		}
	}

	// Check if google_id column needs to be made nullable
	var googleIDNullable string
	d.DB.Raw("SELECT is_nullable FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'google_id'").Scan(&googleIDNullable)
	if googleIDNullable == "NO" {
		log.Println("  - Making google_id column nullable")
		if err := d.DB.Exec("ALTER TABLE users ALTER COLUMN google_id DROP NOT NULL").Error; err != nil {
			log.Printf("⚠️  Warning: Failed to make google_id nullable: %v", err)
		}
	}

	// Check if line_id column needs to be made nullable
	var lineIDNullable string
	d.DB.Raw("SELECT is_nullable FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'line_id'").Scan(&lineIDNullable)
	if lineIDNullable == "NO" {
		log.Println("  - Making line_id column nullable")
		if err := d.DB.Exec("ALTER TABLE users ALTER COLUMN line_id DROP NOT NULL").Error; err != nil {
			log.Printf("⚠️  Warning: Failed to make line_id nullable: %v", err)
		}
	}

	log.Println("✅ Database migrations completed")
	return nil
}

func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
