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

	// Prepare data
	distributor := &models.Distributor{
		UserID:          1,
		Balance:         1000,
		TotalCommission: 2000,
	}
	err := database.DB.Create(distributor).Error
	assert.NoError(t, err)

	pendingCommission := &models.CommissionRecord{
		OrderNo:       "ORDER001",
		DistributorID: distributor.ID,
		OrderAmount:   10000,
		Rate:          10,
		Amount:        1000,
		Status:        models.CommissionStatusPending,
	}
	err = database.DB.Create(pendingCommission).Error
	assert.NoError(t, err)

	svc := NewCommissionService()
	operatorID := uint(100)
	err = svc.SettleCommission(pendingCommission.ID, operatorID)
	assert.NoError(t, err)

	// Assert commission status
	var updatedCommission models.CommissionRecord
	database.DB.First(&updatedCommission, pendingCommission.ID)
	assert.Equal(t, models.CommissionStatusSettled, updatedCommission.Status)
	assert.NotNil(t, updatedCommission.SettledAt)

	// Assert distributor balance and total_commission
	var updatedDistributor models.Distributor
	database.DB.First(&updatedDistributor, distributor.ID)
	assert.Equal(t, int64(2000), updatedDistributor.Balance)
	assert.Equal(t, int64(3000), updatedDistributor.TotalCommission)

	// Assert audit log
	var auditLog models.AuditLog
	err = database.DB.Where("resource_type = ? AND resource_id = ?", "commission", pendingCommission.ID).First(&auditLog).Error
	assert.NoError(t, err)
	assert.Equal(t, operatorID, auditLog.OperatorID)
	assert.Equal(t, distributor.ID, auditLog.DistributorID)
	assert.Equal(t, "settle", auditLog.Action)
	assert.Equal(t, string(models.CommissionStatusPending), auditLog.OldStatus)
	assert.Equal(t, string(models.CommissionStatusSettled), auditLog.NewStatus)
}

func TestCommissionService_SettleCommission_NotFound(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	svc := NewCommissionService()

	err := svc.SettleCommission(9999, 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record not found")
}

func TestCommissionService_SettleCommission_NotPending(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	// Prepare data
	distributor := &models.Distributor{
		UserID:          1,
		Balance:         1000,
		TotalCommission: 2000,
	}
	err := database.DB.Create(distributor).Error
	assert.NoError(t, err)

	settledCommission := &models.CommissionRecord{
		OrderNo:       "ORDER002",
		DistributorID: distributor.ID,
		OrderAmount:   10000,
		Rate:          10,
		Amount:        1000,
		Status:        models.CommissionStatusSettled,
	}
	err = database.DB.Create(settledCommission).Error
	assert.NoError(t, err)

	svc := NewCommissionService()
	operatorID := uint(100)
	
	err = svc.SettleCommission(settledCommission.ID, operatorID)
	assert.Error(t, err)
	assert.Equal(t, "只有待结算状态的佣金可以结算", err.Error())
}
