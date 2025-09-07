package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/schema"

	"github.com/crazyfrankie/voidx/internal/core/embedding"
	"github.com/crazyfrankie/voidx/internal/core/retrievers"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/internal/retriever"
	"github.com/crazyfrankie/voidx/internal/segment/repository"
	"github.com/crazyfrankie/voidx/internal/vecstore"
	"github.com/crazyfrankie/voidx/pkg/util"
	"github.com/crazyfrankie/voidx/types/consts"
	"github.com/crazyfrankie/voidx/types/errno"
)

type SegmentService struct {
	repo         *repository.SegmentRepo
	embeddingSvc *embedding.EmbeddingService
	jiebaSvc     *retrievers.JiebaService
	vecSvc       *vecstore.VecStoreService
	keywordSvc   *retriever.KeyWordService
}

func NewSegmentService(repo *repository.SegmentRepo, embeddingSvc *embedding.EmbeddingService,
	jiebaSvc *retrievers.JiebaService, vecSvc *vecstore.VecStoreService, keywordSvc *retriever.KeyWordService) *SegmentService {
	return &SegmentService{repo: repo, embeddingSvc: embeddingSvc, jiebaSvc: jiebaSvc, vecSvc: vecSvc, keywordSvc: keywordSvc}
}

func (s *SegmentService) CreateSegment(ctx context.Context, datasetID, documentID uuid.UUID, createReq req.CreateSegmentReq) (*entity.Segment, error) {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	// 1.校验上传内容的token长度总数，不能超过1000
	tokenCount := s.embeddingSvc.CalculateTokenCount(createReq.Content)
	if tokenCount > 1000 {
		return nil, errno.ErrValidate.AppendBizMessage(errors.New("片段内容的长度不能超过1000 token"))
	}

	// 2.验证数据集和文档权限
	if err := s.repo.ValidateDatasetAccess(ctx, datasetID, userID); err != nil {
		return nil, err
	}
	if err := s.repo.ValidateDocumentAccess(ctx, documentID, datasetID); err != nil {
		return nil, err
	}

	doc, err := s.repo.GetDocument(ctx, documentID)
	if err != nil {
		return nil, err
	}
	// 3.判断文档的状态是否可以新增片段数据，只有completed才可以新增
	if doc.Status != "completed" {
		return nil, errors.New("当前文档不可新增片段，请稍后尝试")
	}

	// 4.提取文档片段的最大位置
	position, err := s.repo.GetMaxSegmentPosition(ctx, documentID)
	if err != nil {
		return nil, err
	}

	// 5.检测是否传递了keywords，如果没有传递的话，调用jieba服务生成关键词
	if createReq.Keywords == nil {
		createReq.Keywords = s.jiebaSvc.ExtractKeywords(createReq.Content, 10)
	}

	// 6. 创建片段
	position += 1
	segment := &entity.Segment{
		ID:             uuid.New(),
		AccountID:      userID,
		DatasetID:      datasetID,
		DocumentID:     documentID,
		NodeID:         uuid.New(),
		Position:       position,
		Content:        createReq.Content,
		CharacterCount: len(createReq.Content),
		TokenCount:     tokenCount,
		Keywords:       createReq.Keywords,
		Hash:           util.GenerateHash(createReq.Content),
		Enabled:        true,
		Status:         consts.SegmentStatusCompleted,
		CompletedAt:    time.Now().Unix(),
	}
	err = s.repo.CreateSegment(ctx, segment)

	// 7.往向量数据库中新增数据
	_, err = s.vecSvc.AddDocument(ctx, []schema.Document{
		{
			PageContent: createReq.Content,
			Metadata: map[string]any{
				"account_id":       userID,
				"dataset_id":       datasetID,
				"document_id":      documentID,
				"segment_id":       segment.ID,
				"node_id":          segment.NodeID,
				"document_enabled": doc.Enabled,
				"segment_enabled":  true,
			},
		},
	})

	// 8.重新计算片段的字符总数以及token总数
	docCharCnt, docTokenCnt, err := s.repo.GetDocumentSegmentCounts(ctx, documentID)
	if err != nil {
		return nil, err
	}

	// 9.更新文档的对应信息
	if err := s.repo.UpdateDocument(ctx, documentID, map[string]any{
		"character_count": docCharCnt,
		"token_count":     docTokenCnt,
	}); err != nil {
		return nil, err
	}

	// 10.更新关键词表信息
	if err := s.keywordSvc.AddKeywords(ctx, datasetID, []uuid.UUID{segment.ID}); err != nil {
		return nil, err
	}

	return segment, nil
}

