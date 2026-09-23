package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/petsocial/petsocial/internal/dto"
	"github.com/petsocial/petsocial/internal/service"
	"github.com/petsocial/petsocial/internal/util"
)

// AuditHandler 审计日志接口层（管理员）。
type AuditHandler struct {
	svc    *service.AuditService
	logger *slog.Logger
}

// NewAuditHandler 构造注入。
func NewAuditHandler(svc *service.AuditService, logger *slog.Logger) *AuditHandler {
	return &AuditHandler{svc: svc, logger: logger}
}

// List 审计日志分页查询。
func (h *AuditHandler) List(c *gin.Context) {
	var req dto.ListAuditRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(util.Validation("查询参数不合法", err))
		return
	}
	page, pageSize := pageParams(c)
	logs, total, err := h.svc.List(c.Request.Context(), req.Action, req.Username, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	items := make([]interface{}, 0, len(logs))
	for i := range logs {
		items = append(items, dto.ToAuditLogResponse(&logs[i]))
	}
	util.Page(c, items, total, page, pageSize)
}
