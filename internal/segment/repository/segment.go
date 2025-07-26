package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/segment/repository/dao"
)

type SegmentRepo struct {
	dao *dao.SegmentDao
}

func NewSegmentRepo(d *dao.SegmentDao) *SegmentRepo {
	return &SegmentRepo{dao: d}
}

func (r *SegmentRepo) ValidateDatasetAccess(ctx context.Context, datasetID, userID uuid.UUID) error {
	return r.dao.ValidateDatasetAccess(ctx, datasetID, userID)
}

func (r *SegmentRepo) ValidateDocumentAccess(ctx context.Context, documentID, datasetID uuid.UUID) error {
	return r.dao.ValidateDocumentAccess(ctx, documentID, datasetID)
}

func (r *SegmentRepo) CreateSegment(ctx context.Context, segment *entity.Segment) error {
	return r.dao.CreateSegment(ctx, segment)
}

func (r *SegmentRepo) GetSegmentsByDocumentID(ctx context.Context, documentID uuid.UUID, pageReq req.GetSegmentsWithPageReq) ([]entity.Segment, int64, error) {
	return r.dao.GetSegmentsByDocumentID(ctx, documentID, pageReq)
}

func (r *SegmentRepo) GetSegmentByID(ctx context.Context, id uuid.UUID) (*entity.Segment, error) {
	return r.dao.GetSegmentByID(ctx, id)
}

func (r *SegmentRepo) UpdateSegment(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return r.dao.UpdateSegment(ctx, id, updates)
}

func (r *SegmentRepo) DeleteSegment(ctx context.Context, id uuid.UUID) error {
	return r.dao.DeleteSegment(ctx, id)
}

func (r *SegmentRepo) GetSegments(ctx context.Context, segmentIDs []uuid.UUID) ([]entity.Segment, error) {
	return r.dao.GetSegments(ctx, segmentIDs)
}

// document call

func (r *SegmentRepo) GetDocument(ctx context.Context, docID uuid.UUID) (*entity.Document, error) {
	return r.dao.GetDocument(ctx, docID)
}

func (r *SegmentRepo) GetMaxSegmentPosition(ctx context.Context, docID uuid.UUID) (int, error) {
	return r.dao.GetMaxSegmentPosition(ctx, docID)
}

func (r *SegmentRepo) GetDocumentSegmentCounts(ctx context.Context, docID uuid.UUID) (int, int, error) {
	return r.dao.GetDocumentSegmentCounts(ctx, docID)
}

func (r *SegmentRepo) UpdateDocument(ctx context.Context, docID uuid.UUID, updates map[string]any) error {
	return r.dao.UpdateDocument(ctx, docID, updates)
}
