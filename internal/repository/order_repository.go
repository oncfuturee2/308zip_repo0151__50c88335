package repository

import "distribution-commission/internal/models"

type OrderRepository interface {
	Create(order *models.Order) error
	GetByOrderNo(orderNo string) (*models.Order, error)
	GetByID(id uint) (*models.Order, error)
	ListByDistributor(distributorID uint, page, pageSize int) ([]models.Order, int64, error)
}
