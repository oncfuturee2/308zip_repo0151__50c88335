package service

import (
	"errors"
	"testing"

	"distribution-commission/internal/database"
	"distribution-commission/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestSettleCommission_Success(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	userService := NewUserService()
	distributorService := NewDistributorService()
	commissionService := NewCommissionService()

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

	initialBalance := distributor.Balance
	initialTotalCommission := distributor.TotalCommission

	commission := &models.CommissionRecord{
		OrderNo:       "TEST_ORDER_001",
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
	assert.Equal(t, initialBalance+1000, updatedDistributor.Balance)
	assert.Equal(t, initialTotalCommission+1000, updatedDistributor.TotalCommission)

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
}

func TestSettleCommission_CommissionNotFound(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	commissionService := NewCommissionService()

	err := commissionService.SettleCommission(999, 1)
	assert.Error(t, err)
}

func TestSettleCommission_StatusNotPending(t *testing.T) {
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

	commission := &models.CommissionRecord{
		OrderNo:       "TEST_ORDER_002",
		DistributorID: distributor.ID,
		OrderAmount:   10000,
		Rate:          10,
		Amount:        1000,
		Status:        models.CommissionStatusSettled,
	}
	err = database.DB.Create(commission).Error
	assert.NoError(t, err)

	commissionService := NewCommissionService()
	err = commissionService.SettleCommission(commission.ID, 1)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, errors.New("只有待结算状态的佣金可以结算")) || err.Error() == "只有待结算状态的佣金可以结算")
}

func TestSettleCommission_CancelledStatus(t *testing.T) {
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

	commission := &models.CommissionRecord{
		OrderNo:       "TEST_ORDER_003",
		DistributorID: distributor.ID,
		OrderAmount:   10000,
		Rate:          10,
		Amount:        1000,
		Status:        models.CommissionStatusCancelled,
	}
	err = database.DB.Create(commission).Error
	assert.NoError(t, err)

	commissionService := NewCommissionService()
	err = commissionService.SettleCommission(commission.ID, 1)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, errors.New("只有待结算状态的佣金可以结算")) || err.Error() == "只有待结算状态的佣金可以结算")
}

func TestSettleCommission_DoubleSettleBlocked(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	userService := NewUserService()
	distributorService := NewDistributorService()
	commissionService := NewCommissionService()

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
		OrderNo:       "TEST_ORDER_004",
		DistributorID: distributor.ID,
		OrderAmount:   10000,
		Rate:          10,
		Amount:        1000,
		Status:        models.CommissionStatusPending,
	}
	err = database.DB.Create(commission).Error
	assert.NoError(t, err)

	err = commissionService.SettleCommission(commission.ID, 1)
	assert.NoError(t, err)

	err = commissionService.SettleCommission(commission.ID, 2)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, errors.New("只有待结算状态的佣金可以结算")) || err.Error() == "只有待结算状态的佣金可以结算")

	var updatedDistributor models.Distributor
	err = database.DB.First(&updatedDistributor, distributor.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1000), updatedDistributor.Balance)
	assert.Equal(t, int64(1000), updatedDistributor.TotalCommission)

	var auditLogCount int64
	database.DB.Model(&models.AuditLog{}).Where("resource_type = ? AND resource_id = ?", "commission", commission.ID).Count(&auditLogCount)
	assert.Equal(t, int64(1), auditLogCount)
}