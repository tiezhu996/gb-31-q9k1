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

// MeetupService 同城遛狗搭子业务：约伴帖 CRUD 与报名。
type MeetupService struct {
	repo   *repository.MeetupRepository
	audit  *AuditService
	logger *slog.Logger
}

// NewMeetupService 构造注入。
func NewMeetupService(repo *repository.MeetupRepository, audit *AuditService, logger *slog.Logger) *MeetupService {
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
	meetup := &model.Meetup{
		CreatorID:    creatorID,
		Title:        req.Title,
		Description:  req.Description,
		City:         req.City,
		Location:     req.Location,
		MeetTime:     meetTime,
		MaxPeople:    req.MaxPeople,
		Status:       constants.MeetupStatusOpen,
		Participants: []model.MeetupParticipant{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := s.repo.Create(ctx, meetup); err != nil {
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogMeetupCreated, meetup.ID.Hex(), creatorID.Hex(), meetup.City))
	s.audit.Record(ctx, creatorID, "", "meetup.create", "meetup", meetup.ID.Hex(), "发布约伴帖", "")
	return meetup, nil
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

// Join 报名约伴：并发安全依赖 participants.user_id 唯一索引。
func (s *MeetupService) Join(ctx context.Context, userID, meetupID primitive.ObjectID, ip string) error {
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
