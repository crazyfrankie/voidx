package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/document/repository"
	"github.com/crazyfrankie/voidx/internal/document/task"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/util"
)

type DocumentService struct {
	repo         *repository.DocumentRepo
	taskProducer *task.DocumentProducer
}

func NewDocumentService(repo *repository.DocumentRepo, taskProducer *task.DocumentProducer) *DocumentService {
	return &DocumentService{
		repo:         repo,
		taskProducer: taskProducer,
	}
}

// CreateDocuments 根据传递的信息创建文档列表并调用异步任务
func (s *DocumentService) CreateDocuments(ctx context.Context, datasetID uuid.UUID, createReq req.CreateDocumentsReq) ([]entity.Document, string, error) {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return nil, "", err
	}

	// 1. 检测知识库权限
	dataset, err := s.repo.GetDatasetByID(ctx, datasetID)
	if err != nil {
		return nil, "", errno.ErrNotFound.AppendBizMessage(errors.New("知识库不存在"))
	}

	if dataset.AccountID != userID {
		return nil, "", errno.ErrForbidden.AppendBizMessage(errors.New("当前用户无该知识库权限或知识库不存在"))
	}

	// 2. 提取文件并校验文件权限与文件扩展
	uploadFiles, err := s.repo.GetUploadFilesByIDs(ctx, userID, createReq.UploadFileIDs)
	if err != nil {
		return nil, "", err
	}

	// 过滤合法的文档文件
	validUploadFiles := make([]entity.UploadFile, 0)
	allowedExtensions := map[string]bool{
		".txt": true, ".md": true, ".pdf": true, ".doc": true, ".docx": true,
		".html": true, ".htm": true, ".rtf": true, ".csv": true, ".json": true,
	}

	for _, file := range uploadFiles {
		if allowedExtensions[file.Extension] {
			validUploadFiles = append(validUploadFiles, file)
		}
	}

	if len(validUploadFiles) == 0 {
		return nil, "", errno.ErrValidate.AppendBizMessage(errors.New("暂未解析到合法文件，请重新上传"))
	}

	// 3. 创建批次与处理规则并记录到数据库中
	batch := fmt.Sprintf("%d%06d", time.Now().Unix(), time.Now().Nanosecond()%1000000)

	// 创建处理规则
	processRule := &entity.ProcessRule{
		AccountID: userID,
		DatasetID: datasetID,
		Mode:      createReq.ProcessType,
		Rule:      createReq.Rule,
	}

	err = s.repo.CreateProcessRule(ctx, processRule)
	if err != nil {
		return nil, "", err
	}

	// 4. 获取当前知识库的最新文档位置
	position, err := s.repo.GetLatestDocumentPosition(ctx, datasetID)
	if err != nil {
		return nil, "", err
	}

	// 5. 循环遍历所有合法的上传文件列表并记录
	documents := make([]entity.Document, 0, len(validUploadFiles))
	for _, uploadFile := range validUploadFiles {
		position++
		document := entity.Document{
			AccountID:     userID,
			DatasetID:     datasetID,
			UploadFileID:  uploadFile.ID,
			ProcessRuleID: processRule.ID,
			Batch:         batch,
			Name:          uploadFile.Name,
			Position:      position,
			Enabled:       true,
		}

		err = s.repo.CreateDocument(ctx, &document)
		if err != nil {
			return nil, "", err
		}
		documents = append(documents, document)
	}

	// 6. 调用异步任务，完成后续操作
	var documentIDs []uuid.UUID
	for _, doc := range documents {
		documentIDs = append(documentIDs, doc.ID)
	}

	// 为每个文档发布构建任务
	for _, documentID := range documentIDs {
		err = s.taskProducer.PublishBuildDocumentTask(ctx, documentID)
		if err != nil {
			// 记录错误但不中断整个流程
			fmt.Printf("Failed to publish build task for document %s: %v\n", documentID, err)
		}
	}

	return documents, batch, nil
}

