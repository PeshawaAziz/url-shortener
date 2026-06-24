package database

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/PeshawaAziz/url-shortener/internal/config"
	"github.com/PeshawaAziz/url-shortener/internal/domain/url"
	"github.com/PeshawaAziz/url-shortener/internal/domain/user"
)

func NewPostgresDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := cfg.DSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := runMigrations(db); err != nil {
		return nil, err
	}

	return db, nil
}

func runMigrations(db *gorm.DB) error {
	log.Println("running database migrations...")
	err := db.AutoMigrate(
		&user.User{},
		&url.URL{},
		&url.Click{},
	)
	if err != nil {
		return err
	}
	log.Println("migrations completed successfully")
	return nil
}
