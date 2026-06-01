package repository

import (
	"distribution-commission/internal/models"

	"gorm.io/gorm"
)

// OrderRepository 订单数据访问接口
type OrderRepository interface {
	Create(order *models.Order) error
	FindByOrderNo(orderNo string) (*models.Order, error)
	GetByOrderNo(orderNo string) (*models.Order, error)
	GetByID(id uint) (*models.Order, error)
	ListByDistributor(distributorID uint, offset, limit int) ([]models.Order, int64, error)
}

type orderRepositoryImpl struct {
	db *gorm.DB
}

// NewOrderRepository 创建订单数据访问实例
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepositoryImpl{
		db: db,
	}
}

func (r *orderRepositoryImpl) Create(order *models.Order) error {
	return r.db.Create(order).Error
}

func (r *orderRepositoryImpl) FindByOrderNo(orderNo string) (*models.Order, error) {
	var order models.Order
	if err := r.db.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepositoryImpl) GetByOrderNo(orderNo string) (*models.Order, error) {
	var order models.Order
	if err := r.db.Where("order_no = ?", orderNo).Preload("Distributor").First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepositoryImpl) GetByID(id uint) (*models.Order, error) {
	var order models.Order
	if err := r.db.Preload("Distributor").First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepositoryImpl) ListByDistributor(distributorID uint, offset, limit int) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.db.Model(&models.Order{}).Where("distributor_id = ?", distributorID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}
