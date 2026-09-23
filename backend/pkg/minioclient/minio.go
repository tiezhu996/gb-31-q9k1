package minioclient

import (
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client MinIO 客户端封装。
type Client struct {
	Minio     *minio.Client
	Bucket    string
	PublicURL string
}

// NewMinIO 创建 MinIO 客户端并确保 bucket 存在。
func NewMinIO(ctx context.Context, endpoint, accessKey, secretKey, bucket, publicURL string, useSSL bool) (*Client, error) {
	mc, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}
	bucketCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	exists, err := mc.BucketExists(bucketCtx, bucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket: %w", err)
	}
	if !exists {
		if err := mc.MakeBucket(bucketCtx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("make bucket: %w", err)
		}
	}
	// 设置 bucket 公开读策略，使媒体 URL 可直接访问
	policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`, bucket)
	if err := mc.SetBucketPolicy(bucketCtx, bucket, policy); err != nil {
		return nil, fmt.Errorf("set bucket policy: %w", err)
	}
	return &Client{Minio: mc, Bucket: bucket, PublicURL: publicURL}, nil
}
