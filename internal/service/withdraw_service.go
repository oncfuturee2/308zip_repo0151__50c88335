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

	var validatedDistributor models.Distributor
	if err := database.DB.Select("id", "bank_card_no", "bank_name", "real_name").First(&validatedDistributor, distributorID).Error; err != nil {
		return nil, err
	}

	if validatedDistributor.BankCardNo == "" || validatedDistributor.RealName == "" {
		return nil, errors.New("请先完善银行卡信息")
	}

	valid, err := s.payoutProvider.ValidateBankCard(validatedDistributor.BankCardNo, validatedDistributor.BankName, validatedDistributor.RealName)
	if err != nil {
		return nil, errors.New("银行卡校验失败: " + err.Error())
	}
	if !valid {
		return nil, errors.New("银行卡信息校验不通过")
	}

	var withdraw *models.WithdrawRequest

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		updateResult := tx.Model(&models.Distributor{}).Where("id = ? AND balance >= ?", distributorID, amount).Updates(map[string]interface{}{
			"balance":        gorm.Expr("balance - ?", amount),
			"frozen_balance": gorm.Expr("frozen_balance + ?", amount),
		})
		if updateResult.Error != nil {
			return updateResult.Error
		}
		if updateResult.RowsAffected == 0 {
			return errors.New("余额不足")
		}

		var currentDistributor models.Distributor
		if err := tx.Select("id", "bank_card_no", "bank_name", "real_name").First(&currentDistributor, distributorID).Error; err != nil {
			return err
		}

		if currentDistributor.BankCardNo == "" || currentDistributor.RealName == "" {
			return errors.New("请先完善银行卡信息")
		}

		if currentDistributor.BankCardNo != validatedDistributor.BankCardNo || currentDistributor.BankName != validatedDistributor.BankName || currentDistributor.RealName != validatedDistributor.RealName {
			return errors.New("银行卡信息已变更，请重新提交提现申请")
		}

		withdraw = &models.WithdrawRequest{
			RequestNo:     utils.GenerateWithdrawRequestNo(),
			DistributorID: distributorID,
			Amount:        amount,
			BankCardNo:    currentDistributor.BankCardNo,
			BankName:      currentDistributor.BankName,
			RealName:      currentDistributor.RealName,
			Status:        models.WithdrawStatusPending,
		}

		if err := tx.Create(withdraw).Error; err != nil {
			return err
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
