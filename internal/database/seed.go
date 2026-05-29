package database

import (
	"log"

	"distribution-commission/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedData() {
	seedFinanceUser()
	seedDistributorUser()
}

func seedFinanceUser() {
	var count int64
	DB.Model(&models.User{}).Where("role = ?", models.RoleFinance).Count(&count)
	if count > 0 {
		log.Println("Finance user already exists, skipping seed")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("finance123"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		return
	}

	financeUser := &models.User{
		Username: "finance",
		Password: string(hashedPassword),
		Role:     models.RoleFinance,
	}

	if err := DB.Create(financeUser).Error; err != nil {
		log.Printf("Failed to create finance user: %v", err)
		return
	}

	log.Println("Finance user created: username=finance, password=finance123")
}

func seedDistributorUser() {
	var count int64
	DB.Model(&models.User{}).Where("username = ?", "distributor1").Count(&count)
	if count > 0 {
		log.Println("Distributor user already exists, skipping seed")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("dist123"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		return
	}

	err = DB.Transaction(func(tx *gorm.DB) error {
		distUser := &models.User{
			Username: "distributor1",
			Password: string(hashedPassword),
			Role:     models.RoleDistributor,
		}
		if err := tx.Create(distUser).Error; err != nil {
			return err
		}

		distributor := &models.Distributor{
			UserID:      distUser.ID,
			RealName:    "张三",
			Phone:       "13800138000",
			BankCardNo:  "6222021234567890123",
			BankName:    "中国工商银行",
			Status:      1,
		}
		if err := tx.Create(distributor).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Printf("Failed to create distributor user: %v", err)
		return
	}

	log.Println("Distributor user created: username=distributor1, password=dist123")
}
