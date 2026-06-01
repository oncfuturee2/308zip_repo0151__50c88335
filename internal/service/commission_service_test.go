package service

import (
	"testing"
	"time"

	"distribution-commission/internal/database"
	"distribution-commission/internal/models"

	"github.com/stretchr/testify/assert"
)

func setupTestCommission(t *testing.T, status models.CommissionStatus) (*models.User, *models.Distributor, *models.CommissionRecord) {
	t.Helper()

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

	commission := &models.CommissionRecord{
		OrderNo:       "ORD-20240101-001",
		DistributorID: distributor.ID,
		OrderAmount:   100000,
		Rate:          10,
		Amount:        10000,
		Status:        status,
	}
	if status == models.CommissionStatusSettled {
		now := time.Now()
		commission.SettledAt = &now
	}
	err = database.DB.Create(commission).Error
	assert.NoError(t, err)

	return user, distributor, commission
}

func TestCommissionService_SettleCommission_Success(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	commissionService := NewCommissionService()

	user, distributor, commission := setupTestCommission(t, models.CommissionStatusPending)

	distributorBefore := models.Distributor{}
	err := database.DB.First(&distributorBefore, distributor.ID).Error
	assert.NoError(t, err)

	operatorID := uint(999)

	err = commissionService.SettleCommission(commission.ID, operatorID)

	assert.NoError(t, err)

	var updatedCommission models.CommissionRecord
	err = database.DB.First(&updatedCommission, commission.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, models.CommissionStatusSettled, updatedCommission.Status)
	assert.NotNil(t, updatedCommission.SettledAt)
	assert.WithinDuration(t, time.Now(), *updatedCommission.SettledAt, 2*time.Second)

	var updatedDistributor models.Distributor
	err = database.DB.First(&updatedDistributor, distributor.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, distributorBefore.Balance+commission.Amount, updatedDistributor.Balance)
	assert.Equal(t, distributorBefore.TotalCommission+commission.Amount, updatedDistributor.TotalCommission)

	var auditLog models.AuditLog
	err = database.DB.Where("resource_type = ? AND resource_id = ?", "commission", commission.ID).First(&auditLog).Error
	assert.NoError(t, err)
	assert.Equal(t, distributor.ID, auditLog.DistributorID)
	assert.Equal(t, operatorID, auditLog.OperatorID)
	assert.Equal(t, "commission", auditLog.ResourceType)
	assert.Equal(t, commission.ID, auditLog.ResourceID)
	assert.Equal(t, "settle", auditLog.Action)
	assert.Equal(t, string(models.CommissionStatusPending), auditLog.OldStatus)
	assert.Equal(t, string(models.CommissionStatusSettled), auditLog.NewStatus)

	var auditCount int64
	database.DB.Model(&models.AuditLog{}).Count(&auditCount)
	assert.Equal(t, int64(1), auditCount)

	_ = user
}

func TestCommissionService_SettleCommission_NotFound(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	commissionService := NewCommissionService()

	err := commissionService.SettleCommission(999, 1)

	assert.Error(t, err)
}

func TestCommissionService_SettleCommission_NotPendingStatus(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	commissionService := NewCommissionService()

	_, _, commission := setupTestCommission(t, models.CommissionStatusSettled)

	err := commissionService.SettleCommission(commission.ID, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "只有待结算状态的佣金可以结算")
}

func TestCommissionService_SettleCommission_CancelledStatus(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	commissionService := NewCommissionService()

	_, _, commission := setupTestCommission(t, models.CommissionStatusCancelled)

	err := commissionService.SettleCommission(commission.ID, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "只有待结算状态的佣金可以结算")
}

