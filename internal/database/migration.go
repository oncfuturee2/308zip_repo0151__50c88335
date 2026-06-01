package database

import (
	"log"

	"distribution-commission/internal/models"
)

func RunMigrations() {
	err := DB.AutoMigrate(
		&models.User{},
		&models.Distributor{},
		&models.Order{},
		&models.CommissionRecord{},
		&models.WithdrawRequest{},
		&models.AuditLog{},
	)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	if err := DB.Exec("DROP INDEX IF EXISTS idx_order_commission").Error; err != nil {
		log.Fatalf("Failed to drop legacy commission index: %v", err)
	}

	log.Println("Migrations completed successfully")
}
