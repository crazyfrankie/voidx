package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/document/repository/dao"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
)

type DocumentRepo struct {
	dao *dao.DocumentDao
}

func NewDocumentRepo(d *dao.DocumentDao) *DocumentRepo {
	return &DocumentRepo{dao: d}
}

func (r *DocumentRepo) CreateDocument(ctx context.Context, document *entity.Document) error {
	return r.dao.CreateDocument(ctx, document)
}

func (r *DocumentRepo) GetDocumentByID(ctx context.Context, id uuid.UUID) (*entity.Document, error) {
	return r.dao.GetDocumentByID(ctx, id)
}

func (r *DocumentRepo) GetDocumentsByDatasetID(
	ctx context.Context,
	datasetID uuid.UUID,
	pageReq req.GetDocumentsWithPageReq,
) ([]entity.Document, int64, error) {
	return r.dao.GetDocumentsByDatasetID(ctx, datasetID, pageReq)
}

func (r *DocumentRepo) UpdateDocument(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return r.dao.UpdateDocument(ctx, id, updates)
}

func (r *DocumentRepo) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	return r.dao.DeleteDocument(ctx, id)
}

func (r *DocumentRepo) GetNextDocumentPosition(ctx context.Context, datasetID uuid.UUID) (int, error) {
	return r.dao.GetNextDocumentPosition(ctx, datasetID)
}

func (r *DocumentRepo) GetDatasetByID(ctx context.Context, id uuid.UUID) (*entity.Dataset, error) {
	return r.dao.GetDatasetByID(ctx, id)
}

// GetUploadFilesByIDs 根据ID列表获取上传文件
func (r *DocumentRepo) GetUploadFilesByIDs(ctx context.Context, userID uuid.UUID, fileIDs []uuid.UUID) ([]entity.UploadFile, error) {
	return r.dao.GetUploadFilesByIDs(ctx, userID, fileIDs)
}

// CreateProcessRule 创建处理规则
func (r *DocumentRepo) CreateProcessRule(ctx context.Context, rule *entity.ProcessRule) error {
	return r.dao.CreateProcessRule(ctx, rule)
}

// GetLatestDocumentPosition 获取最新文档位置
func (r *DocumentRepo) GetLatestDocumentPosition(ctx context.Context, datasetID uuid.UUID) (int, error) {
	return r.dao.GetLatestDocumentPosition(ctx, datasetID)
}

// GetDocumentsByBatch 根据批次获取文档列表
func (r *DocumentRepo) GetDocumentsByBatch(ctx context.Context, datasetID uuid.UUID, batch string) ([]entity.Document, error) {
	return r.dao.GetDocumentsByBatch(ctx, datasetID, batch)
}

// GetSegmentCountByDocument 获取文档的片段总数
func (r *DocumentRepo) GetSegmentCountByDocument(ctx context.Context, documentID uuid.UUID) (int, error) {
	return r.dao.GetSegmentCountByDocument(ctx, documentID)
}

func (r *DocumentRepo) GetHitCountByDocument(ctx context.Context, documentID uuid.UUID) (int, error) {
	return r.dao.GetHitCountByDocument(ctx, documentID)
}

// GetCompletedSegmentCountByDocument 获取文档的已完成片段数
func (r *DocumentRepo) GetCompletedSegmentCountByDocument(ctx context.Context, documentID uuid.UUID) (int, error) {
	return r.dao.GetCompletedSegmentCountByDocument(ctx, documentID)
}

// GetUploadFileByID 根据ID获取上传文件
func (r *DocumentRepo) GetUploadFileByID(ctx context.Context, fileID uuid.UUID) (*entity.UploadFile, error) {
	return r.dao.GetUploadFileByID(ctx, fileID)
}
