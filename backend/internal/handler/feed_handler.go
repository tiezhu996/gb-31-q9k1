package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/middleware"
	"github.com/petsocial/petsocial/internal/model"
	"github.com/petsocial/petsocial/internal/service"
	"github.com/petsocial/petsocial/internal/util"
)

// FeedHandler Feed 接口层：关注页/发现页/同城页。
type FeedHandler struct {
	svc    *service.FeedService
	logger *slog.Logger
}

// NewFeedHandler 构造注入。
func NewFeedHandler(svc *service.FeedService, logger *slog.Logger) *FeedHandler {
	return &FeedHandler{svc: svc, logger: logger}
}

// Following 关注页 Feed。
func (h *FeedHandler) Following(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	page, pageSize := pageParams(c)
	posts, total, err := h.svc.FollowingFeed(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	h.respond(c, posts, total, page, pageSize)
}

// Discover 发现页 Feed。
func (h *FeedHandler) Discover(c *gin.Context) {
	page, pageSize := pageParams(c)
	posts, total, err := h.svc.DiscoverFeed(c.Request.Context(), page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	h.respond(c, posts, total, page, pageSize)
}

// Nearby 同城页 Feed。
func (h *FeedHandler) Nearby(c *gin.Context) {
	page, pageSize := pageParams(c)
	city := c.Query("city")
	posts, total, err := h.svc.NearbyFeed(c.Request.Context(), city, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	h.respond(c, posts, total, page, pageSize)
}

func (h *FeedHandler) respond(c *gin.Context, posts []model.Post, total int64, page, pageSize int) {
	userID := primitive.NilObjectID
	if uid, ok := middleware.GetUserID(c); ok {
		userID = uid
	}
	resps, err := h.svc.EnrichPosts(c.Request.Context(), posts, userID, map[primitive.ObjectID]*model.User{})
	if err != nil {
		c.Error(err)
		return
	}
	util.Page(c, resps, total, page, pageSize)
}

var _ = constants.MsgOK