// GetDocumentsStatus 根据传递的知识库id+处理批次获取文档列表的状态
func (s *DocumentService) GetDocumentsStatus(ctx context.Context, datasetID uuid.UUID, batch string) ([]resp.DocumentStatusResp, error) {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	// 1. 检测知识库权限
	dataset, err := s.repo.GetDatasetByID(ctx, datasetID)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("知识库不存在"))
	}

	if dataset.AccountID != userID {
		return nil, errno.ErrForbidden.AppendBizMessage(errors.New("当前用户无该知识库权限或知识库不存在"))
	}

	// 2. 查询当前知识库下该批次的文档列表
	documents, err := s.repo.GetDocumentsByBatch(ctx, datasetID, batch)
	if err != nil {
		return nil, err
	}

	if len(documents) == 0 {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("该处理批次未发现文档，请核实后重试"))
	}

	// 3. 循环遍历文档列表提取文档的状态信息
	documentsStatus := make([]resp.DocumentStatusResp, 0, len(documents))
	for _, document := range documents {
		// 4. 查询每个文档的总片段数和已构建完成的片段数
		segmentCount, err := s.repo.GetSegmentCountByDocument(ctx, document.ID)
		if err != nil {
			segmentCount = 0
		}

		completedSegmentCount, err := s.repo.GetCompletedSegmentCountByDocument(ctx, document.ID)
		if err != nil {
			completedSegmentCount = 0
		}

		// 获取上传文件信息
		var uploadFile *entity.UploadFile
		if document.UploadFileID != uuid.Nil {
			uploadFile, _ = s.repo.GetUploadFileByID(ctx, document.UploadFileID)
		}

		status := resp.DocumentStatusResp{
			ID:                    document.ID,
			Name:                  document.Name,
			Position:              document.Position,
			SegmentCount:          segmentCount,
			CompletedSegmentCount: completedSegmentCount,
			Status:                document.Status,
			ProcessingStartedAt:   document.ProcessingStartedAt,
			ParsingCompletedAt:    document.ParsingCompletedAt,
			SplittingCompletedAt:  document.SplittingCompletedAt,
			IndexingCompletedAt:   document.IndexingCompletedAt,
			CompletedAt:           document.CompletedAt,
			StoppedAt:             document.StoppedAt,
			Ctime:                 document.Ctime,
		}

		if uploadFile != nil {
			status.Size = uploadFile.Size
			status.Extension = uploadFile.Extension
			status.MimeType = uploadFile.MimeType
		}

		documentsStatus = append(documentsStatus, status)
	}

	return documentsStatus, nil
}

func (s *DocumentService) GetDocumentsWithPage(ctx context.Context, datasetID uuid.UUID, pageReq req.GetDocumentsWithPageReq) ([]resp.DocumentResp, resp.Paginator, error) {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return nil, resp.Paginator{}, err
	}

	// 验证知识库权限
	dataset, err := s.repo.GetDatasetByID(ctx, datasetID)
	if err != nil {
		return nil, resp.Paginator{}, errno.ErrNotFound.AppendBizMessage(errors.New("知识库不存在"))
	}

	if dataset.AccountID != userID {
		return nil, resp.Paginator{}, errno.ErrForbidden.AppendBizMessage(errors.New("无权限访问该知识库"))
	}

	// 获取文档列表
	documents, total, err := s.repo.GetDocumentsByDatasetID(ctx, datasetID, pageReq)
	if err != nil {
		return nil, resp.Paginator{}, err
	}

	// 转换为响应格式
	documentResps := make([]resp.DocumentResp, len(documents))
	for i, doc := range documents {
		res, err := s.buildDocumentResp(ctx, &doc)
		if err != nil {
			continue
		}
		documentResps[i] = *res
	}

	// 计算分页信息
	totalPages := (int(total) + pageReq.PageSize - 1) / pageReq.PageSize
	paginator := resp.Paginator{
		CurrentPage: pageReq.CurrentPage,
		PageSize:    pageReq.PageSize,
		TotalPage:   totalPages,
		TotalRecord: int(total),
	}

	return documentResps, paginator, nil
}

func (s *DocumentService) GetDocument(ctx context.Context, datasetID, documentID uuid.UUID) (*resp.DocumentResp, error) {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	// 验证知识库权限
	dataset, err := s.repo.GetDatasetByID(ctx, datasetID)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("知识库不存在"))
	}

	if dataset.AccountID != userID {
		return nil, errno.ErrForbidden.AppendBizMessage(errors.New("无权限访问该知识库"))
	}

	// 获取文档
	document, err := s.repo.GetDocumentByID(ctx, documentID)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("文档不存在"))
	}

	if document.DatasetID != datasetID {
		return nil, errno.ErrValidate.AppendBizMessage(errors.New("文档不属于该知识库"))
	}

	return s.buildDocumentResp(ctx, document)
}

