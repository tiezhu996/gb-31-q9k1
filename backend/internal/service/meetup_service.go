package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
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
	// joinMu 串行化所有报名/取消报名：同一用户同时报名两场撞车约伴时，
	// 后执行的那场一定能读到先成功的报名记录并被拒绝，失败方不会追加参与者记录。
	joinMu sync.Mutex
}

// NormalizeDuration 发布时长归一化：未填（0）按 120 分钟，范围 30~480 分钟。
func NormalizeDuration(minutes int) (int, error) {
	if minutes == 0 {
		return constants.MeetupDurationDefault, nil
	}
	if minutes < constants.MeetupDurationMin || minutes > constants.MeetupDurationMax {
		return 0, errors.New("meetup duration out of range")
	}
	return minutes, nil
}

// timeRangesOverlap 两个半开区间 [start,end) 是否相交：首尾相接（一场结束=另一场开始）不算撞车。
func timeRangesOverlap(start1, end1, start2, end2 time.Time) bool {
	return start1.Before(end2) && start2.Before(end1)
}

// findMeetupConflict 在本人有效报名（未取消报名、约伴未取消）中查找与目标时间段撞车的约伴。
func findMeetupConflict(active []model.Meetup, targetID primitive.ObjectID, start time.Time, durationMinutes int) *model.Meetup {
	end := start.Add(time.Duration(durationMinutes) * time.Minute)
	for i := range active {
		m := &active[i]
		if m.ID == targetID {
			continue
		}
		if timeRangesOverlap(start, end, m.MeetTime, m.EndTime()) {
			return m
		}
	}
	return nil
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
	duration, err := NormalizeDuration(req.DurationMinutes)
	if err != nil {
		return nil, util.Validation(constants.MsgMeetupDuration, err)
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

// Join 报名约伴：并发安全依赖 participants.user_id 唯一索引 + joinMu 串行化。
// 报名前检查本人尚未结束或取消的约伴，时间段撞车则拒绝并指明冲突的是哪一场。
func (s *MeetupService) Join(ctx context.Context, userID, meetupID primitive.ObjectID, ip string) error {
	s.joinMu.Lock()
	defer s.joinMu.Unlock()
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
	// 撞车检查：只看本人 joined 且约伴未取消的场次；取消报名/取消约伴后时间段已释放。
	active, err := s.repo.ListActiveByParticipant(ctx, userID)
	if err != nil {
		return util.Internal(constants.MsgInternalError, err)
	}
	if conflict := findMeetupConflict(active, meetupID, meetup.MeetTime, meetup.EffectiveDuration()); conflict != nil {
		s.logger.Info(fmt.Sprintf(constants.LogMeetupConflict, meetupID.Hex(), userID.Hex(), conflict.ID.Hex(), conflict.Title))
		return util.NewAppError(
			constants.CodeMeetupConflict,
			fmt.Sprintf("报名时间与已报名的「%s」（%s ~ %s）冲突，请先取消冲突场次或改报其他时间",
				conflict.Title,
				conflict.MeetTime.Format("2006-01-02 15:04"),
				conflict.EndTime().Format("2006-01-02 15:04"),
			),
			409,
			constants.ErrMeetupConflict,
		)
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

// CancelJoin 取消报名：参与者置为 cancelled 即释放占用的时间段，满员场次回退为招募中。
func (s *MeetupService) CancelJoin(ctx context.Context, userID, meetupID primitive.ObjectID) error {
	s.joinMu.Lock()
	defer s.joinMu.Unlock()
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
// 与 Join/CancelJoin 共用 joinMu：避免取消约伴与报名并发时，整文档回写把 cancelled 状态覆盖回去。
func (s *MeetupService) UpdateStatus(ctx context.Context, userID, meetupID primitive.ObjectID, req dto.UpdateMeetupStatusRequest) error {
	s.joinMu.Lock()
	defer s.joinMu.Unlock()
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
