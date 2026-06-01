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
		0,
		"张三",
		"13800138000",
		"6222021234567890",
		"工商银行",
	)

	assert.NoError(t, err)
	assert.NotNil(t, distributor)
	assert.Equal(t, user.ID, distributor.UserID)
	assert.Nil(t, distributor.ParentID)
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

func TestDistributorService_CreateDistributor_WithParent_Success(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	userService := NewUserService()
	distributorService := NewDistributorService()

	parentUser, err := userService.CreateUser("parentuser", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	parentDistributor, err := distributorService.CreateDistributor(
		parentUser.ID,
		0,
		"父级",
		"13800138000",
		"6222021234567890",
		"工商银行",
	)
	assert.NoError(t, err)

	childUser, err := userService.CreateUser("childuser", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	childDistributor, err := distributorService.CreateDistributor(
		childUser.ID,
		parentDistributor.ID,
		"子级",
		"13900139000",
		"6222020987654321",
		"建设银行",
	)

	assert.NoError(t, err)
	assert.NotNil(t, childDistributor.ParentID)
	assert.Equal(t, parentDistributor.ID, *childDistributor.ParentID)
}

func TestDistributorService_CreateDistributor_WithInvalidParent(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	userService := NewUserService()
	distributorService := NewDistributorService()

	user, err := userService.CreateUser("testuser", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	_, err = distributorService.CreateDistributor(
		user.ID,
		999,
		"张三",
		"13800138000",
		"6222021234567890",
		"工商银行",
	)

	assert.Error(t, err)
	assert.Equal(t, "上级分销员不存在", err.Error())
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
		0,
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
		0,
		"张三",
		"13800138000",
		"6222021234567890",
		"工商银行",
	)
	assert.NoError(t, err)

	_, err = distributorService.CreateDistributor(
		user.ID,
		0,
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
		0,
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
		0,
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

func TestDistributorService_GetUplineChainTx_Success(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	userService := NewUserService()
	distributorService := NewDistributorService()

	user1, err := userService.CreateUser("level1", "password123", models.RoleDistributor)
	assert.NoError(t, err)
	level1, err := distributorService.CreateDistributor(user1.ID, 0, "一级", "", "", "")
	assert.NoError(t, err)

	user2, err := userService.CreateUser("level2", "password123", models.RoleDistributor)
	assert.NoError(t, err)
	level2, err := distributorService.CreateDistributor(user2.ID, level1.ID, "二级", "", "", "")
	assert.NoError(t, err)

	user3, err := userService.CreateUser("level3", "password123", models.RoleDistributor)
	assert.NoError(t, err)
	level3, err := distributorService.CreateDistributor(user3.ID, level2.ID, "三级", "", "", "")
	assert.NoError(t, err)

	chain, err := distributorService.GetUplineChainTx(database.DB, level3.ID, 3)

	assert.NoError(t, err)
	assert.Len(t, chain, 3)
	assert.Equal(t, level3.ID, chain[0].ID)
	assert.Equal(t, level2.ID, chain[1].ID)
	assert.Equal(t, level1.ID, chain[2].ID)
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
		0,
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
