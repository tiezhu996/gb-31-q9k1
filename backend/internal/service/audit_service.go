package service

import (
	"context"
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/model"
	"github.com/petsocial/petsocial/internal/repository"
	"time"
)

// AuditService 审计日志服务：middleware 与业务 service 共用埋点。
type AuditService struct {
	repo   *repository.AuditRepository
	logger *slog.Logger
}

// NewAuditService 构造注入。
func NewAuditService(repo *repository.AuditRepository, logger *slog.Logger) *AuditService {
	return &AuditService{repo: repo, logger: logger}
}

// Record 记录审计日志。
func (s *AuditService) Record(ctx context.Context, userID primitive.ObjectID, username, action, resource, resourceID, detail, ip string) {
	log := &model.AuditLog{
		UserID:     userID,
		Username:   username,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Detail:     detail,
		IP:         ip,
		CreatedAt:  time.Now(),
	}
	if err := s.repo.Create(ctx, log); err != nil {
		s.logger.Error("audit record failed", "action", action, "resource", resource, "error", err)
		return
	}
	s.logger.Info(fmt.Sprintf(constants.LogAuditCreated, userID.Hex(), action, resource), "ip", ip)
}

// List 分页查询审计日志。
func (s *AuditService) List(ctx context.Context, action, username string, page, pageSize int) ([]model.AuditLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	logs, total, err := s.repo.List(ctx, action, username, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("audit service list: %w", err)
	}
	return logs, total, nil
}
