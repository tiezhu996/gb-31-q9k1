package handler

import (
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/dto"
	"github.com/petsocial/petsocial/internal/middleware"
	"github.com/petsocial/petsocial/internal/model"
	"github.com/petsocial/petsocial/internal/service"
	"github.com/petsocial/petsocial/internal/util"
)

// UserHandler 用户接口层。
type UserHandler struct {
	svc    *service.UserService
	logger *slog.Logger
}

// NewUserHandler 构造注入。
func NewUserHandler(svc *service.UserService, logger *slog.Logger) *UserHandler {
	return &UserHandler{svc: svc, logger: logger}
}

// Register 注册。
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidParam, err))
		return
	}
	data, err := h.svc.Register(c.Request.Context(), req, c.ClientIP())
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, data)
}

// Login 登录。
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidParam, err))
		return
	}
	data, err := h.svc.Login(c.Request.Context(), req, c.ClientIP())
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, data)
}

// GetProfile 我的资料。
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	user, err := h.svc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, dto.ToUserResponse(user))
}

// GetUserByID 用户公开主页。
func (h *UserHandler) GetUserByID(c *gin.Context) {
	id, err := dto.ParseObjectID(c.Param("id"))
	if err != nil {
		c.Error(util.Validation("用户 ID 不合法", err))
		return
	}
	user, err := h.svc.GetProfile(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, dto.ToUserResponse(user))
}

// UpdateProfile 更新资料。
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidParam, err))
		return
	}
	user, err := h.svc.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, dto.ToUserResponse(user))
}

// Follow 关注。
func (h *UserHandler) Follow(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	var req dto.FollowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidParam, err))
		return
	}
	followeeID, err := dto.ParseObjectID(req.FolloweeID)
	if err != nil {
		c.Error(util.Validation("用户 ID 不合法", err))
		return
	}
	if err := h.svc.Follow(c.Request.Context(), userID, followeeID, c.ClientIP()); err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"message": constants.MsgFollowSuccess})
}

// Unfollow 取消关注。
func (h *UserHandler) Unfollow(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	followeeID, err := dto.ParseObjectID(c.Param("id"))
	if err != nil {
		c.Error(util.Validation("用户 ID 不合法", err))
		return
	}
	if err := h.svc.Unfollow(c.Request.Context(), userID, followeeID, c.ClientIP()); err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"message": constants.MsgUnfollowSuccess})
}

// ListFollowing 关注列表。
func (h *UserHandler) ListFollowing(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	page, pageSize := pageParams(c)
	users, total, err := h.svc.ListFollowing(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	util.Page(c, mapUsers(users), total, page, pageSize)
}

// ListFollowers 粉丝列表。
func (h *UserHandler) ListFollowers(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	page, pageSize := pageParams(c)
	users, total, err := h.svc.ListFollowers(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	util.Page(c, mapUsers(users), total, page, pageSize)
}

func pageParams(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

func mapUsers(users []model.User) []interface{} {
	out := make([]interface{}, 0, len(users))
	for i := range users {
		out = append(out, dto.ToUserResponse(&users[i]))
	}
	return out
}
