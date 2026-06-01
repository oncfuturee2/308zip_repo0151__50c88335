package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"distribution-commission/internal/config"
	"distribution-commission/internal/database"
	"distribution-commission/internal/models"
	"distribution-commission/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
	os.Setenv("JWT_SECRET", "test_secret_key")
	os.Setenv("JWT_EXPIRE_HOURS", "24")
	config.LoadConfig()
}

func setupTestRouter() *gin.Engine {
	database.SetupTestDB()

	_ = service.NewUserService()

	r := gin.Default()
	return r
}

func TestAuthHandler_CreateDistributor_MissingUsername(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	handler := NewAuthHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := CreateDistributorRequest{
		Password: "password123",
		RealName: "张三",
	}
	jsonBody, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest("POST", "/api/v1/distributors", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateDistributor(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(400), response["code"])
	assert.Contains(t, response["message"], "参数错误")
}

func TestAuthHandler_CreateDistributor_MissingPassword(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	handler := NewAuthHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := CreateDistributorRequest{
		Username: "newdistributor",
		RealName: "张三",
	}
	jsonBody, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest("POST", "/api/v1/distributors", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateDistributor(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(400), response["code"])
	assert.Contains(t, response["message"], "参数错误")
}

func TestAuthHandler_CreateDistributor_DuplicateUsername(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	userService := service.NewUserService()
	_, err := userService.CreateUser("existinguser", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	handler := NewAuthHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := CreateDistributorRequest{
		Username: "existinguser",
		Password: "anotherpassword",
		RealName: "李四",
	}
	jsonBody, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest("POST", "/api/v1/distributors", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateDistributor(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(400), response["code"])
	assert.Equal(t, "用户名已存在", response["message"])

	var userCount int64
	database.DB.Model(&models.User{}).Count(&userCount)
	assert.Equal(t, int64(1), userCount)
}

func TestAuthHandler_CreateDistributor_Success(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	handler := NewAuthHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := CreateDistributorRequest{
		Username:   "newdistributor",
		Password:   "password123",
		RealName:   "张三",
		Phone:      "13800138000",
		BankCardNo: "6222021234567890",
		BankName:   "工商银行",
	}
	jsonBody, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest("POST", "/api/v1/distributors", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateDistributor(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(200), response["code"])
	assert.Equal(t, "success", response["message"])

	data, ok := response["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "张三", data["real_name"])
	assert.Equal(t, "13800138000", data["phone"])
	assert.Equal(t, "6222021234567890", data["bank_card_no"])
	assert.Equal(t, "工商银行", data["bank_name"])
	assert.Equal(t, float64(1), data["status"])

	var userCount int64
	database.DB.Model(&models.User{}).Count(&userCount)
	assert.Equal(t, int64(1), userCount)

	var distributorCount int64
	database.DB.Model(&models.Distributor{}).Count(&distributorCount)
	assert.Equal(t, int64(1), distributorCount)

	var user models.User
	database.DB.First(&user)
	assert.Equal(t, "newdistributor", user.Username)
	assert.Equal(t, models.RoleDistributor, user.Role)

	var distributor models.Distributor
	database.DB.Preload("User").First(&distributor)
	assert.Equal(t, user.ID, distributor.UserID)
	assert.Equal(t, "张三", distributor.RealName)
	assert.Equal(t, user.Username, distributor.User.Username)
}

func TestAuthHandler_CreateDistributor_Success_WithParentDistributor(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	userService := service.NewUserService()
	distributorService := service.NewDistributorService()

	parentUser, err := userService.CreateUser("parentuser", "password123", models.RoleDistributor)
	assert.NoError(t, err)
	parentDistributor, err := distributorService.CreateDistributor(parentUser.ID, 0, "父级", "", "", "")
	assert.NoError(t, err)

	handler := NewAuthHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := CreateDistributorRequest{
		Username:            "childuser",
		Password:            "password123",
		ParentDistributorID: parentDistributor.ID,
		RealName:            "子级",
	}
	jsonBody, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest("POST", "/api/v1/distributors", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateDistributor(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(200), response["code"])
	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(parentDistributor.ID), data["parent_id"])
}

func TestAuthHandler_CreateDistributor_Success_WithMinimalFields(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	handler := NewAuthHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := CreateDistributorRequest{
		Username: "minimaluser",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest("POST", "/api/v1/distributors", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateDistributor(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(200), response["code"])
	assert.Equal(t, "success", response["message"])

	var userCount int64
	database.DB.Model(&models.User{}).Count(&userCount)
	assert.Equal(t, int64(1), userCount)

	var distributorCount int64
	database.DB.Model(&models.Distributor{}).Count(&distributorCount)
	assert.Equal(t, int64(1), distributorCount)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	userService := service.NewUserService()
	_, err := userService.CreateUser("testuser", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	handler := NewAuthHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Login(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(200), response["code"])
	assert.Equal(t, "success", response["message"])

	data, ok := response["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, data["token"])

	user, ok := data["user"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "testuser", user["username"])
}

func TestAuthHandler_Login_WrongPassword(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	userService := service.NewUserService()
	_, err := userService.CreateUser("testuser", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	handler := NewAuthHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}
	jsonBody, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Login(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(401), response["code"])
	assert.Equal(t, "用户名或密码错误", response["message"])
}

func TestAuthHandler_Login_MissingCredentials(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	handler := NewAuthHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]string{
		"username": "testuser",
	}
	jsonBody, _ := json.Marshal(reqBody)
	c.Request, _ = http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Login(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(400), response["code"])
	assert.Contains(t, response["message"], "参数错误")
}

func TestAuthHandler_GetProfile_Success(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	userService := service.NewUserService()
	user, err := userService.CreateUser("testuser", "password123", models.RoleDistributor)
	assert.NoError(t, err)

	handler := NewAuthHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set("user_id", user.ID)

	c.Request, _ = http.NewRequest("GET", "/api/v1/profile", nil)

	handler.GetProfile(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(200), response["code"])
	assert.Equal(t, "success", response["message"])

	data, ok := response["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "testuser", data["username"])
	assert.Equal(t, float64(user.ID), data["id"])
}

func TestAuthHandler_GetProfile_UserNotFound(t *testing.T) {
	database.SetupTestDB()
	defer database.CleanupTestDB()

	handler := NewAuthHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set("user_id", uint(999))

	c.Request, _ = http.NewRequest("GET", "/api/v1/profile", nil)

	handler.GetProfile(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(404), response["code"])
	assert.Equal(t, "用户不存在", response["message"])
}
