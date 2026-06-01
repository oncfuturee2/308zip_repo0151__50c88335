package service

import (
	"testing"

	"distribution-commission/internal/database"
	"distribution-commission/internal/models"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type stubPayoutProvider struct {
	validateResult bool
	validateErr    error
	validateFunc   func(bankCardNo, bankName, realName string) (bool, error)
}

func (p stubPayoutProvider) Payout(withdraw *models.WithdrawRequest) (string, error) {
	return "", nil
}

func (p stubPayoutProvider) ValidateBankCard(bankCardNo, bankName, realName string) (bool, error) {
	if p.validateFunc != nil {
		return p.validateFunc(bankCardNo, bankName, realName)
	}

	return p.validateResult, p.validateErr
}

func setupWithdrawServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	database.DB = db
	require.NoError(t, database.DB.AutoMigrate(
		&models.User{},
		&models.Distributor{},
		&models.WithdrawRequest{},
		&models.AuditLog{},
	))

	t.Cleanup(func() {
		require.NoError(t, sqlDB.Close())
	})

	return db
}

func createTestDistributor(t *testing.T, balance int64) *models.Distributor {
	t.Helper()

	user := &models.User{
		Username: "withdraw-user",
		Password: "password123",
		Role:     models.RoleDistributor,
	}
	require.NoError(t, database.DB.Create(user).Error)

	distributor := &models.Distributor{
		UserID:        user.ID,
		RealName:      "张三",
		BankCardNo:    "6222021234567890",
		BankName:      "工商银行",
		Balance:       balance,
		FrozenBalance: 0,
		Status:        1,
	}
	require.NoError(t, database.DB.Create(distributor).Error)

	return distributor
}

func TestWithdrawService_SubmitWithdraw_Success(t *testing.T) {
	setupWithdrawServiceTestDB(t)
	service := &WithdrawService{
		payoutProvider: stubPayoutProvider{validateResult: true},
	}

	distributor := createTestDistributor(t, 10000)

	withdraw, err := service.SubmitWithdraw(distributor.ID, 3000)

	require.NoError(t, err)
	require.NotNil(t, withdraw)
	assert.Equal(t, distributor.ID, withdraw.DistributorID)
	assert.Equal(t, int64(3000), withdraw.Amount)
	assert.Equal(t, models.WithdrawStatusPending, withdraw.Status)

	var updatedDistributor models.Distributor
	require.NoError(t, database.DB.First(&updatedDistributor, distributor.ID).Error)
	assert.Equal(t, int64(7000), updatedDistributor.Balance)
	assert.Equal(t, int64(3000), updatedDistributor.FrozenBalance)

	var count int64
	require.NoError(t, database.DB.Model(&models.WithdrawRequest{}).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestWithdrawService_SubmitWithdraw_InsufficientBalance(t *testing.T) {
	setupWithdrawServiceTestDB(t)
	service := &WithdrawService{
		payoutProvider: stubPayoutProvider{validateResult: true},
	}

	distributor := createTestDistributor(t, 2000)

	withdraw, err := service.SubmitWithdraw(distributor.ID, 3000)

	require.Error(t, err)
	assert.Equal(t, "余额不足", err.Error())
	assert.Nil(t, withdraw)

	var updatedDistributor models.Distributor
	require.NoError(t, database.DB.First(&updatedDistributor, distributor.ID).Error)
	assert.Equal(t, int64(2000), updatedDistributor.Balance)
	assert.Equal(t, int64(0), updatedDistributor.FrozenBalance)

	var count int64
	require.NoError(t, database.DB.Model(&models.WithdrawRequest{}).Count(&count).Error)
	assert.Equal(t, int64(0), count)
}

func TestWithdrawService_SubmitWithdraw_BankInfoChangedRollsBack(t *testing.T) {
	setupWithdrawServiceTestDB(t)
	distributor := createTestDistributor(t, 10000)

	service := &WithdrawService{
		payoutProvider: stubPayoutProvider{
			validateFunc: func(bankCardNo, bankName, realName string) (bool, error) {
				err := database.DB.Model(&models.Distributor{}).Where("id = ?", distributor.ID).Updates(map[string]interface{}{
					"bank_name": "建设银行",
				}).Error
				require.NoError(t, err)
				return true, nil
			},
		},
	}

	withdraw, err := service.SubmitWithdraw(distributor.ID, 3000)

	require.Error(t, err)
	assert.Equal(t, "银行卡信息已变更，请重新提交提现申请", err.Error())
	assert.Nil(t, withdraw)

	var updatedDistributor models.Distributor
	require.NoError(t, database.DB.First(&updatedDistributor, distributor.ID).Error)
	assert.Equal(t, int64(10000), updatedDistributor.Balance)
	assert.Equal(t, int64(0), updatedDistributor.FrozenBalance)
	assert.Equal(t, "建设银行", updatedDistributor.BankName)

	var count int64
	require.NoError(t, database.DB.Model(&models.WithdrawRequest{}).Count(&count).Error)
	assert.Equal(t, int64(0), count)
}
