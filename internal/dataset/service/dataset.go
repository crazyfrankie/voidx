package service

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/dataset/repository"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/internal/retriever"
	"github.com/crazyfrankie/voidx/internal/segment"
	"github.com/crazyfrankie/voidx/pkg/util"
	"github.com/crazyfrankie/voidx/types/consts"
	"github.com/crazyfrankie/voidx/types/errno"
)

type DatasetService struct {
	repo             *repository.DatasetRepo
	retrieverService *retriever.Service
	segmentService   *segment.Service
	producer         *DatasetProducer
}

func NewDatasetService(repo *repository.DatasetRepo, retrieverSvc *retriever.Service, segmentService *segment.Service, producer *DatasetProducer) *DatasetService {
	return &DatasetService{
		repo:             repo,
		retrieverService: retrieverSvc,
		segmentService:   segmentService,
		producer:         producer,
	}
}

func (s *DatasetService) CreateDataset(ctx context.Context, createReq req.CreateDatasetReq) error {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return err
	}

	dataset := &entity.Dataset{
		AccountID:   userID,
		Name:        createReq.Name,
		Description: createReq.Description,
		Icon:        createReq.Icon,
	}

	return s.repo.CreateDataset(ctx, dataset)
}

func (s *DatasetService) GetDatasetsWithPage(ctx context.Context, pageReq req.GetDatasetsWithPageReq) ([]resp.DatasetResp, resp.Paginator, error) {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return nil, resp.Paginator{}, err
	}

	datasets, total, err := s.repo.GetDatasetsByAccountID(ctx, userID, pageReq)
	if err != nil {
		return nil, resp.Paginator{}, err
	}

	// 转换为响应格式并填充计算字段
	datasetResps := make([]resp.DatasetResp, len(datasets))
	for i, dataset := range datasets {
		datasetResp, err := s.buildDatasetResponse(ctx, &dataset)
		if err != nil {
			// 如果计算字段失败，使用默认值
			datasetResp = &resp.DatasetResp{
				ID:          dataset.ID,
				Name:        dataset.Name,
				Description: dataset.Description,
				Ctime:       dataset.Ctime,
				Utime:       dataset.Utime,
			}
		}
		datasetResps[i] = *datasetResp
	}

	// 计算分页信息
	totalPages := (int(total) + pageReq.PageSize - 1) / pageReq.PageSize
	paginator := resp.Paginator{
		CurrentPage: pageReq.CurrentPage,
		PageSize:    pageReq.PageSize,
		TotalPage:   totalPages,
		TotalRecord: int(total),
	}

	return datasetResps, paginator, nil
}

func (s *DatasetService) GetDataset(ctx context.Context, datasetID uuid.UUID) (*resp.DatasetResp, error) {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	dataset, err := s.repo.GetDatasetByID(ctx, datasetID)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("知识库不存在"))
	}

	// 验证权限
	if dataset.AccountID != userID {
		return nil, errno.ErrForbidden.AppendBizMessage(errors.New("无权限访问该知识库"))
	}

	return s.buildDatasetResponse(ctx, dataset)
}

func (s *DatasetService) UpdateDataset(ctx context.Context, datasetID uuid.UUID, updateReq req.UpdateDatasetReq) error {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return err
	}

	dataset, err := s.repo.GetDatasetByID(ctx, datasetID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("知识库不存在"))
	}

	// 验证权限
	if dataset.AccountID != userID {
		return errno.ErrForbidden.AppendBizMessage(errors.New("无权限修改该知识库"))
	}

	updates := make(map[string]any)
	if updateReq.Name != "" {
		updates["name"] = updateReq.Name
	}
	if updateReq.Description != "" {
		updates["description"] = updateReq.Description
	}
	if updateReq.Permission != "" {
		updates["permission"] = updateReq.Permission
	}

	return s.repo.UpdateDataset(ctx, datasetID, updates)
}

func (s *DatasetService) DeleteDataset(ctx context.Context, datasetID uuid.UUID) error {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return err
	}

	dataset, err := s.repo.GetDatasetByID(ctx, datasetID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("知识库不存在"))
	}

	// 验证权限
	if dataset.AccountID != userID {
		return errno.ErrForbidden.AppendBizMessage(errors.New("无权限删除该知识库"))
	}

	// 发布异步删除任务
	if err := s.producer.PublishDeleteDatasetTask(ctx, datasetID); err != nil {
		return fmt.Errorf("failed to publish delete dataset task: %w", err)
	}

	return s.repo.DeleteDataset(ctx, datasetID)
}

