package service

import (
	"distribution-commission/internal/database"
	"distribution-commission/internal/models"
)

type AuditLogService struct{}

func NewAuditLogService() *AuditLogService {
	return &AuditLogService{}
}

func (s *AuditLogService) Create(log *models.AuditLog) error {
	return database.DB.Create(log).Error
}

func (s *AuditLogService) List(resourceType string, resourceID uint, distributorID uint, page, pageSize int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	offset := (page - 1) * pageSize

	query := database.DB.Model(&models.AuditLog{})

	if resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}

	if resourceID > 0 {
		query = query.Where("resource_id = ?", resourceID)
	}

	if distributorID > 0 {
		query = query.Where("distributor_id = ?", distributorID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Operator").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
