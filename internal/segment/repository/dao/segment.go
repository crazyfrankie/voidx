package dao

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/pkg/errno"
)

type SegmentDao struct {
	db *gorm.DB
}

func NewSegmentDao(db *gorm.DB) *SegmentDao {
	return &SegmentDao{db: db}
}

func (d *SegmentDao) ValidateDatasetAccess(ctx context.Context, datasetID, userID uuid.UUID) error {
	var count int64
	err := d.db.WithContext(ctx).Model(&entity.Dataset{}).
		Where("id = ? AND account_id = ?", datasetID, userID).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return errno.ErrForbidden.AppendBizMessage("无权限访问该知识库")
	}
	return nil
}

func (d *SegmentDao) ValidateDocumentAccess(ctx context.Context, documentID, datasetID uuid.UUID) error {
	var count int64
	err := d.db.WithContext(ctx).Model(&entity.Document{}).
		Where("id = ? AND dataset_id = ?", documentID, datasetID).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return errno.ErrNotFound.AppendBizMessage("文档不存在或不属于该知识库")
	}
	return nil
}

func (d *SegmentDao) CreateSegment(ctx context.Context, segment *entity.Segment) error {
	return d.db.WithContext(ctx).Create(segment).Error
}

func (d *SegmentDao) GetSegmentsByDocumentID(
	ctx context.Context,
	documentID uuid.UUID,
	pageReq req.GetSegmentsWithPageReq,
) ([]entity.Segment, int64, error) {
	var segments []entity.Segment
	var total int64

	query := d.db.WithContext(ctx).Where("document_id = ?", documentID)

	// 添加关键词搜索
	if pageReq.SearchWord != "" {
		query = query.Where("content ILIKE ? OR keywords::text ILIKE ?",
			"%"+pageReq.SearchWord+"%", "%"+pageReq.SearchWord+"%")
	}
	
	// 计算总数
	if err := query.Model(&entity.Segment{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (pageReq.CurrentPage - 1) * pageReq.PageSize
	err := query.Order("ctime DESC").
		Offset(offset).
		Limit(pageReq.PageSize).
		Find(&segments).Error

	if err != nil {
		return nil, 0, err
	}

	return segments, total, nil
}

func (d *SegmentDao) GetSegmentByID(ctx context.Context, id uuid.UUID) (*entity.Segment, error) {
	var segment entity.Segment
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&segment).Error
	if err != nil {
		return nil, err
	}
	return &segment, nil
}

func (d *SegmentDao) UpdateSegment(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return d.db.WithContext(ctx).Model(&entity.Segment{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (d *SegmentDao) DeleteSegment(ctx context.Context, id uuid.UUID) error {
	return d.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.Segment{}).Error
}

func (d *SegmentDao) GetSegments(ctx context.Context, ids []uuid.UUID) ([]entity.Segment, error) {
	res := make([]entity.Segment, 0, len(ids))
	if err := d.db.WithContext(ctx).Model(&entity.Segment{}).Where("id IN ?", ids).Find(&res).Error; err != nil {
		return nil, err
	}

	return res, nil
}

func (d *SegmentDao) GetDocument(ctx context.Context, docID uuid.UUID) (*entity.Document, error) {
	var document entity.Document
	err := d.db.WithContext(ctx).Where("id = ?", docID).First(&document).Error
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func (d *SegmentDao) GetMaxSegmentPosition(ctx context.Context, docID uuid.UUID) (int, error) {
	var position int

	err := d.db.WithContext(ctx).Model(&entity.Segment{}).
		Where("document_id = ?", docID).
		Select("COALESCE(MAX(position), 0)").
		Scan(&position).Error
	if err != nil {
		return 0, err
	}

	return position, nil
}

func (d *SegmentDao) GetDocumentSegmentCounts(ctx context.Context, docID uuid.UUID) (int, int, error) {
	var totalChars, totalTokens int64
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.Segment{}).
			Select("COALESCE(character_count)").
			Where("document_id = ?", docID).
			Scan(&totalChars).Error; err != nil {
			return err
		}

		if err := tx.Model(&entity.Segment{}).
			Select("COALESCE(token_count)").
			Where("document_id = ?", docID).
			Scan(&totalTokens).Error; err != nil {
			return err
		}

		return nil
	})

	return int(totalChars), int(totalTokens), err
}

func (d *SegmentDao) UpdateDocument(ctx context.Context, docID uuid.UUID, updates map[string]any) error {
	return d.db.WithContext(ctx).Model(&entity.Document{}).Where("id = ?", docID).Updates(updates).Error
}
