package service

import (
	"testing"

	"distribution-commission/internal/database"
	"distribution-commission/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestCommissionService_SettleCommission_Success(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	userService := NewUserService()
	commissionService := NewCommissionService()

	user, err := userService.CreateUser("testuser", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	distributor := &models.Distributor{
		UserID:          user.ID,
		Balance:         0,
		TotalCommission: 0,
		Status:          1,
	}
	err = database.DB.Create(distributor).Error
	assert.NoError(t, err)

	commission := &models.CommissionRecord{
		OrderNo:       "ORD001",
		DistributorID: distributor.ID,
		OrderAmount:   10000,
		Rate:          10,
		Amount:        1000,
		Status:        models.CommissionStatusPending,
	}
	err = database.DB.Create(commission).Error
	assert.NoError(t, err)

	operatorID := uint(1)
	err = commissionService.SettleCommission(commission.ID, operatorID)
	assert.NoError(t, err)

	var updatedCommission models.CommissionRecord
	err = database.DB.First(&updatedCommission, commission.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, models.CommissionStatusSettled, updatedCommission.Status)
	assert.NotNil(t, updatedCommission.SettledAt)

	var updatedDistributor models.Distributor
	err = database.DB.First(&updatedDistributor, distributor.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1000), updatedDistributor.Balance)
	assert.Equal(t, int64(1000), updatedDistributor.TotalCommission)

	var auditLog models.AuditLog
	err = database.DB.Where("resource_type = ? AND resource_id = ? AND action = ?", "commission", commission.ID, "settle").First(&auditLog).Error
	assert.NoError(t, err)
	assert.Equal(t, distributor.ID, auditLog.DistributorID)
	assert.Equal(t, operatorID, auditLog.OperatorID)
	assert.Equal(t, string(models.CommissionStatusPending), auditLog.OldStatus)
	assert.Equal(t, string(models.CommissionStatusSettled), auditLog.NewStatus)
}

func TestCommissionService_SettleCommission_CommissionNotFound(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	commissionService := NewCommissionService()

	operatorID := uint(1)
	err := commissionService.SettleCommission(999, operatorID)
	assert.Error(t, err)
}

func TestCommissionService_SettleCommission_NotPendingStatus(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	userService := NewUserService()
	commissionService := NewCommissionService()

	user, err := userService.CreateUser("testuser", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	distributor := &models.Distributor{
		UserID:          user.ID,
		Balance:         0,
		TotalCommission: 0,
		Status:          1,
	}
	err = database.DB.Create(distributor).Error
	assert.NoError(t, err)

	commission := &models.CommissionRecord{
		OrderNo:       "ORD001",
		DistributorID: distributor.ID,
		OrderAmount:   10000,
		Rate:          10,
		Amount:        1000,
		Status:        models.CommissionStatusSettled,
	}
	err = database.DB.Create(commission).Error
	assert.NoError(t, err)

	operatorID := uint(1)
	err = commissionService.SettleCommission(commission.ID, operatorID)
	assert.Error(t, err)
	assert.Equal(t, "只有待结算状态的佣金可以结算", err.Error())

	var unchangedDistributor models.Distributor
	err = database.DB.First(&unchangedDistributor, distributor.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), unchangedDistributor.Balance)
	assert.Equal(t, int64(0), unchangedDistributor.TotalCommission)

	var auditLogCount int64
	database.DB.Model(&models.AuditLog{}).Count(&auditLogCount)
	assert.Equal(t, int64(0), auditLogCount)
}
