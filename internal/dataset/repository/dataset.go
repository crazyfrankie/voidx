package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/dataset/repository/dao"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
)

type DatasetRepo struct {
	dao *dao.DatasetDao
}

func NewDatasetRepo(d *dao.DatasetDao) *DatasetRepo {
	return &DatasetRepo{dao: d}
}

func (r *DatasetRepo) CreateDataset(ctx context.Context, dataset *entity.Dataset) error {
	return r.dao.CreateDataset(ctx, dataset)
}

func (r *DatasetRepo) GetDatasetByID(ctx context.Context, id uuid.UUID) (*entity.Dataset, error) {
	return r.dao.GetDatasetByID(ctx, id)
}

func (r *DatasetRepo) GetDatasetsByAccountID(
	ctx context.Context,
	accountID uuid.UUID,
	pageReq req.GetDatasetsWithPageReq,
) ([]entity.Dataset, int64, error) {
	return r.dao.GetDatasetsByAccountID(ctx, accountID, pageReq)
}

func (r *DatasetRepo) UpdateDataset(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return r.dao.UpdateDataset(ctx, id, updates)
}

func (r *DatasetRepo) DeleteDataset(ctx context.Context, id uuid.UUID) error {
	return r.dao.DeleteDataset(ctx, id)
}

func (r *DatasetRepo) GetDatasetQueries(ctx context.Context, datasetID uuid.UUID, limit int) ([]entity.DatasetQuery, error) {
	return r.dao.GetDatasetQueries(ctx, datasetID, limit)
}

func (r *DatasetRepo) CreateDatasetQuery(ctx context.Context, query *entity.DatasetQuery) error {
	return r.dao.CreateDatasetQuery(ctx, query)
}

// Document related methods
func (r *DatasetRepo) CreateDocument(ctx context.Context, document *entity.Document) error {
	return r.dao.CreateDocument(ctx, document)
}

func (r *DatasetRepo) GetDocumentByID(ctx context.Context, id uuid.UUID) (*entity.Document, error) {
	return r.dao.GetDocumentByID(ctx, id)
}

func (r *DatasetRepo) GetDocumentsByDatasetID(
	ctx context.Context,
	datasetID uuid.UUID,
	pageReq req.GetDocumentsWithPageReq,
) ([]entity.Document, int64, error) {
	return r.dao.GetDocumentsByDatasetID(ctx, datasetID, pageReq)
}

func (r *DatasetRepo) UpdateDocument(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return r.dao.UpdateDocument(ctx, id, updates)
}

func (r *DatasetRepo) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	return r.dao.DeleteDocument(ctx, id)
}

// Segment related methods
func (r *DatasetRepo) CreateSegment(ctx context.Context, segment *entity.Segment) error {
	return r.dao.CreateSegment(ctx, segment)
}

func (r *DatasetRepo) GetSegmentByID(ctx context.Context, id uuid.UUID) (*entity.Segment, error) {
	return r.dao.GetSegmentByID(ctx, id)
}

func (r *DatasetRepo) GetSegmentsByDocumentID(ctx context.Context, documentID uuid.UUID, pageReq req.GetSegmentsWithPageReq) ([]entity.Segment, int64, error) {
	return r.dao.GetSegmentsByDocumentID(ctx, documentID, pageReq)
}

func (r *DatasetRepo) UpdateSegment(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return r.dao.UpdateSegment(ctx, id, updates)
}

func (r *DatasetRepo) DeleteSegment(ctx context.Context, id uuid.UUID) error {
	return r.dao.DeleteSegment(ctx, id)
}

// Statistics related methods

func (r *DatasetRepo) GetDocumentCount(ctx context.Context, datasetID uuid.UUID) (int, error) {
	return r.dao.GetDocumentCount(ctx, datasetID)
}

func (r *DatasetRepo) GetHitCount(ctx context.Context, datasetID uuid.UUID) (int, error) {
	return r.dao.GetHitCount(ctx, datasetID)
}

func (r *DatasetRepo) GetRelatedAppCount(ctx context.Context, datasetID uuid.UUID) (int, error) {
	return r.dao.GetRelatedAppCount(ctx, datasetID)
}

func (r *DatasetRepo) GetCharacterCount(ctx context.Context, datasetID uuid.UUID) (int, error) {
	return r.dao.GetCharacterCount(ctx, datasetID)
}
