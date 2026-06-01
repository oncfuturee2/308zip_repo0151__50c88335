package service

import (
	"context"
	"testing"

	"distribution-commission/internal/config"
	"distribution-commission/internal/database"
	"distribution-commission/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestCommissionService_GenerateCommission_ThreeLevels(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	config.AppConfig = &config.Config{
		CommissionRate:       10,
		CommissionLevel2Rate: 5,
		CommissionLevel3Rate: 2,
	}

	userService := NewUserService()
	distributorService := NewDistributorService()
	orderService := NewOrderService()
	commissionService := NewCommissionService()

	level1User, err := userService.CreateUser("level1", "password123", models.RoleDistributor)
	assert.NoError(t, err)
	level1, err := distributorService.CreateDistributor(level1User.ID, 0, "一级", "", "", "")
	assert.NoError(t, err)

	level2User, err := userService.CreateUser("level2", "password123", models.RoleDistributor)
	assert.NoError(t, err)
	level2, err := distributorService.CreateDistributor(level2User.ID, level1.ID, "二级", "", "", "")
	assert.NoError(t, err)

	level3User, err := userService.CreateUser("level3", "password123", models.RoleDistributor)
	assert.NoError(t, err)
	level3, err := distributorService.CreateDistributor(level3User.ID, level2.ID, "三级", "", "", "")
	assert.NoError(t, err)

	_, err = orderService.CreateCompletedOrder(context.Background(), "ORD-THREE-LEVEL", level3.ID, 10000, "测试商品")
	assert.NoError(t, err)

	commissions, err := commissionService.GenerateCommission(context.Background(), "ORD-THREE-LEVEL")

	assert.NoError(t, err)
	assert.Len(t, commissions, 3)

	assert.Equal(t, level3.ID, commissions[0].DistributorID)
	assert.Equal(t, level3.ID, commissions[0].SourceDistributorID)
	assert.Equal(t, models.CommissionLevel1, commissions[0].Level)
	assert.Equal(t, 10, commissions[0].Rate)
	assert.Equal(t, int64(1000), commissions[0].Amount)

	assert.Equal(t, level2.ID, commissions[1].DistributorID)
	assert.Equal(t, level3.ID, commissions[1].SourceDistributorID)
	assert.Equal(t, models.CommissionLevel2, commissions[1].Level)
	assert.Equal(t, 5, commissions[1].Rate)
	assert.Equal(t, int64(500), commissions[1].Amount)

	assert.Equal(t, level1.ID, commissions[2].DistributorID)
	assert.Equal(t, level3.ID, commissions[2].SourceDistributorID)
	assert.Equal(t, models.CommissionLevel3, commissions[2].Level)
	assert.Equal(t, 2, commissions[2].Rate)
	assert.Equal(t, int64(200), commissions[2].Amount)

	var records []models.CommissionRecord
	err = database.DB.Order("level ASC").Find(&records).Error
	assert.NoError(t, err)
	assert.Len(t, records, 3)
}

func TestCommissionService_GenerateCommission_IdempotentByOrderNo(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	config.AppConfig = &config.Config{
		CommissionRate:       10,
		CommissionLevel2Rate: 5,
		CommissionLevel3Rate: 2,
	}

	userService := NewUserService()
	distributorService := NewDistributorService()
	orderService := NewOrderService()
	commissionService := NewCommissionService()

	parentUser, err := userService.CreateUser("parent", "password123", models.RoleDistributor)
	assert.NoError(t, err)
	parentDistributor, err := distributorService.CreateDistributor(parentUser.ID, 0, "父级", "", "", "")
	assert.NoError(t, err)

	childUser, err := userService.CreateUser("child", "password123", models.RoleDistributor)
	assert.NoError(t, err)
	childDistributor, err := distributorService.CreateDistributor(childUser.ID, parentDistributor.ID, "子级", "", "", "")
	assert.NoError(t, err)

	_, err = orderService.CreateCompletedOrder(context.Background(), "ORD-IDEMPOTENT", childDistributor.ID, 10000, "测试商品")
	assert.NoError(t, err)

	first, err := commissionService.GenerateCommission(context.Background(), "ORD-IDEMPOTENT")
	assert.NoError(t, err)
	assert.Len(t, first, 2)

	second, err := commissionService.GenerateCommission(context.Background(), "ORD-IDEMPOTENT")
	assert.NoError(t, err)
	assert.Len(t, second, 2)
	assert.Equal(t, first[0].ID, second[0].ID)
	assert.Equal(t, first[1].ID, second[1].ID)

	var count int64
	err = database.DB.Model(&models.CommissionRecord{}).Where("order_no = ?", "ORD-IDEMPOTENT").Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}