func (s *DatasetService) Hit(ctx context.Context, datasetID uuid.UUID, hitReq req.HitReq) ([]resp.HitResult, error) {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	dataset, err := s.repo.GetDatasetByID(ctx, datasetID)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("知识库不存在"))
	}

	// 验证权限
	if dataset.AccountID != userID {
		return nil, errno.ErrForbidden.AppendBizMessage(errors.New("无权限访问该知识库"))
	}

	// 调用检索服务进行检索
	searchReq := req.SearchRequest{
		DatasetIDs:     []uuid.UUID{datasetID},
		Query:          hitReq.Query,
		K:              hitReq.K,
		ScoreThreshold: hitReq.Score,
	}
	if hitReq.RetrievalStrategy != "" {
		searchReq.RetrieverType = hitReq.RetrievalStrategy
	} else {
		searchReq.RetrieverType = string(consts.RetrievalStrategySemantic)
	}

	searchResults, err := s.retrieverService.SearchInDatasets(ctx, userID, searchReq)
	if err != nil {
		return nil, err
	}

	segmentIDs := make([]uuid.UUID, 0, len(searchResults))
	for _, res := range searchResults {
		segmentIDs = append(segmentIDs, res.SegmentID)
	}

	segments, err := s.segmentService.GetSegments(ctx, segmentIDs)
	if err != nil {
		return nil, err
	}

	sortSearchs := s.filterSearchResult(searchResults, segments)

	// 转换为响应格式
	hitRes := make([]resp.HitResult, len(segments))
	for i, res := range segments {
		hit := resp.HitResult{
			Document: resp.DocumentResp{
				ID:   res.DocumentID,
				Name: sortSearchs[i].DocumentName,
			},
			SegmentID:      res.ID,
			DocumentID:     res.DocumentID,
			DatasetID:      res.DatasetID,
			Content:        res.Content,
			Score:          searchResults[i].Score,
			Position:       res.Position,
			Keywords:       res.Keywords,
			CharacterCount: res.CharacterCount,
			HitCount:       res.HitCount,
			TokenCount:     res.TokenCount,
			DisabledAt:     res.DisabledAt,
			Enabled:        res.Enabled,
			Status:         string(res.Status),
			Ctime:          res.Ctime,
			Utime:          res.Utime,
		}
		hitRes = append(hitRes, hit)
	}

	return hitRes, nil
}

func (s *DatasetService) GetDatasetQueries(ctx context.Context, datasetID uuid.UUID) ([]resp.DatasetQueryResp, error) {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	dataset, err := s.repo.GetDatasetByID(ctx, datasetID)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("知识库不存在"))
	}

	// 验证权限
	if dataset.AccountID != userID {
		return nil, errno.ErrForbidden.AppendBizMessage(errors.New("无权限访问该知识库"))
	}

	queries, err := s.repo.GetDatasetQueries(ctx, datasetID, 10) // 获取最近10条查询记录
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	queryResps := make([]resp.DatasetQueryResp, len(queries))
	for i, query := range queries {
		queryResps[i] = resp.DatasetQueryResp{
			ID:        query.ID,
			DatasetID: query.DatasetID,
			Content:   query.Query,
			Source:    query.Source,
			Ctime:     query.Ctime,
		}
	}

	return queryResps, nil
}

func (s *DatasetService) filterSearchResult(search []resp.SearchResult, segments []entity.Segment) []resp.SearchResult {
	segmentMap := make(map[uuid.UUID]bool)
	for _, seg := range segments {
		segmentMap[seg.ID] = true
	}
	res := make([]resp.SearchResult, 0, len(segmentMap))
	for _, r := range search {
		if ok := segmentMap[r.SegmentID]; ok {
			res = append(res, r)
		}
	}

	return res
}

// buildDatasetResponse 构建包含计算字段的数据集响应
func (s *DatasetService) buildDatasetResponse(ctx context.Context, dataset *entity.Dataset) (*resp.DatasetResp, error) {
	// 并发获取所有计算字段，提高性能
	type statsResult struct {
		documentCount   int
		hitCount        int
		relatedAppCount int
		characterCount  int
	}

	var res statsResult

	errCh := make(chan error, 1)
	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		// 获取文档数量
		if count, err := s.repo.GetDocumentCount(ctx, dataset.ID); err != nil {
			errCh <- err
		} else {
			res.documentCount = count
		}
		wg.Done()
	}()

	go func() {
		wg.Add(1)
		// 获取命中次数
		if count, err := s.repo.GetHitCount(ctx, dataset.ID); err != nil {
			errCh <- err
		} else {
			res.hitCount = count
		}
		wg.Done()
	}()
	go func() {
		// 获取关联应用数量
		wg.Add(1)
		if count, err := s.repo.GetRelatedAppCount(ctx, dataset.ID); err != nil {
			errCh <- err
		} else {
			res.relatedAppCount = count
		}
		wg.Done()
	}()
	go func() {
		// 获取字符总数
		wg.Add(1)
		if count, err := s.repo.GetCharacterCount(ctx, dataset.ID); err != nil {
			errCh <- err
		} else {
			res.characterCount = count
		}
		wg.Done()
	}()

	// 等待结果
	err := <-errCh
	if err != nil {
		return nil, err
	}
	wg.Wait()

	return &resp.DatasetResp{
		ID:              dataset.ID,
		Name:            dataset.Name,
		Description:     dataset.Description,
		DocumentCount:   res.documentCount,
		HitCount:        res.hitCount,
		RelatedAppCount: res.relatedAppCount,
		CharacterCount:  res.characterCount,
		Ctime:           dataset.Ctime,
		Utime:           dataset.Utime,
	}, nil
}
