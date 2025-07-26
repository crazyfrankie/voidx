package dao

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

// KeywordDao 关键词表存储库
type KeywordDao struct {
	db *gorm.DB
}

func NewKeywordDao(db *gorm.DB) *KeywordDao {
	return &KeywordDao{
		db: db,
	}
}

// GetByDatasetID 根据数据集ID获取关键词表
func (d *KeywordDao) GetByDatasetID(ctx context.Context, datasetID uuid.UUID) (*entity.Keyword, error) {
	var keywordTable entity.Keyword
	result := d.db.WithContext(ctx).Where("dataset_id = ?", datasetID).First(&keywordTable)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return &keywordTable, nil
}

// Create 创建关键词表
func (d *KeywordDao) Create(ctx context.Context, keywordTable *entity.Keyword) error {
	return d.db.WithContext(ctx).Create(keywordTable).Error
}

// Update 更新关键词表
func (d *KeywordDao) Update(ctx context.Context, keywordTable *entity.Keyword) error {
	return d.db.WithContext(ctx).Save(keywordTable).Error
}

func (d *KeywordDao) GetKeywordBySegments(ctx context.Context, segmentIDs []uuid.UUID) ([]entity.Segment, error) {
	var segments []entity.Segment

	if err := d.db.WithContext(ctx).Model(&entity.Segment{}).
		Where("id IN ?", segmentIDs).
		Select("id, keywords").
		Scan(&segments).Error; err != nil {
		return nil, err
	}

	return segments, nil
}
