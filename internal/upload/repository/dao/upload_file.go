package dao

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

type UploadFileDao struct {
	db *gorm.DB
}

func NewUploadFileDao(db *gorm.DB) *UploadFileDao {
	return &UploadFileDao{db: db}
}

// CreateUploadFile 创建上传文件记录
func (d *UploadFileDao) CreateUploadFile(ctx context.Context, uploadFile *entity.UploadFile) error {
	return d.db.WithContext(ctx).Create(uploadFile).Error
}

// GetUploadFileByID 根据ID获取上传文件记录
func (d *UploadFileDao) GetUploadFileByID(ctx context.Context, id uuid.UUID) (*entity.UploadFile, error) {
	var uploadFile entity.UploadFile
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&uploadFile).Error
	if err != nil {
		return nil, err
	}
	return &uploadFile, nil
}

// GetUploadFilesByAccountID 根据账户ID获取上传文件列表
func (d *UploadFileDao) GetUploadFilesByAccountID(
	ctx context.Context,
	accountID uuid.UUID,
	page, pageSize int,
) ([]entity.UploadFile, int64, error) {
	var uploadFiles []entity.UploadFile
	var total int64

	// 获取总数
	if err := d.db.WithContext(ctx).Model(&entity.UploadFile{}).
		Where("account_id = ?", accountID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := d.db.WithContext(ctx).
		Where("account_id = ?", accountID).
		Order("ctime DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&uploadFiles).Error

	return uploadFiles, total, err
}

// DeleteUploadFile 删除上传文件记录
func (d *UploadFileDao) DeleteUploadFile(ctx context.Context, id uuid.UUID) error {
	return d.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.UploadFile{}).Error
}
