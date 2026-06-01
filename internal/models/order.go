package models

import (
	"time"

	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusRefunded  OrderStatus = "refunded"
)

type Order struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	OrderNo       string         `json:"order_no" gorm:"uniqueIndex;size:32;not null;comment:订单号"`
	DistributorID uint           `json:"distributor_id" gorm:"index;not null;comment:分销员ID"`
	Distributor   Distributor    `json:"distributor" gorm:"foreignKey:DistributorID"`
	Amount        int64          `json:"amount" gorm:"not null;comment:订单金额(分)"`
	GoodsName     string         `json:"goods_name" gorm:"size:255;comment:商品名称"`
	Status        OrderStatus    `json:"status" gorm:"size:20;not null;default:'pending'"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Order) TableName() string {
	return "orders"
}
