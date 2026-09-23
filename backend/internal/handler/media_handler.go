package handler

import (
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/middleware"
	"github.com/petsocial/petsocial/internal/service"
	"github.com/petsocial/petsocial/internal/util"
)

// MediaHandler 媒体上传接口层。
type MediaHandler struct {
	svc    *service.MediaService
	logger *slog.Logger
}

// NewMediaHandler 构造注入。
func NewMediaHandler(svc *service.MediaService, logger *slog.Logger) *MediaHandler {
	return &MediaHandler{svc: svc, logger: logger}
}

// Upload 上传媒体文件。
func (h *MediaHandler) Upload(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, constants.ErrUnauthorized))
		return
	}
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.Error(util.BadRequest("请选择要上传的文件", err))
		return
	}
	defer file.Close()
	if header.Size > 50<<20 {
		c.Error(util.BadRequest("文件不能超过 50MB", nil))
		return
	}
	url, err := h.svc.Upload(c.Request.Context(), userID, file, header)
	if err != nil {
		c.Error(err)
		return
	}
	typ := "image"
	if strings.HasPrefix(strings.ToLower(filepath.Ext(header.Filename)), ".mp4") {
		typ = "video"
	}
	util.OK(c, gin.H{"url": url, "type": typ, "message": constants.MsgMediaUploaded})
}
