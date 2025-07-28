package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/index/repository/dao"
	"github.com/crazyfrankie/voidx/internal/models/entity"
)

type IndexingRepo struct {
	dao *dao.IndexingDao
}

func NewIndexingRepo(dao *dao.IndexingDao) *IndexingRepo {
	return &IndexingRepo{dao: dao}
}

func (r *IndexingRepo) GetDocumentsByIDs(ctx context.Context, documentIDs []uuid.UUID) ([]*entity.Document, error) {
	return r.dao.GetDocumentsByIDs(ctx, documentIDs)
}

func (r *IndexingRepo) GetDocumentByID(ctx context.Context, documentID uuid.UUID) (*entity.Document, error) {
	return r.dao.GetDocumentByID(ctx, documentID)
}

func (r *IndexingRepo) UpdateDocument(ctx context.Context, documentID uuid.UUID, updates map[string]any) error {
	return r.dao.UpdateDocument(ctx, documentID, updates)
}

func (r *IndexingRepo) GetSegmentsByDocumentID(ctx context.Context, documentID uuid.UUID) ([]*entity.Segment, error) {
	return r.dao.GetSegmentsByDocumentID(ctx, documentID)
}

func (r *IndexingRepo) UpdateSegmentByNodeID(ctx context.Context, nodeID uuid.UUID, updates map[string]any) error {
	return r.dao.UpdateSegmentByNodeID(ctx, nodeID, updates)
}

func (r *IndexingRepo) DeleteSegmentsByDocumentID(ctx context.Context, documentID uuid.UUID) error {
	return r.dao.DeleteSegmentsByDocumentID(ctx, documentID)
}

func (r *IndexingRepo) DeleteDocumentsByDatasetID(ctx context.Context, datasetID uuid.UUID) error {
	return r.dao.DeleteDocumentsByDatasetID(ctx, datasetID)
}

func (r *IndexingRepo) DeleteSegmentsByDatasetID(ctx context.Context, datasetID uuid.UUID) error {
	return r.dao.DeleteSegmentsByDatasetID(ctx, datasetID)
}

func (r *IndexingRepo) DeleteKeywordTablesByDatasetID(ctx context.Context, datasetID uuid.UUID) error {
	return r.dao.DeleteKeywordTablesByDatasetID(ctx, datasetID)
}

func (r *IndexingRepo) DeleteDatasetQueriesByDatasetID(ctx context.Context, datasetID uuid.UUID) error {
	return r.dao.DeleteDatasetQueriesByDatasetID(ctx, datasetID)
}

func (r *IndexingRepo) GetUploadFileByID(ctx context.Context, uploadFileID uuid.UUID) (*entity.UploadFile, error) {
	return r.dao.GetUploadFileByID(ctx, uploadFileID)
}

func (r *IndexingRepo) GetProcessRuleByID(ctx context.Context, processRuleID uuid.UUID) (*entity.ProcessRule, error) {
	return r.dao.GetProcessRuleByID(ctx, processRuleID)
}

func (r *IndexingRepo) GetMaxSegmentPosition(ctx context.Context, documentID uuid.UUID) (int, error) {
	return r.dao.GetMaxSegmentPosition(ctx, documentID)
}

func (r *IndexingRepo) CreateSegment(ctx context.Context, segment *entity.Segment) error {
	return r.dao.CreateSegment(ctx, segment)
}

func (r *IndexingRepo) UpdateSegment(ctx context.Context, segmentID uuid.UUID, updates map[string]any) error {
	return r.dao.UpdateSegment(ctx, segmentID, updates)
}

func (r *IndexingRepo) UpdateKeywordTable(ctx context.Context, keywordTableID uuid.UUID, updates map[string]any) error {
	return r.dao.UpdateKeywordTable(ctx, keywordTableID, updates)
}
