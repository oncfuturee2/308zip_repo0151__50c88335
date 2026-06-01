package service

import (
	"context"
	"testing"

	"distribution-commission/internal/database"
	"distribution-commission/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestDistributorService_CreateDistributor_Success(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	userService := NewUserService()
	distributorService := NewDistributorService()

	user, err := userService.CreateUser("testuser", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	distributor, err := distributorService.CreateDistributor(
		user.ID,
		"张三",
		"13800138000",
		"6222021234567890",
		"工商银行",
		nil,
	)

	assert.NoError(t, err)
	assert.NotNil(t, distributor)
	assert.Equal(t, user.ID, distributor.UserID)
	assert.Equal(t, "张三", distributor.RealName)
	assert.Equal(t, "13800138000", distributor.Phone)
	assert.Equal(t, "6222021234567890", distributor.BankCardNo)
	assert.Equal(t, "工商银行", distributor.BankName)
	assert.Equal(t, 1, distributor.Status)
	assert.Greater(t, distributor.ID, uint(0))

	var count int64
	database.DB.Model(&models.Distributor{}).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestDistributorService_CreateDistributor_WithOptionalFieldsEmpty(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	userService := NewUserService()
	distributorService := NewDistributorService()

	user, err := userService.CreateUser("testuser", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	distributor, err := distributorService.CreateDistributor(
		user.ID,
		"",
		"",
		"",
		"",
		nil,
	)

	assert.NoError(t, err)
	assert.NotNil(t, distributor)
	assert.Equal(t, user.ID, distributor.UserID)
	assert.Empty(t, distributor.RealName)
	assert.Empty(t, distributor.Phone)
	assert.Empty(t, distributor.BankCardNo)
	assert.Empty(t, distributor.BankName)
}

func TestDistributorService_CreateDistributor_DuplicateUserID(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	userService := NewUserService()
	distributorService := NewDistributorService()

	user, err := userService.CreateUser("testuser", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	_, err = distributorService.CreateDistributor(
		user.ID,
		"张三",
		"13800138000",
		"6222021234567890",
		"工商银行",
		nil,
	)
	assert.NoError(t, err)

	_, err = distributorService.CreateDistributor(
		user.ID,
		"李四",
		"13900139000",
		"6222020987654321",
		"建设银行",
		nil,
	)

	assert.Error(t, err)

	var count int64
	database.DB.Model(&models.Distributor{}).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestDistributorService_GetByUserID_Success(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	userService := NewUserService()
	distributorService := NewDistributorService()

	user, err := userService.CreateUser("testuser", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	createdDistributor, err := distributorService.CreateDistributor(
		user.ID,
		"张三",
		"13800138000",
		"6222021234567890",
		"工商银行",
		nil,
	)
	assert.NoError(t, err)

	foundDistributor, err := distributorService.GetByUserID(user.ID)

	assert.NoError(t, err)
	assert.NotNil(t, foundDistributor)
	assert.Equal(t, createdDistributor.ID, foundDistributor.ID)
	assert.Equal(t, "张三", foundDistributor.RealName)
	assert.Equal(t, user.Username, foundDistributor.User.Username)
}

func TestDistributorService_GetByUserID_NotFound(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	distributorService := NewDistributorService()

	foundDistributor, err := distributorService.GetByUserID(999)

	assert.Error(t, err)
	assert.Nil(t, foundDistributor)
}

func TestDistributorService_GetByID_Success(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	userService := NewUserService()
	distributorService := NewDistributorService()

	user, err := userService.CreateUser("testuser", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	createdDistributor, err := distributorService.CreateDistributor(
		user.ID,
		"张三",
		"13800138000",
		"6222021234567890",
		"工商银行",
		nil,
	)
	assert.NoError(t, err)

	foundDistributor, err := distributorService.GetByID(createdDistributor.ID)

	assert.NoError(t, err)
	assert.NotNil(t, foundDistributor)
	assert.Equal(t, createdDistributor.ID, foundDistributor.ID)
	assert.Equal(t, user.Username, foundDistributor.User.Username)
}

func TestDistributorService_GetByID_NotFound(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	distributorService := NewDistributorService()

	foundDistributor, err := distributorService.GetByID(999)

	assert.Error(t, err)
	assert.Nil(t, foundDistributor)
}

func TestDistributorService_GetBalance_Success(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	userService := NewUserService()
	distributorService := NewDistributorService()

	user, err := userService.CreateUser("testuser", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	distributor := &models.Distributor{
		UserID:          user.ID,
		Balance:         10000,
		FrozenBalance:   5000,
		TotalCommission: 15000,
		Status:          1,
	}
	database.DB.Create(distributor)

	balance, frozenBalance, totalCommission, err := distributorService.GetBalance(distributor.ID)

	assert.NoError(t, err)
	assert.Equal(t, int64(10000), balance)
	assert.Equal(t, int64(5000), frozenBalance)
	assert.Equal(t, int64(15000), totalCommission)
}

func TestDistributorService_UpdateBankInfo_Success(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	userService := NewUserService()
	distributorService := NewDistributorService()

	user, err := userService.CreateUser("testuser", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	distributor, err := distributorService.CreateDistributor(
		user.ID,
		"张三",
		"13800138000",
		"6222021234567890",
		"工商银行",
		nil,
	)
	assert.NoError(t, err)

	err = distributorService.UpdateBankInfo(
		distributor.ID,
		"6222020987654321",
		"建设银行",
		"李四",
	)

	assert.NoError(t, err)

	var updatedDistributor models.Distributor
	database.DB.First(&updatedDistributor, distributor.ID)
	assert.Equal(t, "6222020987654321", updatedDistributor.BankCardNo)
	assert.Equal(t, "建设银行", updatedDistributor.BankName)
	assert.Equal(t, "李四", updatedDistributor.RealName)
}

func TestDistributorService_CreateDistributor_WithParent(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	userService := NewUserService()
	distributorService := NewDistributorService()

	// 创建一级分销商
	user1, err := userService.CreateUser("user1", "password123", models.RoleDistributor)
	assert.NoError(t, err)
	distributor1, err := distributorService.CreateDistributor(user1.ID, "一级", "138001", "622", "银行", nil)
	assert.NoError(t, err)

	// 创建二级分销商
	user2, err := userService.CreateUser("user2", "password123", models.RoleDistributor)
	assert.NoError(t, err)
	distributor2, err := distributorService.CreateDistributor(user2.ID, "二级", "138002", "623", "银行", &distributor1.ID)
	assert.NoError(t, err)
	assert.Equal(t, distributor1.ID, *distributor2.ParentID)
	assert.Equal(t, "1,2", distributor2.Path)

	// 创建三级分销商
	user3, err := userService.CreateUser("user3", "password123", models.RoleDistributor)
	assert.NoError(t, err)
	distributor3, err := distributorService.CreateDistributor(user3.ID, "三级", "138003", "624", "银行", &distributor2.ID)
	assert.NoError(t, err)
	assert.Equal(t, distributor2.ID, *distributor3.ParentID)
	assert.Equal(t, "1,2,3", distributor3.Path)
}

func TestCommissionService_GenerateCommission_ThreeLevels(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	ctx := context.Background()
	userService := NewUserService()
	distributorService := NewDistributorService()
	commissionService := NewCommissionService()

	// 创建三级分销商
	user1, _ := userService.CreateUser("user1", "pass", models.RoleDistributor)
	distributor1, _ := distributorService.CreateDistributor(user1.ID, "一级", "", "", "", nil)

	user2, _ := userService.CreateUser("user2", "pass", models.RoleDistributor)
	distributor2, _ := distributorService.CreateDistributor(user2.ID, "二级", "", "", "", &distributor1.ID)

	user3, _ := userService.CreateUser("user3", "pass", models.RoleDistributor)
	distributor3, _ := distributorService.CreateDistributor(user3.ID, "三级", "", "", "", &distributor2.ID)

	// 直接创建已完成订单
	orderNo := "TEST_ORDER_001"
	order := &models.Order{
		OrderNo:       orderNo,
		DistributorID: distributor3.ID,
		Amount:        10000,
		GoodsName:     "测试商品",
		Status:        models.OrderStatusCompleted,
	}
	err := database.DB.Create(order).Error
	assert.NoError(t, err)

	// 生成佣金
	commissions, err := commissionService.GenerateCommission(ctx, orderNo)
	assert.NoError(t, err)
	assert.Len(t, commissions, 3)

	// 验证各级佣金
	expectedLevels := []int{1, 2, 3}
	expectedDistributorIDs := []uint{distributor3.ID, distributor2.ID, distributor1.ID}
	expectedAmounts := []int64{1000, 500, 300} // 10% 5% 3% of 10000

	for i, comm := range commissions {
		assert.Equal(t, expectedLevels[i], comm.Level)
		assert.Equal(t, expectedDistributorIDs[i], comm.DistributorID)
		assert.Equal(t, expectedAmounts[i], comm.Amount)
	}
}

