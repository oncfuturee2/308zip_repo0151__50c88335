package router

import (
	"distribution-commission/internal/database"
	"distribution-commission/internal/handler"
	"distribution-commission/internal/middleware"
	"distribution-commission/internal/repository"
	"distribution-commission/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	orderRepo := repository.NewGormOrderRepository(database.DB)
	orderService := service.NewOrderService(orderRepo)
	commissionService := service.NewCommissionService(orderService)

	authHandler := handler.NewAuthHandler()
	orderHandler := handler.NewOrderHandler(orderRepo)
	commissionHandler := handler.NewCommissionHandler(commissionService)
	withdrawHandler := handler.NewWithdrawHandler()
	auditLogHandler := handler.NewAuditLogHandler()

	api := r.Group("/api/v1")
	{
		api.POST("/login", authHandler.Login)

		api.POST("/distributors", middleware.JWTAuth(), middleware.RequireRoles("admin", "finance"), authHandler.CreateDistributor)

		api.GET("/profile", middleware.JWTAuth(), authHandler.GetProfile)

		api.POST("/orders", middleware.JWTAuth(), middleware.RequireRoles("distributor"), orderHandler.CreateCompletedOrder)
		api.GET("/orders", middleware.JWTAuth(), middleware.RequireRoles("distributor"), orderHandler.GetOrderList)

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
