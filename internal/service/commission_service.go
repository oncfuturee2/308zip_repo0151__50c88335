package service

import (
	"context"
	"errors"
	"time"

	"distribution-commission/internal/config"
	"distribution-commission/internal/database"
	"distribution-commission/internal/models"

	"gorm.io/gorm"
)

type CommissionService struct {
	idempotentService  *IdempotentService
	orderService       *OrderService
	distributorService *DistributorService
}

func NewCommissionService() *CommissionService {
	return &CommissionService{
		idempotentService:  NewIdempotentService(),
		orderService:       NewOrderService(),
		distributorService: NewDistributorService(),
	}
}

func (s *CommissionService) GenerateCommission(ctx context.Context, orderNo string) ([]models.CommissionRecord, error) {
	locked, err := s.idempotentService.AcquireCommissionLock(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	if !locked {
		existingCommissions, err := s.findByOrderNo(database.DB, orderNo)
		if err != nil {
			return nil, err
		}
		if len(existingCommissions) > 0 {
			return existingCommissions, nil
		}
		return nil, errors.New("佣金正在生成中，请稍后重试")
	}
	defer s.idempotentService.ReleaseCommissionLock(ctx, orderNo)

	var commissions []models.CommissionRecord
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		order, err := s.orderService.getByOrderNoTx(tx, orderNo)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("订单不存在")
			}
			return err
		}

		if order.Status != models.OrderStatusCompleted {
			return errors.New("订单未完成，不能生成佣金")
		}

		commissions, err = s.findByOrderNo(tx, orderNo)
		if err != nil {
			return err
		}
		if len(commissions) > 0 {
			return nil
		}

		ratePlan := config.AppConfig.CommissionRates()
		chain, err := s.distributorService.GetUplineChainTx(tx, order.DistributorID, len(ratePlan))
		if err != nil {
			return err
		}

		newCommissions := make([]models.CommissionRecord, 0, len(chain))
		for index, distributor := range chain {
			rate := ratePlan[index]
			if rate <= 0 {
				continue
			}

			newCommissions = append(newCommissions, models.CommissionRecord{
				OrderNo:             order.OrderNo,
				DistributorID:       distributor.ID,
				SourceDistributorID: order.DistributorID,
				OrderAmount:         order.Amount,
				Level:               index + 1,
				Rate:                rate,
				Amount:              calculateCommissionAmount(order.Amount, rate),
				Status:              models.CommissionStatusPending,
			})
		}

		if len(newCommissions) == 0 {
			return errors.New("未配置有效佣金比例")
		}

		if err := tx.Create(&newCommissions).Error; err != nil {
			return err
		}

		commissions = newCommissions
		return nil
	}); err != nil {
		return nil, err
	}

	return commissions, nil
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
	if err := database.DB.Preload("Distributor").Preload("SourceDistributor").First(&commission, id).Error; err != nil {
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

	if err := query.Preload("SourceDistributor").Offset(offset).Limit(pageSize).Order("created_at DESC, level ASC").Find(&commissions).Error; err != nil {
		return nil, 0, err
	}

	return commissions, total, nil
}

func (s *CommissionService) findByOrderNo(db *gorm.DB, orderNo string) ([]models.CommissionRecord, error) {
	var commissions []models.CommissionRecord
	if err := db.Where("order_no = ?", orderNo).Preload("Distributor").Preload("SourceDistributor").Order("level ASC, id ASC").Find(&commissions).Error; err != nil {
		return nil, err
	}
	return commissions, nil
}

func calculateCommissionAmount(orderAmount int64, rate int) int64 {
	return orderAmount * int64(rate) / 100
}
