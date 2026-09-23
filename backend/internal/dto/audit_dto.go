package dto

import "github.com/petsocial/petsocial/internal/model"

// AuditLogResponse 审计日志出参。
type AuditLogResponse struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	Username   string `json:"username"`
	Action     string `json:"action"`
	Resource   string `json:"resource"`
	ResourceID string `json:"resource_id"`
	Detail     string `json:"detail"`
	IP         string `json:"ip"`
	CreatedAt  string `json:"created_at"`
}

// ToAuditLogResponse 模型转出参。
func ToAuditLogResponse(a *model.AuditLog) AuditLogResponse {
	return AuditLogResponse{
		ID:         a.ID.Hex(),
		UserID:     a.UserID.Hex(),
		Username:   a.Username,
		Action:     a.Action,
		Resource:   a.Resource,
		ResourceID: a.ResourceID,
		Detail:     a.Detail,
		IP:         a.IP,
		CreatedAt:  a.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ListAuditRequest 审计查询入参。
type ListAuditRequest struct {
	Action   string `form:"action"`
	Username string `form:"username"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}