func (s *SegmentService) GetSegmentsWithPage(ctx context.Context, datasetID, documentID uuid.UUID, pageReq req.GetSegmentsWithPageReq) ([]resp.SegmentResp, resp.Paginator, error) {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return nil, resp.Paginator{}, err
	}

	// 验证数据集和文档权限
	if err := s.repo.ValidateDatasetAccess(ctx, datasetID, userID); err != nil {
		return nil, resp.Paginator{}, err
	}

	if err := s.repo.ValidateDocumentAccess(ctx, documentID, datasetID); err != nil {
		return nil, resp.Paginator{}, err
	}

	// 获取片段列表
	segments, total, err := s.repo.GetSegmentsByDocumentID(ctx, documentID, pageReq)
	if err != nil {
		return nil, resp.Paginator{}, err
	}

	// 转换为响应格式
	segmentResps := make([]resp.SegmentResp, len(segments))
	for i, segment := range segments {
		segmentResps[i] = resp.SegmentResp{
			ID:         segment.ID,
			DatasetID:  segment.DatasetID,
			DocumentID: segment.DocumentID,
			Content:    segment.Content,
			Keywords:   segment.Keywords,
			Enabled:    segment.Enabled,
			Status:     string(segment.Status),
			Utime:      segment.Utime,
		}
	}

	// 计算分页信息
	totalPages := (int(total) + pageReq.PageSize - 1) / pageReq.PageSize
	paginator := resp.Paginator{
		CurrentPage: pageReq.CurrentPage,
		PageSize:    pageReq.PageSize,
		TotalPage:   totalPages,
		TotalRecord: int(total),
	}

	return segmentResps, paginator, nil
}

func (s *SegmentService) GetSegment(ctx context.Context, datasetID, documentID, segmentID uuid.UUID) (*resp.SegmentResp, error) {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	// 验证数据集和文档权限
	if err := s.repo.ValidateDatasetAccess(ctx, datasetID, userID); err != nil {
		return nil, err
	}

	if err := s.repo.ValidateDocumentAccess(ctx, documentID, datasetID); err != nil {
		return nil, err
	}

	// 获取片段
	segment, err := s.repo.GetSegmentByID(ctx, segmentID)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("片段不存在"))
	}

	// 验证片段属于指定文档
	if segment.DocumentID != documentID {
		return nil, errno.ErrValidate.AppendBizMessage(errors.New("片段不属于指定文档"))
	}

	segmentResp := &resp.SegmentResp{
		ID:         segment.ID,
		DatasetID:  segment.DatasetID,
		DocumentID: segment.DocumentID,
		Content:    segment.Content,
		Keywords:   segment.Keywords,
		Enabled:    segment.Enabled,
		Status:     string(segment.Status),
		Utime:      segment.Utime,
	}

	return segmentResp, nil
}

func (s *SegmentService) UpdateSegment(ctx context.Context, datasetID, documentID, segmentID uuid.UUID, updateReq req.UpdateSegmentReq) error {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return err
	}

	// 验证数据集和文档权限
	if err := s.repo.ValidateDatasetAccess(ctx, datasetID, userID); err != nil {
		return err
	}

	if err := s.repo.ValidateDocumentAccess(ctx, documentID, datasetID); err != nil {
		return err
	}

	// 验证片段存在
	segment, err := s.repo.GetSegmentByID(ctx, segmentID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("片段不存在"))
	}

	// 验证片段属于指定文档
	if segment.DocumentID != documentID {
		return errno.ErrValidate.AppendBizMessage(errors.New("片段不属于指定文档"))
	}

	// 构建更新字段
	updates := make(map[string]any)
	if updateReq.Content != "" {
		updates["content"] = updateReq.Content
	}
	if updateReq.Keywords != nil {
		updates["keywords"] = updateReq.Keywords
	}
	if updateReq.EnabledStatus != nil {
		updates["enabled_status"] = *updateReq.EnabledStatus
	}

	return s.repo.UpdateSegment(ctx, segmentID, updates)
}

func (s *SegmentService) DeleteSegment(ctx context.Context, datasetID, documentID, segmentID uuid.UUID) error {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return err
	}

	// 验证数据集和文档权限
	if err := s.repo.ValidateDatasetAccess(ctx, datasetID, userID); err != nil {
		return err
	}

	if err := s.repo.ValidateDocumentAccess(ctx, documentID, datasetID); err != nil {
		return err
	}

	// 验证片段存在
	segment, err := s.repo.GetSegmentByID(ctx, segmentID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("片段不存在"))
	}

	// 验证片段属于指定文档
	if segment.DocumentID != documentID {
		return errno.ErrValidate.AppendBizMessage(errors.New("片段不属于指定文档"))
	}

	return s.repo.DeleteSegment(ctx, segmentID)
}

func (s *SegmentService) UpdateSegmentEnabled(ctx context.Context, datasetID, documentID, segmentID uuid.UUID, enabled bool) error {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return err
	}

	// 验证数据集和文档权限
	if err := s.repo.ValidateDatasetAccess(ctx, datasetID, userID); err != nil {
		return err
	}

	if err := s.repo.ValidateDocumentAccess(ctx, documentID, datasetID); err != nil {
		return err
	}

	// 验证片段存在
	segment, err := s.repo.GetSegmentByID(ctx, segmentID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("片段不存在"))
	}

	// 验证片段属于指定文档
	if segment.DocumentID != documentID {
		return errno.ErrValidate.AppendBizMessage(errors.New("片段不属于指定文档"))
	}

	updates := map[string]any{
		"enabled_status": enabled,
	}

	return s.repo.UpdateSegment(ctx, segmentID, updates)
}

func (s *SegmentService) GetSegments(ctx context.Context, segmentIDS []uuid.UUID) ([]entity.Segment, error) {
	return s.repo.GetSegments(ctx, segmentIDS)
}
