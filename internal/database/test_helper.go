package database

import (
	"distribution-commission/internal/models"
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SetupTestDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying test database: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	if err := DB.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		log.Fatalf("Failed to enable SQLite foreign keys: %v", err)
	}

	err = DB.AutoMigrate(
		&models.User{},
		&models.Distributor{},
		&models.Order{},
		&models.CommissionRecord{},
		&models.WithdrawRequest{},
		&models.AuditLog{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate test database: %v", err)
	}
}

func CleanupTestDB() {
	if DB == nil {
		return
	}

	db := DB.Session(&gorm.Session{AllowGlobalUpdate: true})
	db.Unscoped().Delete(&models.AuditLog{})
	db.Unscoped().Delete(&models.WithdrawRequest{})
	db.Unscoped().Delete(&models.CommissionRecord{})
	db.Unscoped().Delete(&models.Order{})
	db.Unscoped().Delete(&models.Distributor{})
	db.Unscoped().Delete(&models.User{})

	sqlDB, err := DB.DB()
	if err == nil {
		_ = sqlDB.Close()
	}
	DB = nil
}
