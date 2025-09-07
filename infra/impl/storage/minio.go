package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"golang.org/x/sync/errgroup"

	"github.com/crazyfrankie/voidx/infra/contract/storage"
	"github.com/crazyfrankie/voidx/pkg/logs"
)

type minioStore struct {
	bucketName string
	client     *minio.Client
}

func New(cli *minio.Client, bucketName string) storage.Storage {
	return &minioStore{client: cli, bucketName: bucketName}
}

func (m *minioStore) PutObject(ctx context.Context, objectKey string, content []byte, opts ...storage.PutOptFn) (*minio.UploadInfo, error) {
	option := storage.PutOption{}
	for _, opt := range opts {
		opt(&option)
	}

	minioOpts := minio.PutObjectOptions{}
	if option.ContentType != nil {
		minioOpts.ContentType = *option.ContentType
	}

	if option.ContentEncoding != nil {
		minioOpts.ContentEncoding = *option.ContentEncoding
	}

	if option.ContentDisposition != nil {
		minioOpts.ContentDisposition = *option.ContentDisposition
	}

	if option.ContentLanguage != nil {
		minioOpts.ContentLanguage = *option.ContentLanguage
	}

	if option.Expires != nil {
		minioOpts.Expires = *option.Expires
	}

	info, err := m.client.PutObject(ctx, m.bucketName, objectKey,
		bytes.NewReader(content), int64(len(content)), minioOpts)
	if err != nil {
		return nil, fmt.Errorf("PutObject failed: %v", err)
	}
	return &info, nil
}

func (m *minioStore) GetObject(ctx context.Context, objectKey string) ([]byte, error) {
	obj, err := m.client.GetObject(ctx, m.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("GetObject failed: %v", err)
	}
	defer obj.Close()
	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("ReadObject failed: %v", err)
	}
	return data, nil
}

func (m *minioStore) DeleteObject(ctx context.Context, objectKey string) error {
	err := m.client.RemoveObject(ctx, m.bucketName, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("DeleteObject failed: %v", err)
	}
	return nil
}

func (m *minioStore) GetObjectUrl(ctx context.Context, objectKey string, opts ...storage.GetOptFn) (string, error) {
	option := storage.GetOption{}
	for _, opt := range opts {
		opt(&option)
	}

	if option.Expire == 0 {
		option.Expire = 3600 * 24 * 7
	}

	reqParams := make(url.Values)
	presignedURL, err := m.client.PresignedGetObject(ctx, m.bucketName, objectKey, time.Duration(option.Expire)*time.Second, reqParams)
	if err != nil {
		return "", fmt.Errorf("GetObjectUrl failed: %v", err)
	}

	return presignedURL.String(), nil
}

func (m *minioStore) BatchGetUrls(ctx context.Context, uris []string) ([]string, error) {
	if len(uris) == 0 {
		return nil, nil
	}

	urls := make([]string, len(uris))
	eg, ctx := errgroup.WithContext(ctx)

	for i := range uris {
		i := i
		eg.Go(func() error {
			u, err := m.GetObjectUrl(ctx, uris[i])
			if err != nil {
				logs.Warnf("get oss url failed, uri=%s, err=%v", uris[i], err)
				return nil
			}
			urls[i] = u
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return urls, err
	}
	return urls, nil
}
