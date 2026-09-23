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

// PetService 宠物业务：档案 CRUD 与宠物粉丝。
type PetService struct {
	repo   *repository.PetRepository
	audit  *AuditService
	logger *slog.Logger
}

// NewPetService 构造注入。
func NewPetService(repo *repository.PetRepository, audit *AuditService, logger *slog.Logger) *PetService {
	return &PetService{repo: repo, audit: audit, logger: logger}
}

// Create 创建宠物档案。
func (s *PetService) Create(ctx context.Context, ownerID primitive.ObjectID, req dto.CreatePetRequest) (*model.Pet, error) {
	birthday, err := dto.ParseBirthday(req.Birthday)
	if err != nil {
		return nil, util.Validation("生日格式必须为 2006-01-02", err)
	}
	pet := &model.Pet{
		OwnerID:     ownerID,
		Name:        req.Name,
		Species:     req.Species,
		Breed:       req.Breed,
		Birthday:    birthday,
		Gender:      req.Gender,
		Personality: req.Personality,
		Bio:         req.Bio,
		Avatar:      req.Avatar,
		Album:       req.Album,
		City:        req.City,
		Status:      constants.PetStatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := s.repo.Create(ctx, pet); err != nil {
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogPetCreated, pet.ID.Hex(), ownerID.Hex(), pet.Name))
	s.audit.Record(ctx, ownerID, "", "pet.create", "pet", pet.ID.Hex(), "创建宠物档案", "")
	return pet, nil
}

// Get 获取宠物档案。
func (s *PetService) Get(ctx context.Context, petID primitive.ObjectID) (*model.Pet, error) {
	pet, err := s.repo.FindByID(ctx, petID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, util.NotFound(constants.MsgNotFound, err)
		}
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	return pet, nil
}

// Update 更新宠物档案（仅主人）。
func (s *PetService) Update(ctx context.Context, userID, petID primitive.ObjectID, req dto.UpdatePetRequest) (*model.Pet, error) {
	pet, err := s.repo.FindByID(ctx, petID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, util.NotFound(constants.MsgNotFound, err)
		}
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	if pet.OwnerID != userID {
		return nil, util.Forbidden("只有宠物主人才可编辑该档案", constants.ErrPetNotOwned)
	}
	if req.Name != "" {
		pet.Name = req.Name
	}
	if req.Breed != "" {
		pet.Breed = req.Breed
	}
	if req.Birthday != "" {
		birthday, berr := dto.ParseBirthday(req.Birthday)
		if berr != nil {
			return nil, util.Validation("生日格式必须为 2006-01-02", berr)
		}
		pet.Birthday = birthday
	}
	if req.Gender != "" {
		pet.Gender = req.Gender
	}
	if req.Personality != "" {
		pet.Personality = req.Personality
	}
	if req.Bio != "" {
		pet.Bio = req.Bio
	}
	if req.Avatar != "" {
		pet.Avatar = req.Avatar
	}
	if req.Album != nil {
		pet.Album = req.Album
	}
	if req.City != "" {
		pet.City = req.City
	}
	if req.Status != "" {
		pet.Status = req.Status
	}
	pet.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, pet); err != nil {
		return nil, util.Internal(constants.MsgInternalError, err)
	}
	s.logger.Info(fmt.Sprintf(constants.LogPetUpdated, pet.ID.Hex(), userID.Hex()))
	s.audit.Record(ctx, userID, "", "pet.update", "pet", pet.ID.Hex(), "更新宠物档案", "")
	return pet, nil
}

// Delete 删除宠物档案（仅主人）。
func (s *PetService) Delete(ctx context.Context, userID, petID primitive.ObjectID) error {
	pet, err := s.repo.FindByID(ctx, petID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return util.NotFound(constants.MsgNotFound, err)
		}
		return util.Internal(constants.MsgInternalError, err)
	}
	if pet.OwnerID != userID {
		return util.Forbidden("只有宠物主人才可删除该档案", constants.ErrPetNotOwned)
	}
	if err := s.repo.Delete(ctx, petID); err != nil {
		return util.Internal(constants.MsgInternalError, err)
	}
	s.audit.Record(ctx, userID, "", "pet.delete", "pet", petID.Hex(), "删除宠物档案", "")
	return nil
}

// ListMyPets 我的宠物列表。
func (s *PetService) ListMyPets(ctx context.Context, ownerID primitive.ObjectID, page, pageSize int) ([]model.Pet, int64, error) {
	pets, total, err := s.repo.ListByOwner(ctx, ownerID, page, pageSize)
	if err != nil {
		return nil, 0, util.Internal(constants.MsgInternalError, err)
	}
	return pets, total, nil
}

// ListPets 宠物广场列表（复用：宠物广场/同城宠物）。
func (s *PetService) ListPets(ctx context.Context, species, city string, page, pageSize int) ([]model.Pet, int64, error) {
	pets, total, err := s.repo.List(ctx, species, city, page, pageSize)
	if err != nil {
		return nil, 0, util.Internal(constants.MsgInternalError, err)
	}
	return pets, total, nil
}

// FollowPet 关注宠物。
func (s *PetService) FollowPet(ctx context.Context, userID, petID primitive.ObjectID) error {
	pet, err := s.repo.FindByID(ctx, petID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return util.NotFound(constants.MsgNotFound, err)
		}
		return util.Internal(constants.MsgInternalError, err)
	}
	if pet.OwnerID == userID {
		return util.BadRequest("不能关注自己的宠物", errors.New("cannot follow own pet"))
	}
	if err := s.repo.FollowPet(ctx, &model.PetFollow{PetID: petID, UserID: userID, CreatedAt: time.Now()}); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return util.Conflict("已关注该宠物", constants.ErrConflict)
		}
		return util.Internal(constants.MsgInternalError, err)
	}
	if err := s.repo.UpdateFollowerCount(ctx, petID, 1); err != nil {
		return util.Internal(constants.MsgInternalError, err)
	}
	s.audit.Record(ctx, userID, "", "pet.follow", "pet", petID.Hex(), "关注宠物", "")
	return nil
}

// UnfollowPet 取消关注宠物。
func (s *PetService) UnfollowPet(ctx context.Context, userID, petID primitive.ObjectID) error {
	if err := s.repo.UnfollowPet(ctx, petID, userID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return util.BadRequest("尚未关注该宠物", err)
		}
		return util.Internal(constants.MsgInternalError, err)
	}
	if err := s.repo.UpdateFollowerCount(ctx, petID, -1); err != nil {
		return util.Internal(constants.MsgInternalError, err)
	}
	s.audit.Record(ctx, userID, "", "pet.unfollow", "pet", petID.Hex(), "取消关注宠物", "")
	return nil
}
