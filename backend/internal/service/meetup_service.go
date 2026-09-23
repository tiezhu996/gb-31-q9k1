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
	"github.com/petsocial/petsocial/internal/util"
)

// meetupDataStore 约伴仓储接口：线上由 *repository.MeetupRepository 实现，测试可用内存假实现替换。
type meetupDataStore interface {
	Create(ctx context.Context, m *model.Meetup) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.Meetup, error)
	Update(ctx context.Context, m *model.Meetup) error
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status constants.MeetupStatus) error
	List(ctx context.Context, city, status string, page, pageSize int) ([]model.Meetup, int64, error)
	ListByCreator(ctx context.Context, creatorID primitive.ObjectID, page, pageSize int) ([]model.Meetup, int64, error)
	FindConflictingJoined(ctx context.Context, userID primitive.ObjectID, start, end time.Time, excludeID primitive.ObjectID) ([]model.Meetup, error)
	AcquireJoinLock(ctx context.Context, userID primitive.ObjectID) (bool, error)
	ReleaseJoinLock(ctx context.Context, userID primitive.ObjectID) error
}

// MeetupService 同城遛狗搭子业务：约伴帖 CRUD 与报名。
type MeetupService struct {
	repo   meetupDataStore
	audit  *AuditService
	logger *slog.Logger
}

// NewMeetupService 构造注入。
func NewMeetupService(repo meetupDataStore, audit *AuditService, logger *slog.Logger) *MeetupService {
	return &MeetupService{repo: repo, audit: audit, logger: logger}
}

// Create 发布约伴帖。
func (s *MeetupService) Create(ctx context.Context, creatorID primitive.ObjectID, req dto.CreateMeetupRequest) (*model.Meetup, error) {
	meetTime, err := dto.ParseMeetTime(req.MeetTime)
	if err != nil {
		return nil, util.Validation("约伴时间格式必须为 2006-01-02 15:04", err)
	}
	if meetTime.Before(time.Now()) {
		return nil, util.BadRequest("约伴时间不能早于当前时间", errors.New("meet time in past"))
	}
	duration := constants.MeetupDurationDefault
	if req.DurationMinutes != nil {
		duration = *req.DurationMinutes
	}
	if err := ValidateDuration(duration); err != nil {
		return nil, err
	}
	meetup := &model.Meetup{
		CreatorID:       creatorID,
		Title:           req.Title,
		Description:     req.Description,
		City:            req.City,
		Location:        req.Location,
		MeetTime:        meetTime,
		DurationMinutes: duration,
		MaxPeople:       req.MaxPeople,
		Status:          constants.MeetupStatusOpen,
		Participants:    []model.MeetupParticipant{},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if err := s.repo.Create(ctx, meetup); err != nil {
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogMeetupCreated, meetup.ID.Hex(), creatorID.Hex(), meetup.City))
	s.audit.Record(ctx, creatorID, "", "meetup.create", "meetup", meetup.ID.Hex(), "发布约伴帖", "")
	return meetup, nil
}

// ValidateDuration 校验约伴时长（分钟）：未填按 120，范围 30~480。
func ValidateDuration(minutes int) error {
	if minutes < constants.MeetupDurationMin || minutes > constants.MeetupDurationMax {
		return util.Validation(
			fmt.Sprintf("活动时长需在 %d~%d 分钟之间", constants.MeetupDurationMin, constants.MeetupDurationMax),
			constants.ErrValidation,
		)
	}
	return nil
}

// Get 约伴详情。
func (s *MeetupService) Get(ctx context.Context, meetupID primitive.ObjectID) (*model.Meetup, error) {
	meetup, err := s.repo.FindByID(ctx, meetupID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, util.NotFound(constants.MsgNotFound, err)
		}
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	return meetup, nil
}

// List 约伴列表（复用：同城约伴/全部约伴）。
func (s *MeetupService) List(ctx context.Context, city, status string, page, pageSize int) ([]model.Meetup, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.List(ctx, city, status, page, pageSize)
}

// ListMy 我的约伴。
func (s *MeetupService) ListMy(ctx context.Context, creatorID primitive.ObjectID, page, pageSize int) ([]model.Meetup, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListByCreator(ctx, creatorID, page, pageSize)
}

// joinLockWait 获取用户报名锁的最长等待与重试间隔。
const (
	joinLockWait  = 3 * time.Second
	joinLockRetry = 50 * time.Millisecond
)

