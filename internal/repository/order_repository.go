package repository

import (
	"distribution-commission/internal/models"
)

type OrderRepository interface {
	Create(order *models.Order) error
	FindByOrderNo(orderNo string) (*models.Order, error)
	FindByID(id uint) (*models.Order, error)
	FindByDistributor(distributorID uint, offset, limit int) ([]models.Order, int64, error)
}