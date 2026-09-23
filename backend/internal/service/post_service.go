package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/petsocial/petsocial/internal/config"
	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/dto"
	"github.com/petsocial/petsocial/internal/model"
	"github.com/petsocial/petsocial/internal/repository"
	"github.com/petsocial/petsocial/internal/util"
)

// PostService 动态业务：发布/审核/互动/评论/话题。
type PostService struct {
	repo   *repository.PostRepository
	users  *repository.UserRepository
	audit  *AuditService
	logger *slog.Logger
	cfg    *config.Config
}

// NewPostService 构造注入。
func NewPostService(repo *repository.PostRepository, users *repository.UserRepository, audit *AuditService, logger *slog.Logger, cfg *config.Config) *PostService {
	return &PostService{repo: repo, users: users, audit: audit, logger: logger, cfg: cfg}
}

// CreatePost 发布动态：敏感词过滤 + 状态流转。
func (s *PostService) CreatePost(ctx context.Context, authorID primitive.ObjectID, req dto.CreatePostRequest) (*model.Post, error) {
	if word := s.checkSensitive(req.Content); word != "" {
		s.logger.Warn(fmt.Sprintf(constants.LogSensitiveHit, "new", word))
		return nil, util.ContentRejected(constants.MsgContentRejected, constants.ErrContentRejected)
	}
	petIDs, err := dto.ParsePostIDs(req.PetIDs)
	if err != nil {
		return nil, util.Validation("宠物 ID 不合法", err)
	}
	media := make([]model.MediaItem, 0, len(req.Media))
	for _, m := range req.Media {
		media = append(media, model.MediaItem{Type: m.Type, URL: m.URL})
	}
	status := constants.PostStatusApproved
	post := &model.Post{
		AuthorID:  authorID,
		PetIDs:    petIDs,
		Content:   req.Content,
		Type:      req.Type,
		Media:     media,
		Topics:    req.Topics,
		Location:  req.Location,
		City:      req.City,
		Status:    status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.repo.Create(ctx, post); err != nil {
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	for _, topic := range req.Topics {
		_ = s.repo.IncTopicCount(ctx, topic, 1)
	}
	s.logger.Info(fmt.Sprintf(constants.LogPostCreated, post.ID.Hex(), authorID.Hex(), post.Type, post.Status))
	s.audit.Record(ctx, authorID, "", "post.create", "post", post.ID.Hex(), "发布动态", "")
	return post, nil
}

// checkSensitive 敏感词过滤。
func (s *PostService) checkSensitive(content string) string {
	words := strings.Split(s.cfg.SensitiveWords, ",")
	for _, w := range words {
		w = strings.TrimSpace(w)
		if w != "" && strings.Contains(content, w) {
			return w
		}
	}
	return ""
}

// GetPost 动态详情。
func (s *PostService) GetPost(ctx context.Context, postID primitive.ObjectID) (*model.Post, error) {
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, util.NotFound(constants.MsgNotFound, err)
		}
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	return post, nil
}

// ListPosts 动态列表（复用：发现页/话题页/同城页/用户主页均由 feed 或 post handler 调用）。
func (s *PostService) ListPosts(ctx context.Context, filter bson.M, page, pageSize int) ([]model.Post, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	posts, total, err := s.repo.List(ctx, filter, page, pageSize)
	if err != nil {
		return nil, 0, util.Internal(constants.MsgInternalError, err)
	}
	return posts, total, nil
}

// ListByAuthor 用户主页动态（复用 ListPosts 底层 repository.List）。
func (s *PostService) ListByAuthor(ctx context.Context, authorID primitive.ObjectID, page, pageSize int) ([]model.Post, int64, error) {
	return s.ListPosts(ctx, bson.M{"author_id": authorID}, page, pageSize)
}

// ListByTopic 话题动态（复用 ListPosts 底层 repository.List）。
func (s *PostService) ListByTopic(ctx context.Context, topic string, page, pageSize int) ([]model.Post, int64, error) {
	return s.ListPosts(ctx, bson.M{"topics": topic}, page, pageSize)
}

// ListByCity 同城动态（复用 ListPosts 底层 repository.List）。
func (s *PostService) ListByCity(ctx context.Context, city string, page, pageSize int) ([]model.Post, int64, error) {
	return s.ListPosts(ctx, bson.M{"city": city}, page, pageSize)
}

// UpdateStatus 审核状态流转（管理员）。
func (s *PostService) UpdateStatus(ctx context.Context, operatorID primitive.ObjectID, postID primitive.ObjectID, req dto.UpdatePostStatusRequest) error {
	if _, err := s.repo.FindByID(ctx, postID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return util.NotFound(constants.MsgNotFound, err)
		}
		return util.Internal(constants.MsgInternalError, err)
	}
	if err := s.repo.UpdateStatus(ctx, postID, req.Status, req.RejectReason); err != nil {
		return util.Internal(constants.MsgInternalError, err)
	}
	if req.Status == constants.PostStatusRejected {
		s.logger.Warn(fmt.Sprintf(constants.LogPostRejected, postID.Hex(), req.RejectReason))
		s.audit.Record(ctx, operatorID, "", "post.reject", "post", postID.Hex(), "驳回动态: "+req.RejectReason, "")
	} else {
		s.logger.Info(fmt.Sprintf(constants.LogPostApproved, postID.Hex()))
		s.audit.Record(ctx, operatorID, "", "post.approve", "post", postID.Hex(), "审核通过动态", "")
	}
	return nil
}

