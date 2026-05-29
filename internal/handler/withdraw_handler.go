package handler

import (
	"strconv"

	"distribution-commission/internal/middleware"
	"distribution-commission/internal/models"
	"distribution-commission/internal/pkg/response"
	"distribution-commission/internal/service"

	"github.com/gin-gonic/gin"
)

type WithdrawHandler struct {
	withdrawService    *service.WithdrawService
	distributorService *service.DistributorService
}

func NewWithdrawHandler() *WithdrawHandler {
	return &WithdrawHandler{
		withdrawService:    service.NewWithdrawService(),
		distributorService: service.NewDistributorService(),
	}
}

type SubmitWithdrawRequest struct {
	Amount int64 `json:"amount" binding:"required,gt=0"`
}

type ApproveWithdrawRequest struct {
	WithdrawID uint `json:"withdraw_id" binding:"required"`
}

type RejectWithdrawRequest struct {
	WithdrawID uint   `json:"withdraw_id" binding:"required"`
	Reason     string `json:"reason" binding:"required"`
}

func (h *WithdrawHandler) SubmitWithdraw(c *gin.Context) {
	userID := middleware.GetUserID(c)

	distributor, err := h.distributorService.GetByUserID(userID)
	if err != nil {
		response.NotFound(c, "分销员不存在")
		return
	}

	var req SubmitWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	withdraw, err := h.withdrawService.SubmitWithdraw(distributor.ID, req.Amount)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, withdraw)
}

func (h *WithdrawHandler) ApproveWithdraw(c *gin.Context) {
	var req ApproveWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	operatorID := middleware.GetUserID(c)

	err := h.withdrawService.Approve(req.WithdrawID, operatorID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "审核通过", nil)
}

func (h *WithdrawHandler) RejectWithdraw(c *gin.Context) {
	var req RejectWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	operatorID := middleware.GetUserID(c)

	err := h.withdrawService.Reject(req.WithdrawID, operatorID, req.Reason)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "已驳回", nil)
}

func (h *WithdrawHandler) GetWithdrawList(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetRole(c)

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var withdraws []models.WithdrawRequest
	var total int64
	var err error

	if role == "finance" || role == "admin" {
		withdraws, total, err = h.withdrawService.ListAll(status, page, pageSize)
	} else {
		distributor, err := h.distributorService.GetByUserID(userID)
		if err != nil {
			response.NotFound(c, "分销员不存在")
			return
		}
		withdraws, total, err = h.withdrawService.ListByDistributor(distributor.ID, status, page, pageSize)
	}

	if err != nil {
		response.InternalError(c, "查询提现列表失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":      withdraws,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
