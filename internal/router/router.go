package router

import (
	"distribution-commission/internal/handler"
	"distribution-commission/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	authHandler := handler.NewAuthHandler()
	orderHandler := handler.NewOrderHandler()
	commissionHandler := handler.NewCommissionHandler()
	withdrawHandler := handler.NewWithdrawHandler()
	auditLogHandler := handler.NewAuditLogHandler()

	api := r.Group("/api/v1")
	{
		api.POST("/login", authHandler.Login)

		api.POST("/distributors", middleware.JWTAuth(), middleware.RequireRoles("admin", "finance"), authHandler.CreateDistributor)

		api.GET("/profile", middleware.JWTAuth(), authHandler.GetProfile)

		api.POST("/orders", middleware.JWTAuth(), middleware.RequireRoles("distributor"), orderHandler.CreateCompletedOrder)
		api.GET("/orders", middleware.JWTAuth(), middleware.RequireRoles("distributor"), orderHandler.GetOrderList)
		api.POST("/orders/refund", middleware.JWTAuth(), middleware.RequireRoles("admin", "finance", "distributor"), orderHandler.RefundOrder)

		api.GET("/commissions", middleware.JWTAuth(), middleware.RequireRoles("distributor"), commissionHandler.GetCommissionList)
		api.POST("/commissions/generate", middleware.JWTAuth(), commissionHandler.GenerateCommission)
		api.POST("/commissions/settle", middleware.JWTAuth(), middleware.RequireRoles("finance", "admin"), commissionHandler.SettleCommission)

		api.GET("/balance", middleware.JWTAuth(), middleware.RequireRoles("distributor"), commissionHandler.GetBalance)

		api.POST("/withdraws", middleware.JWTAuth(), middleware.RequireRoles("distributor"), withdrawHandler.SubmitWithdraw)
		api.GET("/withdraws", middleware.JWTAuth(), withdrawHandler.GetWithdrawList)
		api.POST("/withdraws/approve", middleware.JWTAuth(), middleware.RequireRoles("finance", "admin"), withdrawHandler.ApproveWithdraw)
		api.POST("/withdraws/reject", middleware.JWTAuth(), middleware.RequireRoles("finance", "admin"), withdrawHandler.RejectWithdraw)

		api.GET("/audit-logs", middleware.JWTAuth(), middleware.RequireRoles("finance", "admin"), auditLogHandler.GetAuditLogList)
	}

	return r
}
