package service

import (
	"context"
	"errors"

	"distribution-commission/internal/database"
	"distribution-commission/internal/models"
	"distribution-commission/internal/pkg/utils"

	"gorm.io/gorm"
)

type OrderService struct {
	idempotentService *IdempotentService
}

func NewOrderService() *OrderService {
	return &OrderService{
		idempotentService: NewIdempotentService(),
	}
}

func (s *OrderService) CreateCompletedOrder(ctx context.Context, orderNo string, distributorID uint, amount int64, goodsName string) (*models.Order, error) {
	if orderNo == "" {
		orderNo = utils.GenerateOrderNo()
	}

	locked, err := s.idempotentService.AcquireOrderLock(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	if !locked {
		var existingOrder models.Order
		if err := database.DB.Where("order_no = ?", orderNo).First(&existingOrder).Error; err == nil {
			return &existingOrder, nil
		}
		return nil, errors.New("订单正在处理中，请稍后重试")
	}
	defer s.idempotentService.ReleaseOrderLock(ctx, orderNo)

	var existingOrder models.Order
	if err := database.DB.Where("order_no = ?", orderNo).First(&existingOrder).Error; err == nil {
		return &existingOrder, nil
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	order := &models.Order{
		OrderNo:       orderNo,
		DistributorID: distributorID,
		Amount:        amount,
		GoodsName:     goodsName,
		Status:        models.OrderStatusCompleted,
	}

	if err := database.DB.Create(order).Error; err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderService) GetByOrderNo(orderNo string) (*models.Order, error) {
	var order models.Order
	if err := database.DB.Where("order_no = ?", orderNo).Preload("Distributor").First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (s *OrderService) GetByID(id uint) (*models.Order, error) {
	var order models.Order
	if err := database.DB.Preload("Distributor").First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (s *OrderService) ListByDistributor(distributorID uint, page, pageSize int) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	offset := (page - 1) * pageSize

	query := database.DB.Model(&models.Order{}).Where("distributor_id = ?", distributorID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (s *OrderService) RefundOrder(ctx context.Context, orderNo string) (*models.Order, error) {
	order, err := s.GetByOrderNo(orderNo)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("订单不存在")
		}
		return nil, err
	}

	if order.Status == models.OrderStatusRefunded {
		return order, nil
	}

	if order.Status != models.OrderStatusCompleted {
		return nil, errors.New("只有已完成状态的订单可以退款")
	}

	locked, err := s.idempotentService.AcquireRefundLock(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	if !locked {
		var existingOrder models.Order
		if err := database.DB.Where("order_no = ?", orderNo).First(&existingOrder).Error; err == nil {
			if existingOrder.Status == models.OrderStatusRefunded {
				return &existingOrder, nil
			}
		}
		return nil, errors.New("订单退款正在处理中，请稍后重试")
	}
	defer s.idempotentService.ReleaseRefundLock(ctx, orderNo)

	var refreshedOrder models.Order
	if err := database.DB.Where("order_no = ?", orderNo).First(&refreshedOrder).Error; err != nil {
		return nil, err
	}

	if refreshedOrder.Status == models.OrderStatusRefunded {
		return &refreshedOrder, nil
	}

	if refreshedOrder.Status != models.OrderStatusCompleted {
		return nil, errors.New("只有已完成状态的订单可以退款")
	}

	refreshedOrder.Status = models.OrderStatusRefunded
	if err := database.DB.Save(&refreshedOrder).Error; err != nil {
		return nil, err
	}

	return &refreshedOrder, nil
}
