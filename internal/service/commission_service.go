package service

import (
	"context"
	"errors"
	"time"

	"distribution-commission/internal/config"
	"distribution-commission/internal/database"
	"distribution-commission/internal/models"
	"distribution-commission/internal/repository"

	"gorm.io/gorm"
)

type CommissionService struct {
	idempotentService *IdempotentService
	orderService      *OrderService
}

func NewCommissionService(orderRepo repository.OrderRepository) *CommissionService {
	return &CommissionService{
		idempotentService: NewIdempotentService(),
		orderService:      NewOrderService(orderRepo),
	}
}

func (s *CommissionService) GenerateCommission(ctx context.Context, orderNo string) (*models.CommissionRecord, error) {
	order, err := s.orderService.GetByOrderNo(orderNo)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("订单不存在")
		}
		return nil, err
	}

	if order.Status != models.OrderStatusCompleted {
		return nil, errors.New("订单未完成，不能生成佣金")
	}

	locked, err := s.idempotentService.AcquireCommissionLock(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	if !locked {
		var existingCommission models.CommissionRecord
		if err := database.DB.Where("order_no = ?", orderNo).First(&existingCommission).Error; err == nil {
			return &existingCommission, nil
		}
		return nil, errors.New("佣金正在生成中，请稍后重试")
	}
	defer s.idempotentService.ReleaseCommissionLock(ctx, orderNo)

	var existingCommission models.CommissionRecord
	if err := database.DB.Where("order_no = ?", orderNo).First(&existingCommission).Error; err == nil {
		return &existingCommission, nil
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	rate := config.AppConfig.CommissionRate
	amount := int64(float64(order.Amount) * float64(rate) / 100.0)

	commission := &models.CommissionRecord{
		OrderNo:       orderNo,
		DistributorID: order.DistributorID,
		OrderAmount:   order.Amount,
		Rate:          rate,
		Amount:        amount,
		Status:        models.CommissionStatusPending,
	}

	if err := database.DB.Create(commission).Error; err != nil {
		return nil, err
	}

	return commission, nil
}

func (s *CommissionService) SettleCommission(commissionID uint, operatorID uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var commission models.CommissionRecord
		if err := tx.First(&commission, commissionID).Error; err != nil {
			return err
		}

		if commission.Status != models.CommissionStatusPending {
			return errors.New("只有待结算状态的佣金可以结算")
		}

		now := time.Now()
		commission.Status = models.CommissionStatusSettled
		commission.SettledAt = &now

		if err := tx.Save(&commission).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.Distributor{}).Where("id = ?", commission.DistributorID).Updates(map[string]interface{}{
			"balance":          gorm.Expr("balance + ?", commission.Amount),
			"total_commission": gorm.Expr("total_commission + ?", commission.Amount),
		}).Error; err != nil {
			return err
		}

		auditLog := &models.AuditLog{
			DistributorID: commission.DistributorID,
			OperatorID:    operatorID,
			ResourceType:  "commission",
			ResourceID:    commission.ID,
			Action:        "settle",
			OldStatus:     string(models.CommissionStatusPending),
			NewStatus:     string(models.CommissionStatusSettled),
		}
		if err := tx.Create(auditLog).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *CommissionService) GetByID(id uint) (*models.CommissionRecord, error) {
	var commission models.CommissionRecord
	if err := database.DB.Preload("Distributor").First(&commission, id).Error; err != nil {
		return nil, err
	}
	return &commission, nil
}

func (s *CommissionService) ListByDistributor(distributorID uint, status string, page, pageSize int) ([]models.CommissionRecord, int64, error) {
	var commissions []models.CommissionRecord
	var total int64

	offset := (page - 1) * pageSize

	query := database.DB.Model(&models.CommissionRecord{}).Where("distributor_id = ?", distributorID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&commissions).Error; err != nil {
		return nil, 0, err
	}

	return commissions, total, nil
}
