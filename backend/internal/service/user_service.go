package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/dto"
	"github.com/petsocial/petsocial/internal/model"
	"github.com/petsocial/petsocial/internal/repository"
	"github.com/petsocial/petsocial/internal/util"
)

// UserService 用户业务：注册/登录/资料/关注。
type UserService struct {
	repo      *repository.UserRepository
	audit     *AuditService
	logger    *slog.Logger
	jwtSecret string
	jwtExpire int
}

// NewUserService 构造注入。
func NewUserService(repo *repository.UserRepository, audit *AuditService, logger *slog.Logger, jwtSecret string, jwtExpire int) *UserService {
	return &UserService{repo: repo, audit: audit, logger: logger, jwtSecret: jwtSecret, jwtExpire: jwtExpire}
}

// Register 注册用户。
func (s *UserService) Register(ctx context.Context, req dto.RegisterRequest, ip string) (*dto.TokenResponse, error) {
	if _, err := s.repo.FindByUsername(ctx, req.Username); err == nil {
		return nil, util.Conflict(constants.MsgUserExists, constants.ErrUserExists)
	} else if !errors.Is(err, constants.ErrNotFound) {
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	hash, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	user := &model.User{
		Username:     req.Username,
		PasswordHash: hash,
		Nickname:     req.Nickname,
		City:         req.City,
		Role:         constants.RoleUser,
		Status:       constants.UserStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogUserRegistered, user.ID.Hex(), user.Username), "ip", ip)
	s.audit.Record(ctx, user.ID, user.Username, "user.register", "user", user.ID.Hex(), "注册账号", ip)
	token, err := util.GenerateToken(s.jwtSecret, s.jwtExpire, user.ID, user.Username, user.Role)
	if err != nil {
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogTokenIssued, user.ID.Hex(), user.Role))
	return &dto.TokenResponse{Token: token, User: dto.ToUserResponse(user)}, nil
}

// Login 登录。
func (s *UserService) Login(ctx context.Context, req dto.LoginRequest, ip string) (*dto.TokenResponse, error) {
	user, err := s.repo.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, util.Unauthorized(constants.MsgWrongPassword, constants.ErrWrongPassword)
		}
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	if !util.CheckPassword(user.PasswordHash, req.Password) {
		return nil, util.Unauthorized(constants.MsgWrongPassword, constants.ErrWrongPassword)
	}
	if user.Status == constants.UserStatusBanned {
		return nil, util.Forbidden(constants.MsgUserBanned, constants.ErrUserBanned)
	}
	s.logger.Info(fmt.Sprintf(constants.LogUserLogin, user.ID.Hex(), user.Username), "ip", ip)
	s.audit.Record(ctx, user.ID, user.Username, "user.login", "user", user.ID.Hex(), "登录系统", ip)
	token, err := util.GenerateToken(s.jwtSecret, s.jwtExpire, user.ID, user.Username, user.Role)
	if err != nil {
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogTokenIssued, user.ID.Hex(), user.Role))
	return &dto.TokenResponse{Token: token, User: dto.ToUserResponse(user)}, nil
}

// GetProfile 获取用户资料。
func (s *UserService) GetProfile(ctx context.Context, userID primitive.ObjectID) (*model.User, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, util.NotFound(constants.MsgNotFound, err)
		}
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	return user, nil
}

// UpdateProfile 更新资料。
func (s *UserService) UpdateProfile(ctx context.Context, userID primitive.ObjectID, req dto.UpdateProfileRequest) (*model.User, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, util.NotFound(constants.MsgNotFound, err)
		}
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}
	if req.City != "" {
		user.City = req.City
	}
	user.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	s.audit.Record(ctx, user.ID, user.Username, "user.update_profile", "user", user.ID.Hex(), "更新资料", "")
	return user, nil
}

// Follow 关注用户：事务内写入关注关系并更新双方计数。
func (s *UserService) Follow(ctx context.Context, followerID, followeeID primitive.ObjectID, ip string) error {
	if followerID == followeeID {
		return util.BadRequest("不能关注自己", errors.New("cannot follow self"))
	}
	if _, err := s.repo.FindByID(ctx, followeeID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return util.NotFound(constants.MsgNotFound, err)
		}
		return util.Internal(constants.MsgInternalError, err)
	}
	follow := &model.Follow{FollowerID: followerID, FolloweeID: followeeID, CreatedAt: time.Now()}
	if err := s.repo.CreateFollow(ctx, follow); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return util.Conflict("已关注该用户", constants.ErrConflict)
		}
		return util.Internal(constants.MsgInternalError, err)
	}
	if err := s.repo.UpdateCounts(ctx, followerID, 1, 0); err != nil {
		return util.Internal(constants.MsgInternalError, err)
	}
	if err := s.repo.UpdateCounts(ctx, followeeID, 0, 1); err != nil {
		return util.Internal(constants.MsgInternalError, err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogUserFollow, followerID.Hex(), followeeID.Hex()), "ip", ip)
	s.audit.Record(ctx, followerID, "", "user.follow", "user", followeeID.Hex(), "关注用户", ip)
	return nil
}

// Unfollow 取消关注。
func (s *UserService) Unfollow(ctx context.Context, followerID, followeeID primitive.ObjectID, ip string) error {
	if err := s.repo.DeleteFollow(ctx, followerID, followeeID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return util.BadRequest("尚未关注该用户", err)
		}
		return util.Internal(constants.MsgInternalError, err)
	}
	if err := s.repo.UpdateCounts(ctx, followerID, -1, 0); err != nil {
		return util.Internal(constants.MsgInternalError, err)
	}
	if err := s.repo.UpdateCounts(ctx, followeeID, 0, -1); err != nil {
		return util.Internal(constants.MsgInternalError, err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogUserUnfollow, followerID.Hex(), followeeID.Hex()), "ip", ip)
	s.audit.Record(ctx, followerID, "", "user.unfollow", "user", followeeID.Hex(), "取消关注", ip)
	return nil
}

// ListFollowing 关注列表。
func (s *UserService) ListFollowing(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]model.User, int64, error) {
	ids, total, err := s.repo.ListFollowingIDs(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, util.Internal(constants.MsgInternalError, err)
	}
	users, err := s.repo.ListByIDs(ctx, ids)
	if err != nil {
		return nil, 0, util.Internal(constants.MsgInternalError, err)
	}
	return users, total, nil
}

// ListFollowers 粉丝列表。
func (s *UserService) ListFollowers(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]model.User, int64, error) {
	ids, total, err := s.repo.ListFollowerIDs(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, util.Internal(constants.MsgInternalError, err)
	}
	users, err := s.repo.ListByIDs(ctx, ids)
	if err != nil {
		return nil, 0, util.Internal(constants.MsgInternalError, err)
	}
	return users, total, nil
}

// IsFollowing 查询关注状态。
func (s *UserService) IsFollowing(ctx context.Context, followerID, followeeID primitive.ObjectID) (bool, error) {
	return s.repo.IsFollowing(ctx, followerID, followeeID)
}
