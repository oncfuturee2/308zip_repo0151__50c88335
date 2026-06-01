package service

import (
	"fmt"
	"testing"

	"distribution-commission/internal/database"
	"distribution-commission/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCommissionService_SettleCommission_Success(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	service := NewCommissionService()
	initialBalance := int64(2000)
	initialTotalCommission := int64(5000)
	commissionAmount := int64(1800)

	distributor, commission, operator := createSettlementFixture(t, models.CommissionStatusPending, initialBalance, initialTotalCommission, commissionAmount)

	err := service.SettleCommission(commission.ID, operator.ID)

	assert.NoError(t, err)

	var updatedCommission models.CommissionRecord
	err = database.DB.First(&updatedCommission, commission.ID).Error
	require.NoError(t, err)
	assert.Equal(t, models.CommissionStatusSettled, updatedCommission.Status)
	if assert.NotNil(t, updatedCommission.SettledAt) {
		assert.False(t, updatedCommission.SettledAt.IsZero())
	}

	var updatedDistributor models.Distributor
	err = database.DB.First(&updatedDistributor, distributor.ID).Error
	require.NoError(t, err)
	assert.Equal(t, initialBalance+commissionAmount, updatedDistributor.Balance)
	assert.Equal(t, initialTotalCommission+commissionAmount, updatedDistributor.TotalCommission)

	var auditLogs []models.AuditLog
	err = database.DB.Where("resource_type = ? AND resource_id = ?", "commission", commission.ID).Find(&auditLogs).Error
	require.NoError(t, err)
	require.Len(t, auditLogs, 1)
	assert.Equal(t, distributor.ID, auditLogs[0].DistributorID)
	assert.Equal(t, operator.ID, auditLogs[0].OperatorID)
	assert.Equal(t, "commission", auditLogs[0].ResourceType)
	assert.Equal(t, commission.ID, auditLogs[0].ResourceID)
	assert.Equal(t, "settle", auditLogs[0].Action)
	assert.Equal(t, string(models.CommissionStatusPending), auditLogs[0].OldStatus)
	assert.Equal(t, string(models.CommissionStatusSettled), auditLogs[0].NewStatus)
}

func TestCommissionService_SettleCommission_CommissionNotFound(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	service := NewCommissionService()

	err := service.SettleCommission(999, 1)

	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	var auditLogCount int64
	err = database.DB.Model(&models.AuditLog{}).Count(&auditLogCount).Error
	require.NoError(t, err)
	assert.Zero(t, auditLogCount)
}

func TestCommissionService_SettleCommission_StatusNotPending(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	service := NewCommissionService()
	initialBalance := int64(3000)
	initialTotalCommission := int64(7000)
	commissionAmount := int64(2200)

	distributor, commission, operator := createSettlementFixture(t, models.CommissionStatusCancelled, initialBalance, initialTotalCommission, commissionAmount)

	err := service.SettleCommission(commission.ID, operator.ID)

	assert.EqualError(t, err, "只有待结算状态的佣金可以结算")

	var unchangedCommission models.CommissionRecord
	err = database.DB.First(&unchangedCommission, commission.ID).Error
	require.NoError(t, err)
	assert.Equal(t, models.CommissionStatusCancelled, unchangedCommission.Status)
	assert.Nil(t, unchangedCommission.SettledAt)

	var unchangedDistributor models.Distributor
	err = database.DB.First(&unchangedDistributor, distributor.ID).Error
	require.NoError(t, err)
	assert.Equal(t, initialBalance, unchangedDistributor.Balance)
	assert.Equal(t, initialTotalCommission, unchangedDistributor.TotalCommission)

	var auditLogCount int64
	err = database.DB.Model(&models.AuditLog{}).Count(&auditLogCount).Error
	require.NoError(t, err)
	assert.Zero(t, auditLogCount)
}

func createSettlementFixture(t *testing.T, commissionStatus models.CommissionStatus, initialBalance int64, initialTotalCommission int64, commissionAmount int64) (*models.Distributor, *models.CommissionRecord, *models.User) {
	t.Helper()

	distributorUser := &models.User{
		Username: fmt.Sprintf("%s_distributor", t.Name()),
		Password: "password123",
		Role:     models.RoleDistributor,
	}
	require.NoError(t, database.DB.Create(distributorUser).Error)

	operator := &models.User{
		Username: fmt.Sprintf("%s_finance", t.Name()),
		Password: "password123",
		Role:     models.RoleFinance,
	}
	require.NoError(t, database.DB.Create(operator).Error)

	distributor := &models.Distributor{
		UserID:          distributorUser.ID,
		RealName:        "测试分销员",
		Phone:           "13800138000",
		BankCardNo:      "6222021234567890",
		BankName:        "工商银行",
		Balance:         initialBalance,
		TotalCommission: initialTotalCommission,
		Status:          1,
	}
	require.NoError(t, database.DB.Create(distributor).Error)

	commission := &models.CommissionRecord{
		OrderNo:       fmt.Sprintf("%s_order", t.Name()),
		DistributorID: distributor.ID,
		OrderAmount:   10000,
		Rate:          18,
		Amount:        commissionAmount,
		Status:        commissionStatus,
	}
	require.NoError(t, database.DB.Create(commission).Error)

	return distributor, commission, operator
}
