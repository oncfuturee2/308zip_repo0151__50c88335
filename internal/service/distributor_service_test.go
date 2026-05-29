package service

import (
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
	)
	assert.NoError(t, err)

	_, err = distributorService.CreateDistributor(
		user.ID,
		"李四",
		"13900139000",
		"6222020987654321",
		"建设银行",
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
