package models

import (
	"time"

	"gorm.io/gorm"
)

type Distributor struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	UserID          uint           `json:"user_id" gorm:"uniqueIndex;not null"`
	User            User           `json:"user" gorm:"foreignKey:UserID"`
	ParentID        *uint          `json:"parent_id" gorm:"index;comment:上级分销商ID"`
	Parent          *Distributor   `json:"parent" gorm:"foreignKey:ParentID"`
	Path            string         `json:"path" gorm:"index;size:255;comment:分销商层级路径，如 1,2,3"`
	RealName        string         `json:"real_name" gorm:"size:50"`
	Phone           string         `json:"phone" gorm:"size:20"`
	BankCardNo      string         `json:"bank_card_no" gorm:"size:30"`
	BankName        string         `json:"bank_name" gorm:"size:50"`
	Balance         int64          `json:"balance" gorm:"default:0;comment:可提现余额(分)"`
	FrozenBalance   int64          `json:"frozen_balance" gorm:"default:0;comment:冻结余额(分)"`
	TotalCommission int64          `json:"total_commission" gorm:"default:0;comment:累计佣金(分)"`
	Status          int            `json:"status" gorm:"default:1;comment:1-正常 0-禁用"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Distributor) TableName() string {
	return "distributors"
}
