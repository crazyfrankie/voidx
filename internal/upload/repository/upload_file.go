package repository

import (
	"context"

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
