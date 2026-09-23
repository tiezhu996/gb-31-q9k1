package service

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/dto"
	"github.com/petsocial/petsocial/internal/model"
	"github.com/petsocial/petsocial/internal/repository"
	"github.com/petsocial/petsocial/internal/util"
)

// FeedService Feed 推荐业务：关注页/发现页/同城页。
type FeedService struct {
	posts  *repository.PostRepository
	users  *repository.UserRepository
	logger *slog.Logger
}

// NewFeedService 构造注入。
func NewFeedService(posts *repository.PostRepository, users *repository.UserRepository, logger *slog.Logger) *FeedService {
	return &FeedService{posts: posts, users: users, logger: logger}
}

// FollowingFeed 关注页 Feed：复用 PostRepository.ListByIDs。
func (s *FeedService) FollowingFeed(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]model.Post, int64, error) {
	ids, _, err := s.users.ListFollowingIDs(ctx, userID, 1, 500)
	if err != nil {
		return nil, 0, util.Internal(constants.MsgInternalError, err)
	}
	posts, total, err := s.posts.ListByIDs(ctx, ids, page, pageSize)
	if err != nil {
		return nil, 0, util.Internal(constants.MsgInternalError, err)
	}
	return posts, total, nil
}

// DiscoverFeed 发现页 Feed：复用 PostRepository.List。
func (s *FeedService) DiscoverFeed(ctx context.Context, page, pageSize int) ([]model.Post, int64, error) {
	posts, total, err := s.posts.List(ctx, bson.M{}, page, pageSize)
	if err != nil {
		return nil, 0, util.Internal(constants.MsgInternalError, err)
	}
	s.logger.Info("discover feed fetched", "total", total)
	return posts, total, nil
}

// NearbyFeed 同城页 Feed：复用 PostRepository.ListByCity。
func (s *FeedService) NearbyFeed(ctx context.Context, city string, page, pageSize int) ([]model.Post, int64, error) {
	if city == "" {
		city = "上海"
	}
	posts, total, err := s.posts.ListByCity(ctx, city, page, pageSize)
	if err != nil {
		return nil, 0, util.Internal(constants.MsgInternalError, err)
	}
	return posts, total, nil
}

// EnrichPosts 组装带作者信息与互动状态的动态出参（复用：feed 三页与 post 详情）。
func (s *FeedService) EnrichPosts(ctx context.Context, posts []model.Post, currentUserID primitive.ObjectID, userCache map[primitive.ObjectID]*model.User) ([]dto.PostResponse, error) {
	responses := make([]dto.PostResponse, 0, len(posts))
	for i := range posts {
		post := &posts[i]
		author, ok := userCache[post.AuthorID]
		if !ok {
			var err error
			author, err = s.users.FindByID(ctx, post.AuthorID)
			if err != nil {
				return nil, err
			}
			userCache[post.AuthorID] = author
		}
		resp := dto.ToPostResponse(post)
		resp.AuthorName = author.Nickname
		resp.AuthorAvatar = author.Avatar
		liked, _ := s.posts.HasInteraction(ctx, post.ID, currentUserID, constants.InteractionLike)
		favorited, _ := s.posts.HasInteraction(ctx, post.ID, currentUserID, constants.InteractionFavorite)
		resp.Liked = liked
		resp.Favorited = favorited
		responses = append(responses, resp)
	}
	return responses, nil
}
