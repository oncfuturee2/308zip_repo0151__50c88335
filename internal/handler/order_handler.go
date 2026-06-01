package handler

import (
	"strconv"

	"distribution-commission/internal/middleware"
	"distribution-commission/internal/pkg/response"
	"distribution-commission/internal/service"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService     *service.OrderService
	distributorService *service.DistributorService
}

func NewOrderHandler() *OrderHandler {
	return &OrderHandler{
		orderService:       service.NewOrderService(),
		distributorService: service.NewDistributorService(),
	}
}

type CreateOrderRequest struct {
	OrderNo   string `json:"order_no"`
	Amount    int64  `json:"amount" binding:"required,gt=0"`
	GoodsName string `json:"goods_name"`
}

type RefundOrderRequest struct {
	OrderNo string `json:"order_no" binding:"required"`
}

func (h *OrderHandler) CreateCompletedOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)

	distributor, err := h.distributorService.GetByUserID(userID)
	if err != nil {
		response.NotFound(c, "分销员不存在")
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	order, err := h.orderService.CreateCompletedOrder(c.Request.Context(), req.OrderNo, distributor.ID, req.Amount, req.GoodsName)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, order)
}

func (h *OrderHandler) GetOrderList(c *gin.Context) {
	userID := middleware.GetUserID(c)

	distributor, err := h.distributorService.GetByUserID(userID)
	if err != nil {
		response.NotFound(c, "分销员不存在")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	orders, total, err := h.orderService.ListByDistributor(distributor.ID, page, pageSize)
	if err != nil {
		response.InternalError(c, "查询订单列表失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":      orders,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *OrderHandler) RefundOrder(c *gin.Context) {
	var req RefundOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	operatorID := middleware.GetUserID(c)

	if err := h.orderService.RefundOrder(c.Request.Context(), req.OrderNo, operatorID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "订单退款成功", nil)
}
