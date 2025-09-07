package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"

	"github.com/crazyfrankie/voidx/infra/contract/storage"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/internal/upload/repository"
	"github.com/crazyfrankie/voidx/types/consts"
	"github.com/crazyfrankie/voidx/types/errno"
)

type OssService struct {
	repo        *repository.UploadFileRepo
	minioClient storage.Storage
}

func NewOssService(repo *repository.UploadFileRepo, minioClient storage.Storage) *OssService {
	return &OssService{
		repo:        repo,
		minioClient: minioClient,
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
			return resp.UploadFileResp{}, errno.ErrValidate.AppendBizMessage(fmt.Errorf("不支持的图片格式: .%s", extension))
		}
	} else {
		allowedExtensions := append(consts.AllowedDocumentExtension, consts.AllowedImageExtension...)
		if !s.isAllowedExtension(extension, allowedExtensions) {
			return resp.UploadFileResp{}, errno.ErrValidate.AppendBizMessage(fmt.Errorf("不支持的文件格式: .%s", extension))
		}
	}

	key := fmt.Sprintf("%s/%s", userID.String(), filename)

	info, err := s.uploadToStorage(ctx, key, data)
	if err != nil {
		return resp.UploadFileResp{}, errno.ErrInternalServer.AppendBizMessage(fmt.Errorf("上传文件失败, %w", err))
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
		if err := s.minioClient.DeleteObject(ctx, key); err != nil {
			return resp.UploadFileResp{}, err
		}
		return resp.UploadFileResp{}, err
	}

	res := resp.UploadFileResp{
		ID:        uploadFile.ID,
		AccountID: uploadFile.AccountID,
		Name:      uploadFile.Name,
		Key:       uploadFile.Key,
		Size:      uploadFile.Size,
		Extension: uploadFile.Extension,
		Ctime:     uploadFile.Ctime,
	}

	url, err := s.minioClient.GetObjectUrl(ctx, uploadFile.Key)
	if err == nil {
		res.URL = url
	}

	return res, nil
}

// DownloadFile 下载文件
func (s *OssService) DownloadFile(ctx context.Context, key string, targetPath string) error {
	if err := ensureDirExists(targetPath); err != nil {
		return fmt.Errorf("目标目录创建失败: %w", err)
	}

	data, err := s.minioClient.GetObject(ctx, key)
	if err != nil {
		return fmt.Errorf("获取文件对象失败: %w", err)
	}

	file, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("创建本地文件失败: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
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
func (s *OssService) uploadToStorage(ctx context.Context, key string, data []byte) (*minio.UploadInfo, error) {
	info, err := s.minioClient.PutObject(ctx, key, data)
	if err != nil {
		return nil, err
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
