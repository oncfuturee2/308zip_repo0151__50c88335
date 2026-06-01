package handler

import (
	"strconv"

	"distribution-commission/internal/middleware"
	"distribution-commission/internal/pkg/response"
	"distribution-commission/internal/service"

	"github.com/gin-gonic/gin"
)

type CommissionHandler struct {
	commissionService  *service.CommissionService
	distributorService *service.DistributorService
}

func NewCommissionHandler() *CommissionHandler {
	return &CommissionHandler{
		commissionService:  service.NewCommissionService(),
		distributorService: service.NewDistributorService(),
	}
}

type GenerateCommissionRequest struct {
	OrderNo string `json:"order_no" binding:"required"`
}

type SettleCommissionRequest struct {
	CommissionID uint `json:"commission_id" binding:"required"`
}

func (h *CommissionHandler) GenerateCommission(c *gin.Context) {
	var req GenerateCommissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	commissions, err := h.commissionService.GenerateCommission(c.Request.Context(), req.OrderNo)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, commissions)
}

func (h *CommissionHandler) SettleCommission(c *gin.Context) {
	var req SettleCommissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	operatorID := middleware.GetUserID(c)

	err := h.commissionService.SettleCommission(req.CommissionID, operatorID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "结算成功", nil)
}

func (h *CommissionHandler) GetCommissionList(c *gin.Context) {
	userID := middleware.GetUserID(c)

	distributor, err := h.distributorService.GetByUserID(userID)
	if err != nil {
		response.NotFound(c, "分销员不存在")
		return
	}

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	commissions, total, err := h.commissionService.ListByDistributor(distributor.ID, status, page, pageSize)
	if err != nil {
		response.InternalError(c, "查询佣金列表失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":      commissions,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *CommissionHandler) GetBalance(c *gin.Context) {
	userID := middleware.GetUserID(c)

	distributor, err := h.distributorService.GetByUserID(userID)
	if err != nil {
		response.NotFound(c, "分销员不存在")
		return
	}

	balance, frozenBalance, totalCommission, err := h.distributorService.GetBalance(distributor.ID)
	if err != nil {
		response.InternalError(c, "查询余额失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"balance":           balance,
		"frozen_balance":    frozenBalance,
		"total_commission":  totalCommission,
	})
}
