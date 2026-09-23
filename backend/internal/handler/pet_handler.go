package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/dto"
	"github.com/petsocial/petsocial/internal/middleware"
	"github.com/petsocial/petsocial/internal/model"
	"github.com/petsocial/petsocial/internal/service"
	"github.com/petsocial/petsocial/internal/util"
)

// PetHandler 宠物接口层。
type PetHandler struct {
	svc    *service.PetService
	logger *slog.Logger
}

// NewPetHandler 构造注入。
func NewPetHandler(svc *service.PetService, logger *slog.Logger) *PetHandler {
	return &PetHandler{svc: svc, logger: logger}
}

// Create 创建宠物档案。
func (h *PetHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	var req dto.CreatePetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidParam, err))
		return
	}
	pet, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, dto.ToPetResponse(pet))
}

// Get 宠物详情。
func (h *PetHandler) Get(c *gin.Context) {
	id, err := dto.ParseObjectID(c.Param("id"))
	if err != nil {
		c.Error(util.Validation("宠物 ID 不合法", err))
		return
	}
	pet, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, dto.ToPetResponse(pet))
}

// Update 更新宠物档案。
func (h *PetHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	id, err := dto.ParseObjectID(c.Param("id"))
	if err != nil {
		c.Error(util.Validation("宠物 ID 不合法", err))
		return
	}
	var req dto.UpdatePetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidParam, err))
		return
	}
	pet, err := h.svc.Update(c.Request.Context(), userID, id, req)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, dto.ToPetResponse(pet))
}

// Delete 删除宠物档案。
func (h *PetHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	id, err := dto.ParseObjectID(c.Param("id"))
	if err != nil {
		c.Error(util.Validation("宠物 ID 不合法", err))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), userID, id); err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"message": "宠物档案已删除"})
}

// ListMy 我的宠物。
func (h *PetHandler) ListMy(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	page, pageSize := pageParams(c)
	pets, total, err := h.svc.ListMyPets(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	util.Page(c, mapPets(pets), total, page, pageSize)
}

// List 宠物广场（复用 PetService.ListPets）。
func (h *PetHandler) List(c *gin.Context) {
	page, pageSize := pageParams(c)
	species := c.Query("species")
	city := c.Query("city")
	pets, total, err := h.svc.ListPets(c.Request.Context(), species, city, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	util.Page(c, mapPets(pets), total, page, pageSize)
}

// Follow 关注宠物。
func (h *PetHandler) Follow(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	id, err := dto.ParseObjectID(c.Param("id"))
	if err != nil {
		c.Error(util.Validation("宠物 ID 不合法", err))
		return
	}
	if err := h.svc.FollowPet(c.Request.Context(), userID, id); err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"message": "关注宠物成功"})
}

// Unfollow 取消关注宠物。
func (h *PetHandler) Unfollow(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	id, err := dto.ParseObjectID(c.Param("id"))
	if err != nil {
		c.Error(util.Validation("宠物 ID 不合法", err))
		return
	}
	if err := h.svc.UnfollowPet(c.Request.Context(), userID, id); err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"message": "取消关注宠物成功"})
}

func mapPets(pets []model.Pet) []interface{} {
	out := make([]interface{}, 0, len(pets))
	for i := range pets {
		out = append(out, dto.ToPetResponse(&pets[i]))
	}
	return out
}
