package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type WithdrawStatus string

const (
	WithdrawStatusPending   WithdrawStatus = "pending"
	WithdrawStatusApproved  WithdrawStatus = "approved"
	WithdrawStatusRejected  WithdrawStatus = "rejected"
	WithdrawStatusPaidMock  WithdrawStatus = "paid_mock"
)

var withdrawStateTransitions = map[WithdrawStatus][]WithdrawStatus{
	WithdrawStatusPending:  {WithdrawStatusApproved, WithdrawStatusRejected},
	WithdrawStatusApproved: {WithdrawStatusPaidMock},
	WithdrawStatusRejected: {},
	WithdrawStatusPaidMock: {},
}

func (s WithdrawStatus) CanTransitionTo(target WithdrawStatus) bool {
	validTransitions, exists := withdrawStateTransitions[s]
	if !exists {
		return false
	}
	for _, t := range validTransitions {
		if t == target {
			return true
		}
	}
	return false
}

type WithdrawRequest struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	RequestNo     string         `json:"request_no" gorm:"uniqueIndex;size:32;not null;comment:提现申请单号"`
	DistributorID uint           `json:"distributor_id" gorm:"index;not null;comment:分销员ID"`
	Distributor   Distributor    `json:"distributor" gorm:"foreignKey:DistributorID"`
	Amount        int64          `json:"amount" gorm:"not null;comment:提现金额(分)"`
	BankCardNo    string         `json:"bank_card_no" gorm:"size:30;comment:银行卡号"`
	BankName      string         `json:"bank_name" gorm:"size:50;comment:银行名称"`
	RealName      string         `json:"real_name" gorm:"size:50;comment:真实姓名"`
	Status        WithdrawStatus `json:"status" gorm:"size:20;not null;default:'pending'"`
	RejectReason  string         `json:"reject_reason" gorm:"size:255;comment:驳回原因"`
	PayoutAt      *time.Time     `json:"payout_at" gorm:"comment:打款时间"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

func (WithdrawRequest) TableName() string {
	return "withdraw_requests"
}

func (w *WithdrawRequest) TransitionTo(db *gorm.DB, target WithdrawStatus, operatorID uint, reason string) error {
	if !w.Status.CanTransitionTo(target) {
		return errors.New("invalid state transition")
	}

	oldStatus := w.Status
	w.Status = target

	if target == WithdrawStatusRejected {
		w.RejectReason = reason
	}

	if target == WithdrawStatusPaidMock {
		now := time.Now()
		w.PayoutAt = &now
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(w).Error; err != nil {
			return err
		}

		auditLog := &AuditLog{
			DistributorID:  w.DistributorID,
			OperatorID:     operatorID,
			ResourceType:   "withdraw",
			ResourceID:     w.ID,
			Action:         "status_change",
			OldStatus:      string(oldStatus),
			NewStatus:      string(target),
			ChangeReason:   reason,
		}
		if err := tx.Create(auditLog).Error; err != nil {
			return err
		}

		if target == WithdrawStatusRejected {
			if err := tx.Model(&Distributor{}).Where("id = ?", w.DistributorID).Updates(map[string]interface{}{
				"balance":        gorm.Expr("balance + ?", w.Amount),
				"frozen_balance": gorm.Expr("frozen_balance - ?", w.Amount),
			}).Error; err != nil {
				return err
			}
		}

		if target == WithdrawStatusPaidMock {
			if err := tx.Model(&Distributor{}).Where("id = ?", w.DistributorID).Update("frozen_balance", gorm.Expr("frozen_balance - ?", w.Amount)).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