func (s *DocumentService) UpdateDocument(ctx context.Context, datasetID, documentID uuid.UUID, updateReq req.UpdateDocumentReq) error {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return err
	}

	// 验证知识库权限
	dataset, err := s.repo.GetDatasetByID(ctx, datasetID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("知识库不存在"))
	}

	if dataset.AccountID != userID {
		return errno.ErrForbidden.AppendBizMessage(errors.New("无权限修改该知识库"))
	}

	// 验证文档
	document, err := s.repo.GetDocumentByID(ctx, documentID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("文档不存在"))
	}

	if document.DatasetID != datasetID {
		return errno.ErrValidate.AppendBizMessage(errors.New("文档不属于该知识库"))
	}

	// 构建更新数据
	updates := make(map[string]any)
	if updateReq.Name != "" {
		updates["name"] = updateReq.Name
	}

	return s.repo.UpdateDocument(ctx, documentID, updates)
}

func (s *DocumentService) DeleteDocument(ctx context.Context, datasetID, documentID uuid.UUID) error {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return err
	}

	// 验证知识库权限
	dataset, err := s.repo.GetDatasetByID(ctx, datasetID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("知识库不存在"))
	}

	if dataset.AccountID != userID {
		return errno.ErrForbidden.AppendBizMessage(errors.New("无权限删除该知识库中的文档"))
	}

	// 验证文档
	document, err := s.repo.GetDocumentByID(ctx, documentID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("文档不存在"))
	}

	if document.DatasetID != datasetID {
		return errno.ErrValidate.AppendBizMessage(errors.New("文档不属于该知识库"))
	}

	// 检查文档状态，只有完成或错误状态才能删除
	if document.Status != "completed" && document.Status != "error" {
		return errno.ErrValidate.AppendBizMessage(errors.New("文档正在处理中，无法删除"))
	}

	return s.repo.DeleteDocument(ctx, documentID)
}

func (s *DocumentService) UpdateDocumentEnabled(ctx context.Context, datasetID, documentID uuid.UUID, enabled bool) error {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return err
	}

	// 验证知识库权限
	dataset, err := s.repo.GetDatasetByID(ctx, datasetID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("知识库不存在"))
	}

	if dataset.AccountID != userID {
		return errno.ErrForbidden.AppendBizMessage(errors.New("无权限修改该知识库"))
	}

	// 验证文档
	document, err := s.repo.GetDocumentByID(ctx, documentID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("文档不存在"))
	}

	if document.DatasetID != datasetID {
		return errno.ErrValidate.AppendBizMessage(errors.New("文档不属于该知识库"))
	}

	// 更新启用状态
	updates := map[string]any{
		"enabled_status": enabled,
	}

	if !enabled {
		now := time.Now().UnixMilli()
		updates["disabled_at"] = &now
		updates["disabled_by"] = &userID
	} else {
		updates["disabled_at"] = nil
		updates["disabled_by"] = nil
	}

	return s.repo.UpdateDocument(ctx, documentID, updates)
}

func (s *DocumentService) buildDocumentResp(ctx context.Context, doc *entity.Document) (*resp.DocumentResp, error) {
	type statsResult struct {
		segmentCount int
		hitCount     int
	}

	var res statsResult

	errCh := make(chan error, 1)
	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		if count, err := s.repo.GetSegmentCountByDocument(ctx, doc.ID); err != nil {
			errCh <- err
		} else {
			res.segmentCount = count
		}
		wg.Done()
	}()
	go func() {
		wg.Add(1)
		// 获取命中次数
		if count, err := s.repo.GetHitCountByDocument(ctx, doc.ID); err != nil {
			errCh <- err
		} else {
			res.hitCount = count
		}
		wg.Done()
	}()

	err := <-errCh
	if err != nil {
		return nil, err
	}
	wg.Wait()

	return &resp.DocumentResp{
		ID:             doc.ID,
		DatasetID:      doc.DatasetID,
		Position:       doc.Position,
		Name:           doc.Name,
		Status:         doc.Status,
		SegmentCount:   res.segmentCount,
		CharacterCount: doc.CharacterCount,
		HitCount:       res.hitCount,
		Enabled:        doc.Enabled,
		DisabledAt:     doc.DisabledAt,
		Ctime:          doc.Ctime,
		Utime:          doc.Utime,
	}, nil
}

func (s *DocumentService) RawGetDocument(ctx context.Context, datasetID, documentID uuid.UUID) (*entity.Document, error) {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	// 验证知识库权限
	dataset, err := s.repo.GetDatasetByID(ctx, datasetID)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("知识库不存在"))
	}

	if dataset.AccountID != userID {
		return nil, errno.ErrForbidden.AppendBizMessage(errors.New("无权限访问该知识库"))
	}

	// 获取文档
	document, err := s.repo.GetDocumentByID(ctx, documentID)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("文档不存在"))
	}

	if document.DatasetID != datasetID {
		return nil, errno.ErrValidate.AppendBizMessage(errors.New("文档不属于该知识库"))
	}

	return document, nil
}
