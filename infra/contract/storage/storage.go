package storage

import (
	"context"

	"github.com/minio/minio-go/v7"
)

type Storage interface {
	PutObject(ctx context.Context, objectKey string, content []byte, opts ...PutOptFn) (*minio.UploadInfo, error)
	GetObject(ctx context.Context, objectKey string) ([]byte, error)
	DeleteObject(ctx context.Context, objectKey string) error
	GetObjectUrl(ctx context.Context, objectKey string, opts ...GetOptFn) (string, error)
	BatchGetUrls(ctx context.Context, uris []string) ([]string, error)
}
