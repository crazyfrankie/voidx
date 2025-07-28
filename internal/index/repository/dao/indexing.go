package dao

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

type IndexingDao struct {
	db *gorm.DB
}

func NewIndexingDao(db *gorm.DB) *IndexingDao {
	return &IndexingDao{db: db}
}

func (d *IndexingDao) GetDocumentsByIDs(ctx context.Context, documentIDs []uuid.UUID) ([]*entity.Document, error) {
	var documents []*entity.Document
	err := d.db.WithContext(ctx).Where("id IN ?", documentIDs).Find(&documents).Error
	if err != nil {
		return nil, err
	}
	return documents, nil
}

func (d *IndexingDao) GetDocumentByID(ctx context.Context, documentID uuid.UUID) (*entity.Document, error) {
	var document entity.Document
	err := d.db.WithContext(ctx).Where("id = ?", documentID).First(&document).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &document, nil
}

func (d *IndexingDao) UpdateDocument(ctx context.Context, documentID uuid.UUID, updates map[string]any) error {
	return d.db.WithContext(ctx).Model(&entity.Document{}).Where("id = ?", documentID).Updates(updates).Error
}

func (d *IndexingDao) GetSegmentsByDocumentID(ctx context.Context, documentID uuid.UUID) ([]*entity.Segment, error) {
	var segments []*entity.Segment
	err := d.db.WithContext(ctx).Where("document_id = ?", documentID).Find(&segments).Error
	if err != nil {
		return nil, err
	}
	return segments, nil
}

func (d *IndexingDao) UpdateSegmentByNodeID(ctx context.Context, nodeID uuid.UUID, updates map[string]any) error {
	return d.db.WithContext(ctx).Model(&entity.Segment{}).Where("node_id = ?", nodeID).Updates(updates).Error
}

func (d *IndexingDao) DeleteSegmentsByDocumentID(ctx context.Context, documentID uuid.UUID) error {
	return d.db.WithContext(ctx).Where("document_id = ?", documentID).Delete(&entity.Segment{}).Error
}

func (d *IndexingDao) DeleteDocumentsByDatasetID(ctx context.Context, datasetID uuid.UUID) error {
	return d.db.WithContext(ctx).Where("dataset_id = ?", datasetID).Delete(&entity.Document{}).Error
}

func (d *IndexingDao) DeleteSegmentsByDatasetID(ctx context.Context, datasetID uuid.UUID) error {
	return d.db.WithContext(ctx).Where("dataset_id = ?", datasetID).Delete(&entity.Segment{}).Error
}

func (d *IndexingDao) DeleteKeywordTablesByDatasetID(ctx context.Context, datasetID uuid.UUID) error {
	return d.db.WithContext(ctx).Where("dataset_id = ?", datasetID).Delete(&entity.Keyword{}).Error
}

func (d *IndexingDao) DeleteDatasetQueriesByDatasetID(ctx context.Context, datasetID uuid.UUID) error {
	return d.db.WithContext(ctx).Where("dataset_id = ?", datasetID).Delete(&entity.DatasetQuery{}).Error
}

func (d *IndexingDao) GetUploadFileByID(ctx context.Context, uploadFileID uuid.UUID) (*entity.UploadFile, error) {
	var uploadFile entity.UploadFile
	err := d.db.WithContext(ctx).Where("id = ?", uploadFileID).First(&uploadFile).Error
	if err != nil {
		return nil, err
	}
	return &uploadFile, nil
}

func (d *IndexingDao) GetProcessRuleByID(ctx context.Context, processRuleID uuid.UUID) (*entity.ProcessRule, error) {
	var processRule entity.ProcessRule
	err := d.db.WithContext(ctx).Where("id = ?", processRuleID).First(&processRule).Error
	if err != nil {
		return nil, err
	}
	return &processRule, nil
}

func (d *IndexingDao) GetMaxSegmentPosition(ctx context.Context, documentID uuid.UUID) (int, error) {
	var maxPosition int
	err := d.db.WithContext(ctx).Model(&entity.Segment{}).
		Where("document_id = ?", documentID).
		Select("COALESCE(MAX(position), 0)").
		Scan(&maxPosition).Error
	if err != nil {
		return 0, err
	}
	return maxPosition, nil
}

func (d *IndexingDao) CreateSegment(ctx context.Context, segment *entity.Segment) error {
	return d.db.WithContext(ctx).Create(segment).Error
}

func (d *IndexingDao) UpdateSegment(ctx context.Context, segmentID uuid.UUID, updates map[string]any) error {
	return d.db.WithContext(ctx).Model(&entity.Segment{}).Where("id = ?", segmentID).Updates(updates).Error
}

func (d *IndexingDao) UpdateKeywordTable(ctx context.Context, keywordTableID uuid.UUID, updates map[string]any) error {
	return d.db.WithContext(ctx).Model(&entity.Keyword{}).Where("id = ?", keywordTableID).Updates(updates).Error
}