// Join 报名约伴：先取用户级锁串行化本人的报名，再依次做
// 状态/重复报名/时间撞车/满员校验；任一校验失败都不会写入参与者记录。
func (s *MeetupService) Join(ctx context.Context, userID, meetupID primitive.ObjectID, ip string) error {
	if err := s.acquireJoinLock(ctx, userID); err != nil {
		return err
	}
	defer func() { _ = s.repo.ReleaseJoinLock(context.Background(), userID) }()

	// 锁内重新读取目标约伴，保证所有判断基于最新状态。
	meetup, err := s.repo.FindByID(ctx, meetupID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return util.NotFound(constants.MsgNotFound, err)
		}
		return util.Internal(constants.MsgInternalError, err)
	}
	if meetup.Status != constants.MeetupStatusOpen {
		return util.Conflict("该约伴已停止报名", constants.ErrConflict)
	}
	if meetup.CreatorID == userID {
		return util.BadRequest("发起人无需报名", errors.New("creator cannot join"))
	}
	joinedCount := 0
	for _, p := range meetup.Participants {
		if p.UserID == userID && p.Status == constants.MeetupJoinJoined {
			return util.Conflict(constants.MsgMeetupJoined, constants.ErrMeetupJoined)
		}
		if p.Status == constants.MeetupJoinJoined {
			joinedCount++
		}
	}
	// 时间撞车：本人未结束、未取消的其它约伴时间重叠即拒绝。
	start := meetup.MeetTime
	duration := dto.DurationOrDefault(meetup)
	end := start.Add(time.Duration(duration) * time.Minute)
	conflicts, err := s.repo.FindConflictingJoined(ctx, userID, start, end, meetupID)
	if err != nil {
		return util.Internal(constants.MsgInternalError, err)
	}
	if len(conflicts) > 0 {
		return util.ConflictWithCode(constants.CodeMeetupConflict, conflictMessage(&conflicts[0]), constants.ErrMeetupConflict)
	}
	if joinedCount >= meetup.MaxPeople {
		return util.Conflict(constants.MsgMeetupFull, constants.ErrMeetupFull)
	}
	meetup.Participants = append(meetup.Participants, model.MeetupParticipant{
		UserID:   userID,
		JoinedAt: time.Now(),
		Status:   constants.MeetupJoinJoined,
	})
	joinedCount++
	if joinedCount >= meetup.MaxPeople {
		meetup.Status = constants.MeetupStatusFull
	}
	meetup.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, meetup); err != nil {
		return util.Internal(constants.MsgInternalError, err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogMeetupJoined, meetupID.Hex(), userID.Hex()), "ip", ip)
	s.audit.Record(ctx, userID, "", "meetup.join", "meetup", meetupID.Hex(), "报名约伴", ip)
	return nil
}

// acquireJoinLock 在等待窗口内自旋获取用户报名锁。
func (s *MeetupService) acquireJoinLock(ctx context.Context, userID primitive.ObjectID) error {
	deadline := time.Now().Add(joinLockWait)
	for {
		locked, err := s.repo.AcquireJoinLock(ctx, userID)
		if err != nil {
			return util.Internal(constants.MsgInternalError, err)
		}
		if locked {
			return nil
		}
		if time.Now().After(deadline) {
			return util.ConflictWithCode(constants.CodeConflict, "报名处理中，请稍后重试", constants.ErrJoinLockTimeout)
		}
		select {
		case <-ctx.Done():
			return util.Internal(constants.MsgInternalError, ctx.Err())
		case <-time.After(joinLockRetry):
		}
	}
}

// conflictMessage 生成撞车提示，说明与哪场约伴冲突。
func conflictMessage(c *model.Meetup) string {
	return fmt.Sprintf("时间与已报名的约伴「%s」（%s）冲突", c.Title, c.MeetTime.Format("2006-01-02 15:04"))
}

// CancelJoin 取消报名。
func (s *MeetupService) CancelJoin(ctx context.Context, userID, meetupID primitive.ObjectID) error {
	meetup, err := s.repo.FindByID(ctx, meetupID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return util.NotFound(constants.MsgNotFound, err)
		}
		return util.Internal(constants.MsgInternalError, err)
	}
	found := false
	for i, p := range meetup.Participants {
		if p.UserID == userID && p.Status == constants.MeetupJoinJoined {
			meetup.Participants[i].Status = constants.MeetupJoinCancelled
			found = true
		}
	}
	if !found {
		return util.BadRequest("您尚未报名该约伴", constants.ErrNotFound)
	}
	if meetup.Status == constants.MeetupStatusFull {
		meetup.Status = constants.MeetupStatusOpen
	}
	meetup.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, meetup); err != nil {
		return util.Internal(constants.MsgInternalError, err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogMeetupCancelled, meetupID.Hex(), userID.Hex()))
	s.audit.Record(ctx, userID, "", "meetup.cancel_join", "meetup", meetupID.Hex(), "取消报名", "")
	return nil
}

// UpdateStatus 状态流转（发起人取消/结束）。
func (s *MeetupService) UpdateStatus(ctx context.Context, userID, meetupID primitive.ObjectID, req dto.UpdateMeetupStatusRequest) error {
	meetup, err := s.repo.FindByID(ctx, meetupID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return util.NotFound(constants.MsgNotFound, err)
		}
		return util.Internal(constants.MsgInternalError, err)
	}
	if meetup.CreatorID != userID {
		return util.Forbidden("只有发起人可变更约伴状态", constants.ErrForbidden)
	}
	if err := s.repo.UpdateStatus(ctx, meetupID, req.Status); err != nil {
		return util.Internal(constants.MsgInternalError, err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogMeetupStatus, meetupID.Hex(), req.Status))
	s.audit.Record(ctx, userID, "", "meetup.update_status", "meetup", meetupID.Hex(), "变更状态: "+string(req.Status), "")
	return nil
}

// ensureNoDuplicate 判断参与者是否已加入（供测试引用）。
func (s *MeetupService) ensureNoDuplicate(meetup *model.Meetup, userID primitive.ObjectID) bool {
	for _, p := range meetup.Participants {
		if p.UserID == userID && p.Status == constants.MeetupJoinJoined {
			return false
		}
	}
	return true
}

// isFull 判断约伴是否已满（供测试引用）。
func (s *MeetupService) isFull(meetup *model.Meetup) bool {
	count := 0
	for _, p := range meetup.Participants {
		if p.Status == constants.MeetupJoinJoined {
			count++
		}
	}
	return count >= meetup.MaxPeople
}

// strings 包引用占位，避免误删 import。
var _ = strings.TrimSpace
