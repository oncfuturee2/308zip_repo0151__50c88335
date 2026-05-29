package models

import (
	"time"

	"gorm.io/gorm"
)

type CommissionStatus string

const (
	CommissionStatusPending   CommissionStatus = "pending"
	CommissionStatusSettled   CommissionStatus = "settled"
	CommissionStatusCancelled CommissionStatus = "cancelled"
)

type CommissionRecord struct {
	ID            uint             `json:"id" gorm:"primaryKey"`
	OrderNo       string           `json:"order_no" gorm:"uniqueIndex:idx_order_commission;size:32;not null;comment:订单号"`
	DistributorID uint             `json:"distributor_id" gorm:"index;not null;comment:分销员ID"`
	Distributor   Distributor      `json:"distributor" gorm:"foreignKey:DistributorID"`
	OrderAmount   int64            `json:"order_amount" gorm:"not null;comment:订单金额(分)"`
	Rate          int              `json:"rate" gorm:"not null;comment:佣金比例(%)"`
	Amount        int64            `json:"amount" gorm:"not null;comment:佣金金额(分)"`
	Status        CommissionStatus `json:"status" gorm:"size:20;not null;default:'pending'"`
	SettledAt     *time.Time       `json:"settled_at" gorm:"comment:结算时间"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	DeletedAt     gorm.DeletedAt   `json:"-" gorm:"index"`
}

func (CommissionRecord) TableName() string {
	return "commission_records"
}
