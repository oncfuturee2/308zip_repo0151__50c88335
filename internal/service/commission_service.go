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
	idempotentService *IdempotentService
	orderService      *OrderService
}

func NewCommissionService() *CommissionService {
	return &CommissionService{
		idempotentService: NewIdempotentService(),
		orderService:      NewOrderService(),
	}
}

type commissionTier struct {
	distributorID uint
	level         int
	rate          int
}

func (s *CommissionService) GenerateCommission(ctx context.Context, orderNo string) ([]*models.CommissionRecord, error) {
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
		existingCommissions, err := s.findCommissionsByOrderNo(orderNo)
		if err != nil {
			return nil, err
		}
		if len(existingCommissions) > 0 {
			return existingCommissions, nil
		}
		return nil, errors.New("佣金正在生成中，请稍后重试")
	}
	defer s.idempotentService.ReleaseCommissionLock(ctx, orderNo)

	existingCommissions, err := s.findCommissionsByOrderNo(orderNo)
	if err != nil {
		return nil, err
	}
	if len(existingCommissions) > 0 {
		return existingCommissions, nil
	}

	tiers, err := s.resolveCommissionTiers(order.DistributorID)
	if err != nil {
		return nil, err
	}

	var commissions []*models.CommissionRecord
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		for _, tier := range tiers {
			amount := int64(float64(order.Amount) * float64(tier.rate) / 100.0)
			if amount <= 0 {
				continue
			}

			commission := &models.CommissionRecord{
				OrderNo:       orderNo,
				DistributorID: tier.distributorID,
				Level:         tier.level,
				OrderAmount:   order.Amount,
				Rate:          tier.rate,
				Amount:        amount,
				Status:        models.CommissionStatusPending,
			}

			if err := tx.Create(commission).Error; err != nil {
				return err
			}

			commissions = append(commissions, commission)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return commissions, nil
}

func (s *CommissionService) resolveCommissionTiers(distributorID uint) ([]commissionTier, error) {
	var tiers []commissionTier

	var distributor models.Distributor
	if err := database.DB.First(&distributor, distributorID).Error; err != nil {
		return nil, err
	}

	tiers = append(tiers, commissionTier{
		distributorID: distributor.ID,
		level:         models.CommissionLevelDirect,
		rate:          config.AppConfig.CommissionLevel1Rate,
	})

	if distributor.ParentID != nil && *distributor.ParentID > 0 {
		var parent models.Distributor
		if err := database.DB.First(&parent, *distributor.ParentID).Error; err == nil {
			tiers = append(tiers, commissionTier{
				distributorID: parent.ID,
				level:         models.CommissionLevelParent,
				rate:          config.AppConfig.CommissionLevel2Rate,
			})

			if parent.ParentID != nil && *parent.ParentID > 0 {
				var grandParent models.Distributor
				if err := database.DB.First(&grandParent, *parent.ParentID).Error; err == nil {
					tiers = append(tiers, commissionTier{
						distributorID: grandParent.ID,
						level:         models.CommissionLevelGrandParent,
						rate:          config.AppConfig.CommissionLevel3Rate,
					})
				}
			}
		}
	}

	return tiers, nil
}

func (s *CommissionService) findCommissionsByOrderNo(orderNo string) ([]*models.CommissionRecord, error) {
	var commissions []models.CommissionRecord
	if err := database.DB.Where("order_no = ?", orderNo).Find(&commissions).Error; err != nil {
		return nil, err
	}
	result := make([]*models.CommissionRecord, len(commissions))
	for i := range commissions {
		result[i] = &commissions[i]
	}
	return result, nil
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
