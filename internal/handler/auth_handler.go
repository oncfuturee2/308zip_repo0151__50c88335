package handler

import (
	"distribution-commission/internal/middleware"
	"distribution-commission/internal/models"
	"distribution-commission/internal/pkg/response"
	"distribution-commission/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userService *service.UserService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		userService: service.NewUserService(),
	}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	token, user, err := h.userService.Login(req.Username, req.Password)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, LoginResponse{
		Token: token,
		User:  user,
	})
}

type CreateDistributorRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	RealName   string `json:"real_name"`
	Phone      string `json:"phone"`
	BankCardNo string `json:"bank_card_no"`
	BankName   string `json:"bank_name"`
	ParentID   *uint  `json:"parent_id"`
}

func (h *AuthHandler) CreateDistributor(c *gin.Context) {
	var req CreateDistributorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	user, err := h.userService.CreateUser(req.Username, req.Password, models.RoleDistributor)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	distributorService := service.NewDistributorService()
	distributor, err := distributorService.CreateDistributor(user.ID, req.RealName, req.Phone, req.BankCardNo, req.BankName, req.ParentID)
	if err != nil {
		response.InternalError(c, "创建分销员失败: "+err.Error())
		return
	}

	response.Success(c, distributor)
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	response.Success(c, user)
}
