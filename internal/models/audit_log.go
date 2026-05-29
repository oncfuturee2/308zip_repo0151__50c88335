package models

import (
	"time"

	"gorm.io/gorm"
)

type AuditLog struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	DistributorID uint           `json:"distributor_id" gorm:"index;comment:分销员ID"`
	OperatorID    uint           `json:"operator_id" gorm:"index;comment:操作人ID"`
	Operator      *User          `json:"operator,omitempty" gorm:"foreignKey:OperatorID"`
	ResourceType  string         `json:"resource_type" gorm:"size:50;not null;comment:资源类型: withdraw/commission"`
	ResourceID    uint           `json:"resource_id" gorm:"index;not null;comment:资源ID"`
	Action        string         `json:"action" gorm:"size:50;not null;comment:操作类型"`
	OldStatus     string         `json:"old_status" gorm:"size:50;comment:原状态"`
	NewStatus     string         `json:"new_status" gorm:"size:50;comment:新状态"`
	ChangeReason  string         `json:"change_reason" gorm:"size:500;comment:变更原因"`
	CreatedAt     time.Time      `json:"created_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
