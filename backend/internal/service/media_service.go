package service

import (
	"context"
	"fmt"
	"log/slog"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/minio/minio-go/v7"
	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/util"
	minioclient "github.com/petsocial/petsocial/pkg/minioclient"
)

// MediaService 媒体上传业务：图片/视频存储到 MinIO。
type MediaService struct {
	minio  *minioclient.Client
	audit  *AuditService
	logger *slog.Logger
}

// NewMediaService 构造注入。
func NewMediaService(mc *minioclient.Client, audit *AuditService, logger *slog.Logger) *MediaService {
	return &MediaService{minio: mc, audit: audit, logger: logger}
}

// Upload 上传媒体文件，返回公开 URL。
func (s *MediaService) Upload(ctx context.Context, userID primitive.ObjectID, file multipart.File, header *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExt(ext) {
		return "", util.Validation("仅支持 jpg/jpeg/png/gif/webp/mp4 文件", constants.ErrInvalidMediaType)
	}
	objectKey := fmt.Sprintf("media/%s/%s%s", time.Now().Format("2006/01/02"), uuid.NewString(), ext)
	_, err := s.minio.Minio.PutObject(ctx, s.minio.Bucket, objectKey, file, header.Size, minioPutObjectOptions(header.Header.Get("Content-Type")))
	if err != nil {
		return "", util.Internal("媒体上传失败", err)
	}
	url := fmt.Sprintf("%s/%s/%s", s.minio.PublicURL, s.minio.Bucket, objectKey)
	s.logger.Info(fmt.Sprintf(constants.LogMediaUploaded, objectKey, header.Size))
	s.audit.Record(ctx, userID, "", "media.upload", "media", objectKey, "上传媒体", "")
	return url, nil
}

func allowedExt(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".mp4":
		return true
	default:
		return false
	}
}

func minioPutObjectOptions(contentType string) minio.PutObjectOptions {
	opts := minio.PutObjectOptions{}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	opts.ContentType = contentType
	return opts
}
