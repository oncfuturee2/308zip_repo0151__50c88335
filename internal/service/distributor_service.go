package service

import (
	"distribution-commission/internal/database"
	"distribution-commission/internal/models"

	"gorm.io/gorm"
)

type DistributorService struct{}

func NewDistributorService() *DistributorService {
	return &DistributorService{}
}

func (s *DistributorService) CreateDistributor(userID uint, realName, phone, bankCardNo, bankName string) (*models.Distributor, error) {
	return s.CreateDistributorWithParent(userID, realName, phone, bankCardNo, bankName, nil)
}

func (s *DistributorService) CreateDistributorWithParent(userID uint, realName, phone, bankCardNo, bankName string, parentID *uint) (*models.Distributor, error) {
	var existing models.Distributor
	if err := database.DB.Where("user_id = ?", userID).First(&existing).Error; err == nil {
		return nil, gorm.ErrRecordNotFound
	}

	distributor := &models.Distributor{
		UserID:     userID,
		RealName:   realName,
		Phone:      phone,
		BankCardNo: bankCardNo,
		BankName:   bankName,
		Status:     1,
		ParentID:   parentID,
	}

	if err := database.DB.Create(distributor).Error; err != nil {
		return nil, err
	}

	return distributor, nil
}

func (s *DistributorService) GetByUserID(userID uint) (*models.Distributor, error) {
	var distributor models.Distributor
	if err := database.DB.Where("user_id = ?", userID).Preload("User").First(&distributor).Error; err != nil {
		return nil, err
	}
	return &distributor, nil
}

func (s *DistributorService) GetByID(id uint) (*models.Distributor, error) {
	var distributor models.Distributor
	if err := database.DB.Preload("User").First(&distributor, id).Error; err != nil {
		return nil, err
	}
	return &distributor, nil
}

func (s *DistributorService) GetBalance(distributorID uint) (int64, int64, int64, error) {
	var distributor models.Distributor
	if err := database.DB.Select("balance", "frozen_balance", "total_commission").First(&distributor, distributorID).Error; err != nil {
		return 0, 0, 0, err
	}
	return distributor.Balance, distributor.FrozenBalance, distributor.TotalCommission, nil
}

func (s *DistributorService) UpdateBankInfo(distributorID uint, bankCardNo, bankName, realName string) error {
	return database.DB.Model(&models.Distributor{}).Where("id = ?", distributorID).Updates(map[string]interface{}{
		"bank_card_no": bankCardNo,
		"bank_name":    bankName,
		"real_name":    realName,
	}).Error
}