// AddComment 评论。
func (s *PostService) AddComment(ctx context.Context, userID primitive.ObjectID, postID primitive.ObjectID, content string) (*model.Comment, error) {
	if word := s.checkSensitive(content); word != "" {
		return nil, util.Validation("评论包含敏感词", constants.ErrContentRejected)
	}
	if _, err := s.repo.FindByID(ctx, postID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, util.NotFound(constants.MsgNotFound, err)
		}
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	comment := &model.Comment{PostID: postID, AuthorID: userID, Content: content, CreatedAt: time.Now()}
	if err := s.repo.CreateComment(ctx, comment); err != nil {
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	if err := s.repo.UpdateCounts(ctx, postID, 0, 1, 0, 0); err != nil {
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogCommentAdded, comment.ID.Hex(), postID.Hex(), userID.Hex()))
	return comment, nil
}

// ListComments 评论列表。
func (s *PostService) ListComments(ctx context.Context, postID primitive.ObjectID, page, pageSize int) ([]model.Comment, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListComments(ctx, postID, page, pageSize)
}

// DeleteComment 删除评论（作者本人或管理员）。
func (s *PostService) DeleteComment(ctx context.Context, userID primitive.ObjectID, commentID, postID primitive.ObjectID, isAdmin bool) error {
	comment, err := s.repo.FindComment(ctx, commentID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return util.NotFound(constants.MsgNotFound, err)
		}
		return util.Internal(constants.MsgInternalError, err)
	}
	if comment.AuthorID != userID && !isAdmin {
		return util.Forbidden("没有权限删除该评论", constants.ErrForbidden)
	}
	if err := s.repo.DeleteComment(ctx, commentID, postID, comment.AuthorID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return util.NotFound(constants.MsgNotFound, err)
		}
		return util.Internal(constants.MsgInternalError, err)
	}
	if err := s.repo.UpdateCounts(ctx, postID, 0, -1, 0, 0); err != nil {
		return util.Internal(constants.MsgInternalError, err)
	}
	return nil
}

// Interact 点赞/收藏/转发（幂等切换，唯一索引防重）。
func (s *PostService) Interact(ctx context.Context, userID, postID primitive.ObjectID, typ constants.InteractionType) (map[string]interface{}, error) {
	if _, err := s.repo.FindByID(ctx, postID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, util.NotFound(constants.MsgNotFound, err)
		}
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	exists, err := s.repo.HasInteraction(ctx, postID, userID, typ)
	if err != nil {
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	var incs map[string]int64
	if exists {
		if err := s.repo.DeleteInteraction(ctx, postID, userID, typ); err != nil {
			return nil, util.Internal(constants.MsgInternalError, err)
		}
		incs = interactionInc(typ, -1)
	} else {
		if err := s.repo.CreateInteraction(ctx, &model.Interaction{PostID: postID, UserID: userID, Type: typ, CreatedAt: time.Now()}); err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				return nil, util.Conflict("重复操作", constants.ErrInteractionDup)
			}
			return nil, util.Internal(constants.MsgInternalError, err)
		}
		incs = interactionInc(typ, 1)
	}
	if err := s.repo.UpdateCounts(ctx, postID, incs["like"], incs["comment"], incs["favorite"], incs["forward"]); err != nil {
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	action := map[constants.InteractionType]string{
		constants.InteractionLike:     "post.like",
		constants.InteractionFavorite: "post.favorite",
		constants.InteractionForward:  "post.forward",
	}[typ]
	s.logger.Info(fmt.Sprintf(logForInteraction(typ), postID.Hex(), userID.Hex()))
	s.audit.Record(ctx, userID, "", action, "post", postID.Hex(), "互动: "+string(typ), "")
	return map[string]interface{}{"active": !exists, "type": typ}, nil
}

// ListTopics 话题广场。
func (s *PostService) ListTopics(ctx context.Context, page, pageSize int) ([]model.Topic, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListTopics(ctx, page, pageSize)
}

// CreateTopic 创建话题（管理员）。
func (s *PostService) CreateTopic(ctx context.Context, name, description string) (*model.Topic, error) {
	topic := &model.Topic{Name: name, Description: description, Status: constants.TopicStatusActive, CreatedAt: time.Now()}
	if err := s.repo.CreateTopic(ctx, topic); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, util.Conflict("话题已存在", constants.ErrConflict)
		}
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	s.audit.Record(ctx, primitive.NilObjectID, "", "topic.create", "topic", topic.ID.Hex(), "创建话题", "")
	return topic, nil
}

func interactionInc(typ constants.InteractionType, delta int64) map[string]int64 {
	incs := map[string]int64{"like": 0, "comment": 0, "favorite": 0, "forward": 0}
	switch typ {
	case constants.InteractionLike:
		incs["like"] = delta
	case constants.InteractionFavorite:
		incs["favorite"] = delta
	case constants.InteractionForward:
		incs["forward"] = delta
	}
	return incs
}

func logForInteraction(typ constants.InteractionType) string {
	switch typ {
	case constants.InteractionLike:
		return constants.LogPostLiked
	case constants.InteractionFavorite:
		return constants.LogPostFavorited
	default:
		return constants.LogPostForwarded
	}
}
