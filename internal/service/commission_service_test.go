package service

import (
	"context"
	"testing"

	"distribution-commission/internal/config"
	"distribution-commission/internal/database"
	"distribution-commission/internal/models"

	"github.com/stretchr/testify/assert"
)

func init() {
	config.AppConfig = &config.Config{
		DBHost:               "localhost",
		DBPort:               "5432",
		DBUser:               "postgres",
		DBPassword:           "postgres123",
		DBName:               "distribution",
		RedisHost:            "localhost",
		RedisPort:            "6379",
		RedisPassword:        "",
		JWTSecret:            "test_secret_key",
		JWTExpireHours:       24,
		CommissionRate:       10,
		Level1CommissionRate: 10,
		Level2CommissionRate: 5,
		Level3CommissionRate: 2,
	}
}

func setupTestDistributorHierarchy(t *testing.T) (uint, uint, uint) {
	database.SetupTestDB()

	userService := NewUserService()
	distributorService := NewDistributorService()

	user1, err := userService.CreateUser("dist_level1", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	dist1, err := distributorService.CreateDistributor(
		user1.ID,
		"一级分销商",
		"13800138001",
		"6222021234567890",
		"工商银行",
	)
	assert.NoError(t, err)

	user2, err := userService.CreateUser("dist_level2", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	dist2ParentID := dist1.ID
	dist2, err := distributorService.CreateDistributorWithParent(
		user2.ID,
		"二级分销商",
		"13800138002",
		"6222021234567891",
		"建设银行",
		&dist2ParentID,
	)
	assert.NoError(t, err)

	user3, err := userService.CreateUser("dist_level3", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	dist3ParentID := dist2.ID
	dist3, err := distributorService.CreateDistributorWithParent(
		user3.ID,
		"三级分销商",
		"13800138003",
		"6222021234567892",
		"农业银行",
		&dist3ParentID,
	)
	assert.NoError(t, err)

	return dist1.ID, dist2.ID, dist3.ID
}

func TestGenerateCommission_ThreeLevels_Success(t *testing.T) {
	dist1ID, dist2ID, dist3ID := setupTestDistributorHierarchy(t)
	defer database.CleanupTestDB()

	orderService := NewOrderService()
	commissionService := NewCommissionService()

	ctx := context.Background()
	orderNo := "ORD20260601TEST001"

	order, err := orderService.CreateCompletedOrder(ctx, orderNo, dist3ID, 10000, "测试商品")
	assert.NoError(t, err)
	assert.NotNil(t, order)

	commissions, err := commissionService.GenerateCommission(ctx, orderNo)
	assert.NoError(t, err)
	assert.Len(t, commissions, 3)

	var level1Comm, level2Comm, level3Comm *models.CommissionRecord
	for i := range commissions {
		switch commissions[i].Level {
		case 1:
			level1Comm = &commissions[i]
		case 2:
			level2Comm = &commissions[i]
		case 3:
			level3Comm = &commissions[i]
		}
	}

	assert.NotNil(t, level1Comm)
	assert.Equal(t, dist3ID, level1Comm.DistributorID)
	assert.Equal(t, 1, level1Comm.Level)
	assert.Equal(t, 10, level1Comm.Rate)
	assert.Equal(t, int64(1000), level1Comm.Amount)

	assert.NotNil(t, level2Comm)
	assert.Equal(t, dist2ID, level2Comm.DistributorID)
	assert.Equal(t, 2, level2Comm.Level)
	assert.Equal(t, 5, level2Comm.Rate)
	assert.Equal(t, int64(500), level2Comm.Amount)

	assert.NotNil(t, level3Comm)
	assert.Equal(t, dist1ID, level3Comm.DistributorID)
	assert.Equal(t, 3, level3Comm.Level)
	assert.Equal(t, 2, level3Comm.Rate)
	assert.Equal(t, int64(200), level3Comm.Amount)

	var count int64
	database.DB.Model(&models.CommissionRecord{}).Where("order_no = ?", orderNo).Count(&count)
	assert.Equal(t, int64(3), count)
}

func TestGenerateCommission_Idempotency(t *testing.T) {
	_, _, dist3ID := setupTestDistributorHierarchy(t)
	defer database.CleanupTestDB()

	orderService := NewOrderService()
	commissionService := NewCommissionService()

	ctx := context.Background()
	orderNo := "ORD20260601TEST002"

	_, err := orderService.CreateCompletedOrder(ctx, orderNo, dist3ID, 10000, "测试商品")
	assert.NoError(t, err)

	commissions1, err := commissionService.GenerateCommission(ctx, orderNo)
	assert.NoError(t, err)
	assert.Len(t, commissions1, 3)

	commissions2, err := commissionService.GenerateCommission(ctx, orderNo)
	assert.NoError(t, err)
	assert.Len(t, commissions2, 3)

	assert.Equal(t, commissions1[0].ID, commissions2[0].ID)
	assert.Equal(t, commissions1[1].ID, commissions2[1].ID)
	assert.Equal(t, commissions1[2].ID, commissions2[2].ID)

	var count int64
	database.DB.Model(&models.CommissionRecord{}).Where("order_no = ?", orderNo).Count(&count)
	assert.Equal(t, int64(3), count)
}

func TestGenerateCommission_SingleLevel_NoParent(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	userService := NewUserService()
	distributorService := NewDistributorService()

	user, err := userService.CreateUser("dist_single", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	dist, err := distributorService.CreateDistributor(
		user.ID,
		"独立分销商",
		"13800138004",
		"6222021234567893",
		"工商银行",
	)
	assert.NoError(t, err)

	orderService := NewOrderService()
	commissionService := NewCommissionService()

	ctx := context.Background()
	orderNo := "ORD20260601TEST003"

	_, err = orderService.CreateCompletedOrder(ctx, orderNo, dist.ID, 10000, "测试商品")
	assert.NoError(t, err)

	commissions, err := commissionService.GenerateCommission(ctx, orderNo)
	assert.NoError(t, err)
	assert.Len(t, commissions, 1)

	assert.Equal(t, dist.ID, commissions[0].DistributorID)
	assert.Equal(t, 1, commissions[0].Level)
	assert.Equal(t, 10, commissions[0].Rate)
	assert.Equal(t, int64(1000), commissions[0].Amount)
}

func TestGenerateCommission_TwoLevels_PartialHierarchy(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	userService := NewUserService()
	distributorService := NewDistributorService()

	user1, err := userService.CreateUser("dist_parent", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	dist1, err := distributorService.CreateDistributor(
		user1.ID,
		"上级分销商",
		"13800138005",
		"6222021234567894",
		"工商银行",
	)
	assert.NoError(t, err)

	user2, err := userService.CreateUser("dist_child", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	dist1ParentID := dist1.ID
	dist2, err := distributorService.CreateDistributorWithParent(
		user2.ID,
		"下级分销商",
		"13800138006",
		"6222021234567895",
		"建设银行",
		&dist1ParentID,
	)
	assert.NoError(t, err)

	orderService := NewOrderService()
	commissionService := NewCommissionService()

	ctx := context.Background()
	orderNo := "ORD20260601TEST004"

	_, err = orderService.CreateCompletedOrder(ctx, orderNo, dist2.ID, 10000, "测试商品")
	assert.NoError(t, err)

	commissions, err := commissionService.GenerateCommission(ctx, orderNo)
	assert.NoError(t, err)
	assert.Len(t, commissions, 2)

	var level1Comm, level2Comm *models.CommissionRecord
	for i := range commissions {
		switch commissions[i].Level {
		case 1:
			level1Comm = &commissions[i]
		case 2:
			level2Comm = &commissions[i]
		}
	}

	assert.NotNil(t, level1Comm)
	assert.Equal(t, dist2.ID, level1Comm.DistributorID)
	assert.Equal(t, 1, level1Comm.Level)
	assert.Equal(t, int64(1000), level1Comm.Amount)

	assert.NotNil(t, level2Comm)
	assert.Equal(t, dist1.ID, level2Comm.DistributorID)
	assert.Equal(t, 2, level2Comm.Level)
	assert.Equal(t, int64(500), level2Comm.Amount)
}

func TestGenerateCommission_OrderNotCompleted(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	userService := NewUserService()
	distributorService := NewDistributorService()

	user, err := userService.CreateUser("dist_test", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	dist, err := distributorService.CreateDistributor(
		user.ID,
		"测试分销商",
		"13800138007",
		"6222021234567896",
		"工商银行",
	)
	assert.NoError(t, err)

	order := &models.Order{
		OrderNo:       "ORD20260601TEST005",
		DistributorID: dist.ID,
		Amount:        10000,
		GoodsName:     "测试商品",
		Status:        models.OrderStatusPaid,
	}
	err = database.DB.Create(order).Error
	assert.NoError(t, err)

	commissionService := NewCommissionService()

	_, err = commissionService.GenerateCommission(context.Background(), order.OrderNo)
	assert.Error(t, err)
	assert.Equal(t, "订单未完成，不能生成佣金", err.Error())
}

func TestGenerateCommission_OrderNotFound(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	commissionService := NewCommissionService()

	_, err := commissionService.GenerateCommission(context.Background(), "NONEXISTENT_ORDER")
	assert.Error(t, err)
	assert.Equal(t, "订单不存在", err.Error())
}

func TestGenerateCommission_WithZeroLevel2Rate(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	originalLevel2Rate := config.AppConfig.Level2CommissionRate
	config.AppConfig.Level2CommissionRate = 0
	defer func() {
		config.AppConfig.Level2CommissionRate = originalLevel2Rate
	}()

	userService := NewUserService()
	distributorService := NewDistributorService()

	user1, err := userService.CreateUser("dist_l1", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	dist1, err := distributorService.CreateDistributor(
		user1.ID,
		"一级分销商",
		"13800138008",
		"6222021234567897",
		"工商银行",
	)
	assert.NoError(t, err)

	user2, err := userService.CreateUser("dist_l2", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	dist1ParentID := dist1.ID
	dist2, err := distributorService.CreateDistributorWithParent(
		user2.ID,
		"二级分销商",
		"13800138009",
		"6222021234567898",
		"建设银行",
		&dist1ParentID,
	)
	assert.NoError(t, err)

	orderService := NewOrderService()
	commissionService := NewCommissionService()

	ctx := context.Background()
	orderNo := "ORD20260601TEST006"

	_, err = orderService.CreateCompletedOrder(ctx, orderNo, dist2.ID, 10000, "测试商品")
	assert.NoError(t, err)

	commissions, err := commissionService.GenerateCommission(ctx, orderNo)
	assert.NoError(t, err)
	assert.Len(t, commissions, 1)

	assert.Equal(t, dist2.ID, commissions[0].DistributorID)
	assert.Equal(t, 1, commissions[0].Level)
}

func TestSettleCommission_Success(t *testing.T) {
	dist1ID, dist2ID, dist3ID := setupTestDistributorHierarchy(t)
	defer database.CleanupTestDB()

	orderService := NewOrderService()
	commissionService := NewCommissionService()

	ctx := context.Background()
	orderNo := "ORD20260601TEST007"

	_, err := orderService.CreateCompletedOrder(ctx, orderNo, dist3ID, 10000, "测试商品")
	assert.NoError(t, err)

	commissions, err := commissionService.GenerateCommission(ctx, orderNo)
	assert.NoError(t, err)
	assert.Len(t, commissions, 3)

	var dist3OriginalBalance, dist2OriginalBalance, dist1OriginalBalance int64
	database.DB.Model(&models.Distributor{}).Where("id = ?", dist3ID).Pluck("balance", &dist3OriginalBalance)
	database.DB.Model(&models.Distributor{}).Where("id = ?", dist2ID).Pluck("balance", &dist2OriginalBalance)
	database.DB.Model(&models.Distributor{}).Where("id = ?", dist1ID).Pluck("balance", &dist1OriginalBalance)

	err = commissionService.SettleCommission(commissions[0].ID, 1)
	assert.NoError(t, err)

	var settledCommission models.CommissionRecord
	database.DB.First(&settledCommission, commissions[0].ID)
	assert.Equal(t, models.CommissionStatusSettled, settledCommission.Status)
	assert.NotNil(t, settledCommission.SettledAt)

	var dist3Balance int64
	database.DB.Model(&models.Distributor{}).Where("id = ?", dist3ID).Pluck("balance", &dist3Balance)
	assert.Equal(t, dist3OriginalBalance+commissions[0].Amount, dist3Balance)
}

func TestSettleCommission_AlreadySettled(t *testing.T) {
	_, _, dist3ID := setupTestDistributorHierarchy(t)
	defer database.CleanupTestDB()

	orderService := NewOrderService()
	commissionService := NewCommissionService()

	ctx := context.Background()
	orderNo := "ORD20260601TEST008"

	_, err := orderService.CreateCompletedOrder(ctx, orderNo, dist3ID, 10000, "测试商品")
	assert.NoError(t, err)

	commissions, err := commissionService.GenerateCommission(ctx, orderNo)
	assert.NoError(t, err)

	err = commissionService.SettleCommission(commissions[0].ID, 1)
	assert.NoError(t, err)

	err = commissionService.SettleCommission(commissions[0].ID, 1)
	assert.Error(t, err)
	assert.Equal(t, "只有待结算状态的佣金可以结算", err.Error())
}
