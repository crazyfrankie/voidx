package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"

	"github.com/crazyfrankie/voidx/conf"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/internal/upload/repository"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/util"
)

const urlPrefix = "http://"

type UploadFileService struct {
	repo        *repository.UploadFileRepo
	minioClient *minio.Client
	endpoint    string
	bucketName  string
}

func NewUploadFileService(repo *repository.UploadFileRepo, minioClient *minio.Client) *UploadFileService {
	return &UploadFileService{
		repo:        repo,
		minioClient: minioClient,
		endpoint:    conf.GetConf().MinIO.Endpoint[0],
		bucketName:  conf.GetConf().MinIO.BucketName["user"],
	}
}

// UploadFile 上传文件到存储服务
func (s *UploadFileService) UploadFile(ctx context.Context, fh *multipart.FileHeader, image bool) (resp.UploadFileResp, error) {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return resp.UploadFileResp{}, err
	}
	file, err := fh.Open()
	if err != nil {
		return resp.UploadFileResp{}, err
	}
	defer file.Close()

	// 检查文件大小（15MB限制）
	if fh.Size > 15*1024*1024 {
		return resp.UploadFileResp{}, errno.ErrValidate.AppendBizMessage("上传文件最大不能超过15MB")
	}

	// 1. 提取文件扩展名
	extension := strings.ToLower(filepath.Ext(fh.Filename))
	if extension != "" {
		extension = extension[1:] // 去掉点号
	}

	// 2. 验证文件类型
	if image {
		if !s.isAllowedExtension(extension, entity.AllowedImageExtensions) {
			return resp.UploadFileResp{}, errno.ErrValidate.AppendBizMessage(fmt.Sprintf("不支持的图片格式: .%s", extension))
		}
	} else {
		allowedExtensions := append(entity.AllowedDocumentExtensions, entity.AllowedImageExtensions...)
		if !s.isAllowedExtension(extension, allowedExtensions) {
			return resp.UploadFileResp{}, errno.ErrValidate.AppendBizMessage(fmt.Sprintf("不支持的文件格式: .%s", extension))
		}
	}

	key := fmt.Sprintf("%s/%s", userID.String(), fh.Filename)

	// 3. 上传文件到存储服务
	info, err := s.uploadToStorage(ctx, key, file, fh.Size)
	if err != nil {
		return resp.UploadFileResp{}, errno.ErrInternalServer.AppendBizMessage("上传文件失败: " + err.Error())
	}

	// 4. 创建上传文件记录
	uploadFile := &entity.UploadFile{
		AccountID: userID,
		Name:      fh.Filename,
		Key:       key,
		Size:      fh.Size,
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
func (s *UploadFileService) DownloadFile(ctx context.Context, key string) error {
	// TODO
	return nil
}

// uploadToStorage 上传文件到存储服务
func (s *UploadFileService) uploadToStorage(ctx context.Context, key string, reader io.Reader, size int64) (minio.UploadInfo, error) {
	info, err := s.minioClient.PutObject(ctx, s.bucketName, key, reader, size, minio.PutObjectOptions{})
	if err != nil {
		return minio.UploadInfo{}, err
	}

	return info, nil
}

// isAllowedExtension 检查扩展名是否被允许
func (s *UploadFileService) isAllowedExtension(extension string, allowedExtensions []string) bool {
	for _, allowed := range allowedExtensions {
		if extension == allowed {
			return true
		}
	}
	return false
}

func (s *UploadFileService) getURL(key string) string {
	return urlPrefix + s.endpoint + "/" + s.bucketName + "/" + key
}
