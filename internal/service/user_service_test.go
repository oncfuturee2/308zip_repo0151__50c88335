package service

import (
	"os"
	"testing"

	"distribution-commission/internal/config"
	"distribution-commission/internal/database"
	"distribution-commission/internal/models"

	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("JWT_SECRET", "test_secret_key")
	os.Setenv("JWT_EXPIRE_HOURS", "24")
	config.LoadConfig()
}

func TestUserService_CreateUser_Success(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	service := NewUserService()

	user, err := service.CreateUser("testuser", "password123", models.RoleDistributor)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, models.RoleDistributor, user.Role)
	assert.NotEmpty(t, user.Password)
	assert.Greater(t, user.ID, uint(0))

	var count int64
	database.DB.Model(&models.User{}).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestUserService_CreateUser_DuplicateUsername(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	service := NewUserService()

	_, err := service.CreateUser("testuser", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	_, err = service.CreateUser("testuser", "anotherpassword", models.RoleDistributor)

	assert.Error(t, err)
	assert.Equal(t, "用户名已存在", err.Error())

	var count int64
	database.DB.Model(&models.User{}).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestUserService_CreateUser_WithAdminRole(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	service := NewUserService()

	user, err := service.CreateUser("adminuser", "password123", models.RoleAdmin)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, models.RoleAdmin, user.Role)
}

func TestUserService_CreateUser_WithFinanceRole(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	service := NewUserService()

	user, err := service.CreateUser("financeuser", "password123", models.RoleFinance)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, models.RoleFinance, user.Role)
}

func TestUserService_GetUserByID_Success(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	service := NewUserService()

	createdUser, err := service.CreateUser("testuser", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	foundUser, err := service.GetUserByID(createdUser.ID)

	assert.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, createdUser.ID, foundUser.ID)
	assert.Equal(t, "testuser", foundUser.Username)
}

func TestUserService_GetUserByID_NotFound(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	service := NewUserService()

	foundUser, err := service.GetUserByID(999)

	assert.Error(t, err)
	assert.Nil(t, foundUser)
}

func TestUserService_Login_Success(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	service := NewUserService()

	_, err := service.CreateUser("testuser", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	token, user, err := service.Login("testuser", "password123")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Username)
}

func TestUserService_Login_WrongPassword(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	service := NewUserService()

	_, err := service.CreateUser("testuser", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	token, user, err := service.Login("testuser", "wrongpassword")

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Nil(t, user)
	assert.Equal(t, "用户名或密码错误", err.Error())
}

func TestUserService_Login_UserNotFound(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	service := NewUserService()

	token, user, err := service.Login("nonexistent", "password123")

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Nil(t, user)
	assert.Equal(t, "用户名或密码错误", err.Error())
}
