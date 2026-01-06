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
	log.Println("🔄 Running GORM auto-migrations...")

	err := d.DB.AutoMigrate(
		&model.User{},
		&model.Service{},
		&model.Stylist{},
		&model.StylistSchedule{},
		&model.Booking{},
	)
	if err != nil {
		return fmt.Errorf("failed to run auto-migrations: %w", err)
	}

	log.Println("✅ GORM auto-migrations completed")

	// Run custom migrations
	if err := d.RunMigrations(); err != nil {
		return fmt.Errorf("failed to run custom migrations: %w", err)
	}

	return nil
}

func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
