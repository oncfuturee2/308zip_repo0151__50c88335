package service

import (
	"errors"

	"distribution-commission/internal/database"
	"distribution-commission/internal/models"
	"distribution-commission/internal/pkg/utils"

	"gorm.io/gorm"
)

type WithdrawService struct {
	payoutProvider PayoutProvider
}

func NewWithdrawService() *WithdrawService {
	return &WithdrawService{
		payoutProvider: NewMockPayoutProvider(),
	}
}

func (s *WithdrawService) SubmitWithdraw(distributorID uint, amount int64) (*models.WithdrawRequest, error) {
	if amount <= 0 {
		return nil, errors.New("提现金额必须大于0")
	}

	var distributor models.Distributor
	if err := database.DB.First(&distributor, distributorID).Error; err != nil {
		return nil, err
	}

	if distributor.Balance < amount {
		return nil, errors.New("余额不足")
	}

	if distributor.BankCardNo == "" || distributor.RealName == "" {
		return nil, errors.New("请先完善银行卡信息")
	}

	valid, err := s.payoutProvider.ValidateBankCard(distributor.BankCardNo, distributor.BankName, distributor.RealName)
	if err != nil {
		return nil, errors.New("银行卡校验失败: " + err.Error())
	}
	if !valid {
		return nil, errors.New("银行卡信息校验不通过")
	}

	var withdraw *models.WithdrawRequest

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		requestNo := utils.GenerateWithdrawRequestNo()

		withdraw = &models.WithdrawRequest{
			RequestNo:     requestNo,
			DistributorID: distributorID,
			Amount:        amount,
			BankCardNo:    distributor.BankCardNo,
			BankName:      distributor.BankName,
			RealName:      distributor.RealName,
			Status:        models.WithdrawStatusPending,
		}

		if err := tx.Create(withdraw).Error; err != nil {
			return err
		}

		// 使用基于查询条件的乐观锁策略保证并发下的余额绝对安全
		res := tx.Model(&models.Distributor{}).Where("id = ? AND balance >= ?", distributorID, amount).Updates(map[string]interface{}{
			"balance":        gorm.Expr("balance - ?", amount),
			"frozen_balance": gorm.Expr("frozen_balance + ?", amount),
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("并发操作导致余额不足")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return withdraw, nil
}

func (s *WithdrawService) Approve(withdrawID uint, operatorID uint) error {
	var withdraw models.WithdrawRequest
	if err := database.DB.First(&withdraw, withdrawID).Error; err != nil {
		return err
	}

	if withdraw.Status != models.WithdrawStatusPending {
		return errors.New("只有待审核状态的提现可以通过")
	}

	err := withdraw.TransitionTo(database.DB, models.WithdrawStatusApproved, operatorID, "")
	if err != nil {
		return err
	}

	go s.processPayout(withdrawID, operatorID)

	return nil
}

func (s *WithdrawService) Reject(withdrawID uint, operatorID uint, reason string) error {
	var withdraw models.WithdrawRequest
	if err := database.DB.First(&withdraw, withdrawID).Error; err != nil {
		return err
	}

	if withdraw.Status != models.WithdrawStatusPending {
		return errors.New("只有待审核状态的提现可以驳回")
	}

	if reason == "" {
		return errors.New("驳回原因不能为空")
	}

	return withdraw.TransitionTo(database.DB, models.WithdrawStatusRejected, operatorID, reason)
}

func (s *WithdrawService) processPayout(withdrawID uint, operatorID uint) {
	var withdraw models.WithdrawRequest
	if err := database.DB.First(&withdraw, withdrawID).Error; err != nil {
		return
	}

	if withdraw.Status != models.WithdrawStatusApproved {
		return
	}

	_, err := s.payoutProvider.Payout(&withdraw)
	if err != nil {
		return
	}

	_ = withdraw.TransitionTo(database.DB, models.WithdrawStatusPaidMock, operatorID, "mock打款成功")
}

func (s *WithdrawService) GetByID(id uint) (*models.WithdrawRequest, error) {
	var withdraw models.WithdrawRequest
	if err := database.DB.Preload("Distributor").First(&withdraw, id).Error; err != nil {
		return nil, err
	}
	return &withdraw, nil
}

func (s *WithdrawService) ListByDistributor(distributorID uint, status string, page, pageSize int) ([]models.WithdrawRequest, int64, error) {
	var withdraws []models.WithdrawRequest
	var total int64

	offset := (page - 1) * pageSize

	query := database.DB.Model(&models.WithdrawRequest{}).Where("distributor_id = ?", distributorID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&withdraws).Error; err != nil {
		return nil, 0, err
	}

	return withdraws, total, nil
}

func (s *WithdrawService) ListAll(status string, page, pageSize int) ([]models.WithdrawRequest, int64, error) {
	var withdraws []models.WithdrawRequest
	var total int64

	offset := (page - 1) * pageSize

	query := database.DB.Model(&models.WithdrawRequest{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Distributor").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&withdraws).Error; err != nil {
		return nil, 0, err
	}

	return withdraws, total, nil
}
