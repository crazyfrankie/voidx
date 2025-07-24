package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/upload/repository/dao"
)

type UploadFileRepo struct {
	dao *dao.UploadFileDao
}

func NewUploadFileRepo(d *dao.UploadFileDao) *UploadFileRepo {
	return &UploadFileRepo{dao: d}
}

// CreateUploadFile 创建上传文件记录
func (r *UploadFileRepo) CreateUploadFile(ctx context.Context, uploadFile *entity.UploadFile) error {
	return r.dao.CreateUploadFile(ctx, uploadFile)
}

// GetUploadFileByID 根据ID获取上传文件记录
func (r *UploadFileRepo) GetUploadFileByID(ctx context.Context, id uuid.UUID) (*entity.UploadFile, error) {
	return r.dao.GetUploadFileByID(ctx, id)
}

// GetUploadFilesByAccountID 根据账户ID获取上传文件列表
func (r *UploadFileRepo) GetUploadFilesByAccountID(
	ctx context.Context,
	accountID uuid.UUID,
	page, pageSize int,
) ([]entity.UploadFile, int64, error) {
	return r.dao.GetUploadFilesByAccountID(ctx, accountID, page, pageSize)
}

// DeleteUploadFile 删除上传文件记录
func (r *UploadFileRepo) DeleteUploadFile(ctx context.Context, id uuid.UUID) error {
	return r.dao.DeleteUploadFile(ctx, id)
}
