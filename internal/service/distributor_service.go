package service

import (
	"errors"

	"distribution-commission/internal/database"
	"distribution-commission/internal/models"

	"gorm.io/gorm"
)

type DistributorService struct{}

func NewDistributorService() *DistributorService {
	return &DistributorService{}
}

func (s *DistributorService) CreateDistributor(userID uint, parentDistributorID uint, realName, phone, bankCardNo, bankName string) (*models.Distributor, error) {
	var existing models.Distributor
	if err := database.DB.Where("user_id = ?", userID).First(&existing).Error; err == nil {
		return nil, errors.New("分销员已存在")
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	var parentID *uint
	if parentDistributorID != 0 {
		var parent models.Distributor
		if err := database.DB.First(&parent, parentDistributorID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, errors.New("上级分销员不存在")
			}
			return nil, err
		}
		parentID = &parent.ID
	}

	distributor := &models.Distributor{
		UserID:     userID,
		ParentID:   parentID,
		RealName:   realName,
		Phone:      phone,
		BankCardNo: bankCardNo,
		BankName:   bankName,
		Status:     1,
	}

	if err := database.DB.Create(distributor).Error; err != nil {
		return nil, err
	}

	return distributor, nil
}

func (s *DistributorService) GetByUserID(userID uint) (*models.Distributor, error) {
	var distributor models.Distributor
	if err := database.DB.Where("user_id = ?", userID).Preload("User").Preload("Parent").First(&distributor).Error; err != nil {
		return nil, err
	}
	return &distributor, nil
}

func (s *DistributorService) GetByID(id uint) (*models.Distributor, error) {
	var distributor models.Distributor
	if err := database.DB.Preload("User").Preload("Parent").First(&distributor, id).Error; err != nil {
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

func (s *DistributorService) GetUplineChainTx(tx *gorm.DB, distributorID uint, maxDepth int) ([]models.Distributor, error) {
	if tx == nil {
		tx = database.DB
	}

	chain := make([]models.Distributor, 0, maxDepth)
	visited := make(map[uint]struct{}, maxDepth)
	currentID := distributorID

	for len(chain) < maxDepth && currentID != 0 {
		var distributor models.Distributor
		if err := tx.First(&distributor, currentID).Error; err != nil {
			return nil, err
		}

		if _, exists := visited[distributor.ID]; exists {
			return nil, errors.New("分销层级关系存在循环")
		}

		visited[distributor.ID] = struct{}{}
		chain = append(chain, distributor)

		if distributor.ParentID == nil {
			break
		}

		currentID = *distributor.ParentID
	}

	return chain, nil
}
