package dao

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
)

type DocumentDao struct {
	db *gorm.DB
}

func NewDocumentDao(db *gorm.DB) *DocumentDao {
	return &DocumentDao{db: db}
}

func (d *DocumentDao) CreateDocument(ctx context.Context, document *entity.Document) error {
	return d.db.WithContext(ctx).Create(document).Error
}

func (d *DocumentDao) GetDocumentByID(ctx context.Context, id uuid.UUID) (*entity.Document, error) {
	var document entity.Document
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&document).Error
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func (d *DocumentDao) GetDocumentsByDatasetID(
	ctx context.Context,
	datasetID uuid.UUID,
	pageReq req.GetDocumentsWithPageReq,
) ([]entity.Document, int64, error) {
	var documents []entity.Document
	var total int64

	query := d.db.WithContext(ctx).Where("dataset_id = ?", datasetID)

	// 添加搜索条件
	if pageReq.SearchWord != "" {
		query = query.Where("name ILIKE ?", "%"+pageReq.SearchWord+"%")
	}

	// 计算总数
	if err := query.Model(&entity.Document{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (pageReq.CurrentPage - 1) * pageReq.PageSize
	err := query.Order("position ASC, ctime DESC").
		Offset(offset).
		Limit(pageReq.PageSize).
		Find(&documents).Error

	if err != nil {
		return nil, 0, err
	}

	return documents, total, nil
}

func (d *DocumentDao) UpdateDocument(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return d.db.WithContext(ctx).Model(&entity.Document{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (d *DocumentDao) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除相关的片段
		if err := tx.Where("document_id = ?", id).Delete(&entity.Segment{}).Error; err != nil {
			return err
		}

		// 删除文档
		return tx.Where("id = ?", id).Delete(&entity.Document{}).Error
	})
}

func (d *DocumentDao) GetNextDocumentPosition(ctx context.Context, datasetID uuid.UUID) (int, error) {
	var maxPosition int
	err := d.db.WithContext(ctx).Model(&entity.Document{}).
		Where("dataset_id = ?", datasetID).
		Select("COALESCE(MAX(position), 0)").
		Scan(&maxPosition).Error

	if err != nil {
		return 0, err
	}

	return maxPosition + 1, nil
}

func (d *DocumentDao) GetDatasetByID(ctx context.Context, id uuid.UUID) (*entity.Dataset, error) {
	var dataset entity.Dataset
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&dataset).Error
	if err != nil {
		return nil, err
	}
	return &dataset, nil
}

// GetUploadFilesByIDs 根据ID列表获取上传文件
func (d *DocumentDao) GetUploadFilesByIDs(ctx context.Context, userID uuid.UUID, fileIDs []uuid.UUID) ([]entity.UploadFile, error) {
	var files []entity.UploadFile
	err := d.db.WithContext(ctx).
		Where("id IN ? AND account_id = ?", fileIDs, userID).
		Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, nil
}

// CreateProcessRule 创建处理规则
func (d *DocumentDao) CreateProcessRule(ctx context.Context, rule *entity.ProcessRule) error {
	return d.db.WithContext(ctx).Create(rule).Error
}

// GetLatestDocumentPosition 获取最新文档位置
func (d *DocumentDao) GetLatestDocumentPosition(ctx context.Context, datasetID uuid.UUID) (int, error) {
	var maxPosition int
	err := d.db.WithContext(ctx).Model(&entity.Document{}).
		Where("dataset_id = ?", datasetID).
		Select("COALESCE(MAX(position), 0)").
		Scan(&maxPosition).Error
	if err != nil {
		return 0, err
	}
	return maxPosition, nil
}

// GetDocumentsByBatch 根据批次获取文档列表
func (d *DocumentDao) GetDocumentsByBatch(ctx context.Context, datasetID uuid.UUID, batch string) ([]entity.Document, error) {
	var documents []entity.Document
	err := d.db.WithContext(ctx).
		Where("dataset_id = ? AND batch = ?", datasetID, batch).
		Order("position ASC").
		Find(&documents).Error
	if err != nil {
		return nil, err
	}
	return documents, nil
}

// GetSegmentCountByDocument 获取文档的片段总数
func (d *DocumentDao) GetSegmentCountByDocument(ctx context.Context, documentID uuid.UUID) (int, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&entity.Segment{}).
		Where("document_id = ?", documentID).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (d *DocumentDao) GetHitCountByDocument(ctx context.Context, documentID uuid.UUID) (int, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&entity.Segment{}).
		Where("document_id = ?", documentID).
		Select("COALESCE(SUM(hit_count), 0)").
		Scan(&count).Error
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

// GetCompletedSegmentCountByDocument 获取文档的已完成片段数
func (d *DocumentDao) GetCompletedSegmentCountByDocument(ctx context.Context, documentID uuid.UUID) (int, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&entity.Segment{}).
		Where("document_id = ? AND status = ?", documentID, "completed").
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// GetUploadFileByID 根据ID获取上传文件
func (d *DocumentDao) GetUploadFileByID(ctx context.Context, fileID uuid.UUID) (*entity.UploadFile, error) {
	var file entity.UploadFile
	err := d.db.WithContext(ctx).Where("id = ?", fileID).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}
