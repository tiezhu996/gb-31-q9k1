package handler

import (
	"errors"
	"log/slog"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/dto"
	"github.com/petsocial/petsocial/internal/middleware"
	"github.com/petsocial/petsocial/internal/model"
	"github.com/petsocial/petsocial/internal/repository"
	"github.com/petsocial/petsocial/internal/service"
	"github.com/petsocial/petsocial/internal/util"
)

// MeetupHandler 约伴帖接口层。
type MeetupHandler struct {
	svc    *service.MeetupService
	users  *repository.UserRepository
	logger *slog.Logger
}

// NewMeetupHandler 构造注入。
func NewMeetupHandler(svc *service.MeetupService, users *repository.UserRepository, logger *slog.Logger) *MeetupHandler {
	return &MeetupHandler{svc: svc, users: users, logger: logger}
}

// Create 发布约伴帖。
func (h *MeetupHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	var req dto.CreateMeetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidParam, err))
		return
	}
	meetup, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, dto.ToMeetupResponse(meetup, middleware.GetUsername(c), userID))
}

// Get 约伴详情。
func (h *MeetupHandler) Get(c *gin.Context) {
	id, err := dto.ParseObjectID(c.Param("id"))
	if err != nil {
		c.Error(util.Validation("约伴 ID 不合法", err))
		return
	}
	userID := h.currentUserID(c)
	meetup, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	name := h.resolveUserName(c, meetup.CreatorID)
	util.OK(c, dto.ToMeetupResponse(meetup, name, userID))
}

// List 约伴列表（复用 MeetupService.List）。
func (h *MeetupHandler) List(c *gin.Context) {
	page, pageSize := pageParams(c)
	city := c.Query("city")
	status := c.Query("status")
	meetups, total, err := h.svc.List(c.Request.Context(), city, status, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	userID := h.currentUserID(c)
	items := make([]interface{}, 0, len(meetups))
	for i := range meetups {
		name := h.resolveUserName(c, meetups[i].CreatorID)
		items = append(items, dto.ToMeetupResponse(&meetups[i], name, userID))
	}
	util.Page(c, items, total, page, pageSize)
}

// ListMy 我的约伴。
func (h *MeetupHandler) ListMy(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	page, pageSize := pageParams(c)
	meetups, total, err := h.svc.ListMy(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	items := make([]interface{}, 0, len(meetups))
	for i := range meetups {
		name := h.resolveUserName(c, meetups[i].CreatorID)
		items = append(items, dto.ToMeetupResponse(&meetups[i], name, userID))
	}
	util.Page(c, items, total, page, pageSize)
}

// Join 报名约伴。
func (h *MeetupHandler) Join(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	id, err := dto.ParseObjectID(c.Param("id"))
	if err != nil {
		c.Error(util.Validation("约伴 ID 不合法", err))
		return
	}
	if err := h.svc.Join(c.Request.Context(), userID, id, c.ClientIP()); err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"message": constants.MsgJoinSuccess})
}

// CancelJoin 取消报名。
func (h *MeetupHandler) CancelJoin(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	id, err := dto.ParseObjectID(c.Param("id"))
	if err != nil {
		c.Error(util.Validation("约伴 ID 不合法", err))
		return
	}
	if err := h.svc.CancelJoin(c.Request.Context(), userID, id); err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"message": constants.MsgCancelJoin})
}

// UpdateStatus 变更约伴状态。
func (h *MeetupHandler) UpdateStatus(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	id, err := dto.ParseObjectID(c.Param("id"))
	if err != nil {
		c.Error(util.Validation("约伴 ID 不合法", err))
		return
	}
	var req dto.UpdateMeetupStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidParam, err))
		return
	}
	if err := h.svc.UpdateStatus(c.Request.Context(), userID, id, req); err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"message": "状态已更新"})
}

func (h *MeetupHandler) currentUserID(c *gin.Context) primitive.ObjectID {
	if id, ok := middleware.GetUserID(c); ok {
		return id
	}
	return primitive.NilObjectID
}

func (h *MeetupHandler) resolveUserName(c *gin.Context, userID primitive.ObjectID) string {
	user, err := h.users.FindByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return "铲屎官"
		}
		return ""
	}
	return user.Nickname
}

func currentUserIDHex(c *gin.Context) string {
	if id, ok := middleware.GetUserID(c); ok {
		return id.Hex()
	}
	return ""
}

var _ = model.Meetup{}
