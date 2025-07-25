package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/crazyfrankie/voidx/pkg/consts"
	"github.com/google/uuid"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"

	"github.com/crazyfrankie/voidx/conf"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/internal/upload/repository"
	"github.com/crazyfrankie/voidx/pkg/errno"
)

const urlPrefix = "http://"

type OssService struct {
	repo        *repository.UploadFileRepo
	minioClient *minio.Client
	endpoint    string
	bucketName  string
}

func NewOssService(repo *repository.UploadFileRepo, minioClient *minio.Client) *OssService {
	return &OssService{
		repo:        repo,
		minioClient: minioClient,
		endpoint:    conf.GetConf().MinIO.Endpoint[0],
		bucketName:  conf.GetConf().MinIO.BucketName["user"],
	}
}

// UploadFile 上传文件到存储服务
func (s *OssService) UploadFile(ctx context.Context, data []byte, onlyImage bool, filename string, userID uuid.UUID) (resp.UploadFileResp, error) {
	extension := strings.ToLower(filepath.Ext(filename))
	if extension != "" {
		extension = extension[1:]
	}

	if onlyImage {
		if !s.isAllowedExtension(extension, consts.AllowedImageExtension) {
			return resp.UploadFileResp{}, errno.ErrValidate.AppendBizMessage(fmt.Sprintf("不支持的图片格式: .%s", extension))
		}
	} else {
		allowedExtensions := append(consts.AllowedDocumentExtension, consts.AllowedImageExtension...)
		if !s.isAllowedExtension(extension, allowedExtensions) {
			return resp.UploadFileResp{}, errno.ErrValidate.AppendBizMessage(fmt.Sprintf("不支持的文件格式: .%s", extension))
		}
	}

	key := fmt.Sprintf("%s/%s", userID.String(), filename)

	info, err := s.uploadToStorage(ctx, key, data)
	if err != nil {
		return resp.UploadFileResp{}, errno.ErrInternalServer.AppendBizMessage("上传文件失败: " + err.Error())
	}

	uploadFile := &entity.UploadFile{
		AccountID: userID,
		Name:      filename,
		Key:       key,
		Size:      int64(len(data)),
		Extension: extension,
		Hash:      info.ETag,
	}

	err = s.repo.CreateUploadFile(ctx, uploadFile)
	if err != nil {
		// 如果数据库保存失败，尝试删除已上传的文件
		if err := s.minioClient.RemoveObject(ctx, s.bucketName, key, minio.RemoveObjectOptions{}); err != nil {
			return resp.UploadFileResp{}, err
		}
		return resp.UploadFileResp{}, err
	}

	return resp.UploadFileResp{
		ID:        uploadFile.ID,
		AccountID: uploadFile.AccountID,
		Name:      uploadFile.Name,
		Key:       uploadFile.Key,
		Size:      uploadFile.Size,
		Extension: uploadFile.Extension,
		URL:       s.getURL(uploadFile.Key),
		Ctime:     uploadFile.Ctime,
	}, nil
}

// DownloadFile 下载文件
func (s *OssService) DownloadFile(ctx context.Context, key string, targetPath string) error {
	if err := ensureDirExists(targetPath); err != nil {
		return fmt.Errorf("目标目录创建失败: %w", err)
	}

	object, err := s.minioClient.GetObject(
		ctx,
		s.bucketName,
		key,
		minio.GetObjectOptions{},
	)
	if err != nil {
		return fmt.Errorf("获取文件对象失败: %w", err)
	}
	defer object.Close()

	file, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("创建本地文件失败: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, object); err != nil {
		_ = os.Remove(targetPath)
		return fmt.Errorf("文件下载失败: %w", err)
	}

	return nil
}

func ensureDirExists(path string) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

// uploadToStorage 上传文件到存储服务
func (s *OssService) uploadToStorage(ctx context.Context, key string, data []byte) (minio.UploadInfo, error) {
	info, err := s.minioClient.PutObject(ctx, s.bucketName, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{})
	if err != nil {
		return minio.UploadInfo{}, err
	}

	return info, nil
}

// isAllowedExtension 检查扩展名是否被允许
func (s *OssService) isAllowedExtension(extension string, allowedExtensions []string) bool {
	for _, allowed := range allowedExtensions {
		if extension == allowed {
			return true
		}
	}
	return false
}

func (s *OssService) getURL(key string) string {
	return urlPrefix + s.endpoint + "/" + s.bucketName + "/" + key
}
