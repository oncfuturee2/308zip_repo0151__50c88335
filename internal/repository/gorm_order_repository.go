package repository

import (
	"distribution-commission/internal/models"

	"gorm.io/gorm"
)

type GormOrderRepository struct {
	db *gorm.DB
}

func NewGormOrderRepository(db *gorm.DB) OrderRepository {
	return &GormOrderRepository{db: db}
}

func (r *GormOrderRepository) Create(order *models.Order) error {
	return r.db.Create(order).Error
}

func (r *GormOrderRepository) GetByOrderNo(orderNo string) (*models.Order, error) {
	var order models.Order
	if err := r.db.Where("order_no = ?", orderNo).Preload("Distributor").First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *GormOrderRepository) GetByID(id uint) (*models.Order, error) {
	var order models.Order
	if err := r.db.Preload("Distributor").First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *GormOrderRepository) ListByDistributor(distributorID uint, page, pageSize int) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	offset := (page - 1) * pageSize

	query := r.db.Model(&models.Order{}).Where("distributor_id = ?", distributorID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}
