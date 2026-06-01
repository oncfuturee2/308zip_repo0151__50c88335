package repository

import (
	"context"

	"distribution-commission/internal/models"
)

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	FindByOrderNo(ctx context.Context, orderNo string) (*models.Order, error)
	FindByOrderNoWithDistributor(ctx context.Context, orderNo string) (*models.Order, error)
	FindByIDWithDistributor(ctx context.Context, id uint) (*models.Order, error)
	ListByDistributorID(ctx context.Context, distributorID uint, offset, limit int) ([]models.Order, int64, error)
}
