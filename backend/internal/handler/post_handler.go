package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/dto"
	"github.com/petsocial/petsocial/internal/middleware"
	"github.com/petsocial/petsocial/internal/model"
	"github.com/petsocial/petsocial/internal/service"
	"github.com/petsocial/petsocial/internal/util"
	"go.mongodb.org/mongo-driver/bson"
)

// PostHandler 动态接口层。
type PostHandler struct {
	svc    *service.PostService
	feed   *service.FeedService
	logger *slog.Logger
}

// NewPostHandler 构造注入。
func NewPostHandler(svc *service.PostService, feed *service.FeedService, logger *slog.Logger) *PostHandler {
	return &PostHandler{svc: svc, feed: feed, logger: logger}
}

// Create 发布动态。
func (h *PostHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	var req dto.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidParam, err))
		return
	}
	post, err := h.svc.CreatePost(c.Request.Context(), userID, req)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, dto.ToPostResponse(post))
}

// Get 动态详情（复用 FeedService.EnrichPosts 组装）。
func (h *PostHandler) Get(c *gin.Context) {
	id, err := dto.ParseObjectID(c.Param("id"))
	if err != nil {
		c.Error(util.Validation("动态 ID 不合法", err))
		return
	}
	post, err := h.svc.GetPost(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	userID := primitive.NilObjectID
	if uid, ok := middleware.GetUserID(c); ok {
		userID = uid
	}
	resps, err := h.feed.EnrichPosts(c.Request.Context(), []model.Post{*post}, userID, map[primitive.ObjectID]*model.User{})
	if err != nil {
		c.Error(err)
		return
	}
	if len(resps) > 0 {
		util.OK(c, resps[0])
		return
	}
	util.OK(c, dto.ToPostResponse(post))
}

// List 动态列表（公开发现流，复用 PostService.ListPosts）。
func (h *PostHandler) List(c *gin.Context) {
	page, pageSize := pageParams(c)
	posts, total, err := h.svc.ListPosts(c.Request.Context(), bson.M{"status": constants.PostStatusApproved}, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	util.Page(c, h.enrichList(c, posts), total, page, pageSize)
}

// ListByAuthor 用户主页动态。
func (h *PostHandler) ListByAuthor(c *gin.Context) {
	id, err := dto.ParseObjectID(c.Param("id"))
	if err != nil {
		c.Error(util.Validation("用户 ID 不合法", err))
		return
	}
	page, pageSize := pageParams(c)
	posts, total, err := h.svc.ListByAuthor(c.Request.Context(), id, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	util.Page(c, h.enrichList(c, posts), total, page, pageSize)
}

// ListByTopic 话题动态。
func (h *PostHandler) ListByTopic(c *gin.Context) {
	topic := c.Param("name")
	page, pageSize := pageParams(c)
	posts, total, err := h.svc.ListByTopic(c.Request.Context(), topic, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	util.Page(c, h.enrichList(c, posts), total, page, pageSize)
}

// UpdateStatus 审核状态流转（管理员）。
func (h *PostHandler) UpdateStatus(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	id, err := dto.ParseObjectID(c.Param("id"))
	if err != nil {
		c.Error(util.Validation("动态 ID 不合法", err))
		return
	}
	var req dto.UpdatePostStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidParam, err))
		return
	}
	if err := h.svc.UpdateStatus(c.Request.Context(), userID, id, req); err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"message": "审核完成"})
}

// AddComment 评论。
func (h *PostHandler) AddComment(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	id, err := dto.ParseObjectID(c.Param("id"))
	if err != nil {
		c.Error(util.Validation("动态 ID 不合法", err))
		return
	}
	var req dto.CommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidParam, err))
		return
	}
	comment, err := h.svc.AddComment(c.Request.Context(), userID, id, req.Content)
	if err != nil {
		c.Error(err)
		return
	}
	username := middleware.GetUsername(c)
	util.OK(c, dto.ToCommentResponse(comment, username))
}

// ListComments 评论列表。
func (h *PostHandler) ListComments(c *gin.Context) {
	id, err := dto.ParseObjectID(c.Param("id"))
	if err != nil {
		c.Error(util.Validation("动态 ID 不合法", err))
		return
	}
	page, pageSize := pageParams(c)
	comments, total, err := h.svc.ListComments(c.Request.Context(), id, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	util.Page(c, comments, total, page, pageSize)
}

// DeleteComment 删除评论。
func (h *PostHandler) DeleteComment(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	commentID, err := dto.ParseObjectID(c.Param("comment_id"))
	if err != nil {
		c.Error(util.Validation("评论 ID 不合法", err))
		return
	}
	postID, err := dto.ParseObjectID(c.Param("id"))
	if err != nil {
		c.Error(util.Validation("动态 ID 不合法", err))
		return
	}
	isAdmin := middleware.GetRole(c) == string(constants.RoleAdmin)
	if err := h.svc.DeleteComment(c.Request.Context(), userID, commentID, postID, isAdmin); err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"message": "评论已删除"})
}

// Interact 点赞/收藏/转发。
func (h *PostHandler) Interact(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	id, err := dto.ParseObjectID(c.Param("id"))
	if err != nil {
		c.Error(util.Validation("动态 ID 不合法", err))
		return
	}
	var req dto.InteractionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidParam, err))
		return
	}
	result, err := h.svc.Interact(c.Request.Context(), userID, id, req.Type)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, result)
}

// ListTopics 话题广场。
func (h *PostHandler) ListTopics(c *gin.Context) {
	page, pageSize := pageParams(c)
	topics, total, err := h.svc.ListTopics(c.Request.Context(), page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	util.Page(c, topics, total, page, pageSize)
}

// CreateTopic 创建话题（管理员）。
func (h *PostHandler) CreateTopic(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required,min=1,max=32"`
		Description string `json:"description" binding:"max=255"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidParam, err))
		return
	}
	topic, err := h.svc.CreateTopic(c.Request.Context(), req.Name, req.Description)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, topic)
}

// enrichList 组装带作者信息的动态列表。
func (h *PostHandler) enrichList(c *gin.Context, posts []model.Post) []interface{} {
	userID := primitive.NilObjectID
	if uid, ok := middleware.GetUserID(c); ok {
		userID = uid
	}
	resps, err := h.feed.EnrichPosts(c.Request.Context(), posts, userID, map[primitive.ObjectID]*model.User{})
	if err != nil {
		return []interface{}{}
	}
	out := make([]interface{}, 0, len(resps))
	for i := range resps {
		out = append(out, resps[i])
	}
	return out
}
