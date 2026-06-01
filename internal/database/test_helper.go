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

	err = DB.AutoMigrate(
		&models.User{},
		&models.Distributor{},
		&models.CommissionRecord{},
		&models.Order{},
		&models.AuditLog{},
		&models.WithdrawRequest{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate test database: %v", err)
	}
}

func CleanupTestDB() {
	DB.Exec("DELETE FROM audit_logs")
	DB.Exec("DELETE FROM commission_records")
	DB.Exec("DELETE FROM withdraw_requests")
	DB.Exec("DELETE FROM orders")
	DB.Exec("DELETE FROM distributors")
	DB.Exec("DELETE FROM users")
}
