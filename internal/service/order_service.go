package service

import (
	"context"
	"errors"

	"distribution-commission/internal/models"
	"distribution-commission/internal/pkg/utils"
	"distribution-commission/internal/repository"

	"gorm.io/gorm"
)

type OrderService struct {
	orderRepo        repository.OrderRepository
	idempotentService *IdempotentService
}

func NewOrderService(orderRepo repository.OrderRepository) *OrderService {
	return &OrderService{
		orderRepo:        orderRepo,
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
		existingOrder, repoErr := s.orderRepo.FindByOrderNo(orderNo)
		if repoErr == nil {
			return existingOrder, nil
		}
		return nil, errors.New("订单正在处理中，请稍后重试")
	}
	defer s.idempotentService.ReleaseOrderLock(ctx, orderNo)

	existingOrder, repoErr := s.orderRepo.FindByOrderNo(orderNo)
	if repoErr == nil {
		return existingOrder, nil
	} else if !errors.Is(repoErr, gorm.ErrRecordNotFound) {
		return nil, repoErr
	}

	order := &models.Order{
		OrderNo:       orderNo,
		DistributorID: distributorID,
		Amount:        amount,
		GoodsName:     goodsName,
		Status:        models.OrderStatusCompleted,
	}

	if err := s.orderRepo.Create(order); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderService) GetByOrderNo(orderNo string) (*models.Order, error) {
	return s.orderRepo.FindByOrderNo(orderNo)
}

func (s *OrderService) GetByID(id uint) (*models.Order, error) {
	return s.orderRepo.FindByID(id)
}

func (s *OrderService) ListByDistributor(distributorID uint, page, pageSize int) ([]models.Order, int64, error) {
	offset := (page - 1) * pageSize
	return s.orderRepo.FindByDistributor(distributorID, offset, pageSize)
}