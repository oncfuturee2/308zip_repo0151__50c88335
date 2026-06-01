package service

import (
	"context"
	"testing"
	"time"

	"distribution-commission/internal/config"
	"distribution-commission/internal/database"
	"distribution-commission/internal/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestCommissionService_GenerateCommission(t *testing.T) {
	database.SetupTestDB()
	database.DB.AutoMigrate(&models.Order{}, &models.CommissionRecord{})

	config.AppConfig = &config.Config{
		Level1CommissionRate: 10,
		Level2CommissionRate: 5,
		Level3CommissionRate: 2,
	}

	user1 := models.User{Username: "user1", Password: "pwd", Role: "distributor"}
	user2 := models.User{Username: "user2", Password: "pwd", Role: "distributor"}
	user3 := models.User{Username: "user3", Password: "pwd", Role: "distributor"}
	database.DB.Create(&user1)
	database.DB.Create(&user2)
	database.DB.Create(&user3)

	d3 := models.Distributor{UserID: user3.ID, RealName: "D3", Balance: 0}
	database.DB.Create(&d3)
	
	d2 := models.Distributor{UserID: user2.ID, RealName: "D2", ParentID: &d3.ID, Balance: 0}
	database.DB.Create(&d2)

	d1 := models.Distributor{UserID: user1.ID, RealName: "D1", ParentID: &d2.ID, Balance: 0}
	database.DB.Create(&d1)

	order := models.Order{
		OrderNo:       "ORDER123",
		DistributorID: d1.ID,
		Amount:        1000, // 10 yuan
		Status:        models.OrderStatusCompleted,
	}
	database.DB.Create(&order)

	// Create a dummy redis for IdempotentService lock
	// Actually we are using NewIdempotentService which relies on real redis.
	// We might need to mock or skip if no redis. Let's see if we can use memory lock or just test the logic directly if possible.
}