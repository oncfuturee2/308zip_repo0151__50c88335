package service

import (
	"errors"

	"distribution-commission/internal/database"
	"distribution-commission/internal/models"
	"distribution-commission/internal/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func (s *UserService) Login(username, password string) (string, *models.User, error) {
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}

	token, err := jwt.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return "", nil, err
	}

	return token, &user, nil
}

func (s *UserService) CreateUser(username, password string, role models.Role) (*models.User, error) {
	var count int64
	database.DB.Model(&models.User{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username: username,
		Password: string(hashedPassword),
		Role:     role,
	}

	if err := database.DB.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
