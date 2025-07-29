package dao

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
)

type DatasetDao struct {
	db *gorm.DB
}

func NewDatasetDao(db *gorm.DB) *DatasetDao {
	return &DatasetDao{db: db}
}

func (d *DatasetDao) CreateDataset(ctx context.Context, dataset *entity.Dataset) error {
	return d.db.WithContext(ctx).Create(dataset).Error
}

func (d *DatasetDao) GetDatasetByID(ctx context.Context, id uuid.UUID) (*entity.Dataset, error) {
	var dataset entity.Dataset
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&dataset).Error
	if err != nil {
		return nil, err
	}
	return &dataset, nil
}

func (d *DatasetDao) GetDatasetsByAccountID(
	ctx context.Context,
	accountID uuid.UUID,
	pageReq req.GetDatasetsWithPageReq,
) ([]entity.Dataset, int64, error) {
	var datasets []entity.Dataset
	var total int64

	query := d.db.WithContext(ctx).Where("account_id = ?", accountID)

	// 添加搜索条件
	if pageReq.SearchWord != "" {
		query = query.Where("name ILIKE ?", "%"+pageReq.SearchWord+"%")
	}

	// 计算总数
	if err := query.Model(&entity.Dataset{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (pageReq.CurrentPage - 1) * pageReq.PageSize
	err := query.Order("ctime DESC").
		Offset(offset).
		Limit(pageReq.PageSize).
		Find(&datasets).Error

	if err != nil {
		return nil, 0, err
	}

	return datasets, total, nil
}

func (d *DatasetDao) UpdateDataset(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return d.db.WithContext(ctx).Model(&entity.Dataset{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (d *DatasetDao) DeleteDataset(ctx context.Context, id uuid.UUID) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除相关的文档和片段
		if err := tx.Where("dataset_id = ?", id).Delete(&entity.Document{}).Error; err != nil {
			return err
		}
		if err := tx.Where("dataset_id = ?", id).Delete(&entity.Segment{}).Error; err != nil {
			return err
		}
		if err := tx.Where("dataset_id = ?", id).Delete(&entity.DatasetQuery{}).Error; err != nil {
			return err
		}

		// 删除知识库
		return tx.Where("id = ?", id).Delete(&entity.Dataset{}).Error
	})
}

func (d *DatasetDao) GetDatasetQueries(ctx context.Context, datasetID uuid.UUID, limit int) ([]entity.DatasetQuery, error) {
	var queries []entity.DatasetQuery
	err := d.db.WithContext(ctx).
		Where("dataset_id = ?", datasetID).
		Order("ctime DESC").
		Limit(limit).
		Find(&queries).Error
	return queries, err
}

func (d *DatasetDao) CreateDatasetQuery(ctx context.Context, query *entity.DatasetQuery) error {
	return d.db.WithContext(ctx).Create(query).Error
}

// Document related methods
func (d *DatasetDao) CreateDocument(ctx context.Context, document *entity.Document) error {
	return d.db.WithContext(ctx).Create(document).Error
}

func (d *DatasetDao) GetDocumentByID(ctx context.Context, id uuid.UUID) (*entity.Document, error) {
	var document entity.Document
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&document).Error
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func (d *DatasetDao) GetDocumentsByDatasetID(
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

	// 添加状态过滤
	if pageReq.Status != "" {
		query = query.Where("indexing_status = ?", pageReq.Status)
	}

	// 计算总数
	if err := query.Model(&entity.Document{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (pageReq.Page - 1) * pageReq.PageSize
	err := query.Order("position ASC, ctime DESC").
		Offset(offset).
		Limit(pageReq.PageSize).
		Find(&documents).Error

	if err != nil {
		return nil, 0, err
	}

	return documents, total, nil
}

func (d *DatasetDao) UpdateDocument(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return d.db.WithContext(ctx).Model(&entity.Document{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (d *DatasetDao) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除相关的片段
		if err := tx.Where("document_id = ?", id).Delete(&entity.Segment{}).Error; err != nil {
			return err
		}

		// 删除文档
		return tx.Where("id = ?", id).Delete(&entity.Document{}).Error
	})
}

// Segment related methods
func (d *DatasetDao) CreateSegment(ctx context.Context, segment *entity.Segment) error {
	return d.db.WithContext(ctx).Create(segment).Error
}

func (d *DatasetDao) GetSegmentByID(ctx context.Context, id uuid.UUID) (*entity.Segment, error) {
	var segment entity.Segment
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&segment).Error
	if err != nil {
		return nil, err
	}
	return &segment, nil
}

func (d *DatasetDao) GetSegmentsByDocumentID(
	ctx context.Context,
	documentID uuid.UUID,
	pageReq req.GetSegmentsWithPageReq,
) ([]entity.Segment, int64, error) {
	var segments []entity.Segment
	var total int64

	query := d.db.WithContext(ctx).Where("document_id = ?", documentID)

	// 添加搜索条件
	if pageReq.SearchWord != "" {
		query = query.Where("content ILIKE ?", "%"+pageReq.SearchWord+"%")
	}

	// 添加状态过滤
	if pageReq.Status != "" {
		query = query.Where("status = ?", pageReq.Status)
	}

	// 计算总数
	if err := query.Model(&entity.Segment{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (pageReq.Page - 1) * pageReq.PageSize
	err := query.Order("position ASC").
		Offset(offset).
		Limit(pageReq.PageSize).
		Find(&segments).Error

	if err != nil {
		return nil, 0, err
	}

	return segments, total, nil
}

func (d *DatasetDao) UpdateSegment(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return d.db.WithContext(ctx).Model(&entity.Segment{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (d *DatasetDao) DeleteSegment(ctx context.Context, id uuid.UUID) error {
	return d.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.Segment{}).Error
}

// Statistics related methods

// GetDocumentCount 获取知识库下的文档数
func (d *DatasetDao) GetDocumentCount(ctx context.Context, datasetID uuid.UUID) (int, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&entity.Document{}).
		Where("dataset_id = ?", datasetID).
		Count(&count).Error
	return int(count), err
}

// GetHitCount 获取该知识库的命中次数
func (d *DatasetDao) GetHitCount(ctx context.Context, datasetID uuid.UUID) (int, error) {
	var totalHitCount int64
	err := d.db.WithContext(ctx).Model(&entity.Segment{}).
		Where("dataset_id = ?", datasetID).
		Select("COALESCE(SUM(hit_count), 0)").
		Scan(&totalHitCount).Error
	return int(totalHitCount), err
}

// GetRelatedAppCount 获取该知识库关联的应用数
func (d *DatasetDao) GetRelatedAppCount(ctx context.Context, datasetID uuid.UUID) (int, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&entity.AppDatasetJoin{}).
		Where("dataset_id = ?", datasetID).
		Count(&count).Error
	return int(count), err
}

// GetCharacterCount 获取该知识库下的字符总数
func (d *DatasetDao) GetCharacterCount(ctx context.Context, datasetID uuid.UUID) (int, error) {
	var totalCharacterCount int64
	err := d.db.WithContext(ctx).Model(&entity.Document{}).
		Where("dataset_id = ?", datasetID).
		Select("COALESCE(SUM(character_count), 0)").
		Scan(&totalCharacterCount).Error
	return int(totalCharacterCount), err
}
