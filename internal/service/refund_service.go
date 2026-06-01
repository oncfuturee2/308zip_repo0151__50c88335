package service

import (
	"context"
	"errors"

	"distribution-commission/internal/models"
)

type RefundService struct {
	orderService      *OrderService
	commissionService *CommissionService
}

func NewRefundService() *RefundService {
	return &RefundService{
		orderService:      NewOrderService(),
		commissionService: NewCommissionService(),
	}
}

func (s *RefundService) RefundOrder(ctx context.Context, orderNo string, operatorID uint) (*models.Order, error) {
	order, err := s.orderService.RefundOrder(ctx, orderNo, operatorID)
	if err != nil {
		return nil, err
	}

	if err := s.commissionService.RollbackCommission(orderNo, operatorID); err != nil {
		return nil, errors.New("订单退款成功，但佣金回滚失败: " + err.Error())
	}

	return order, nil
}
