package handler

import (
	"strconv"

	"distribution-commission/internal/pkg/response"
	"distribution-commission/internal/service"

	"github.com/gin-gonic/gin"
)

type AuditLogHandler struct {
	auditLogService *service.AuditLogService
}

func NewAuditLogHandler() *AuditLogHandler {
	return &AuditLogHandler{
		auditLogService: service.NewAuditLogService(),
	}
}

func (h *AuditLogHandler) GetAuditLogList(c *gin.Context) {
	resourceType := c.Query("resource_type")
	resourceIDStr := c.Query("resource_id")
	distributorIDStr := c.Query("distributor_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var resourceID uint
	if resourceIDStr != "" {
		if id, err := strconv.ParseUint(resourceIDStr, 10, 64); err == nil {
			resourceID = uint(id)
		}
	}

	var distributorID uint
	if distributorIDStr != "" {
		if id, err := strconv.ParseUint(distributorIDStr, 10, 64); err == nil {
			distributorID = uint(id)
		}
	}

	logs, total, err := h.auditLogService.List(resourceType, resourceID, distributorID, page, pageSize)
	if err != nil {
		response.InternalError(c, "查询审计日志失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
