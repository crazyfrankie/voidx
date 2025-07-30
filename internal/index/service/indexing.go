package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	milvusentity "github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/redis/go-redis/v9"
	"github.com/tmc/langchaingo/schema"

	"github.com/crazyfrankie/voidx/internal/core/embedding"
	"github.com/crazyfrankie/voidx/internal/core/file_extractor"
	"github.com/crazyfrankie/voidx/internal/core/retrievers"
	"github.com/crazyfrankie/voidx/internal/index/repository"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/process_rule"
	"github.com/crazyfrankie/voidx/internal/retriever"
	"github.com/crazyfrankie/voidx/internal/vecstore"
	"github.com/crazyfrankie/voidx/pkg/consts"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/util"
)

type IndexingService struct {
	repo                  *repository.IndexingRepo
	redisClient           redis.Cmdable
	fileExtractor         *file_extractor.FileExtractor
	processRuleService    *process_rule.Service
	embeddingsService     *embedding.EmbeddingService
	jiebaService          *retrievers.JiebaService
	keywordTableService   *retriever.KeyWordService
	vectorDatabaseService *vecstore.VecStoreService
}

func NewIndexingService(
	repo *repository.IndexingRepo,
	redisClient redis.Cmdable,
	fileExtractor *file_extractor.FileExtractor,
	processRuleService *process_rule.Service,
	embeddingsService *embedding.EmbeddingService,
	jiebaService *retrievers.JiebaService,
	keywordTableService *retriever.KeyWordService,
	vectorDatabaseService *vecstore.VecStoreService,
) *IndexingService {
	return &IndexingService{
		repo:                  repo,
		redisClient:           redisClient,
		fileExtractor:         fileExtractor,
		processRuleService:    processRuleService,
		embeddingsService:     embeddingsService,
		jiebaService:          jiebaService,
		keywordTableService:   keywordTableService,
		vectorDatabaseService: vectorDatabaseService,
	}
}

// BuildDocuments 根据传递的文档id列表构建知识库文档，涵盖了加载、分割、索引构建、数据存储等内容
func (s *IndexingService) BuildDocuments(ctx context.Context, documentIDs []uuid.UUID) error {
	// 1. 根据传递的文档id获取所有文档
	documents, err := s.repo.GetDocumentsByIDs(ctx, documentIDs)
	if err != nil {
		return err
	}

	// 2. 执行循环遍历所有文档完成对每个文档的构建
	for _, document := range documents {
		if err := s.buildSingleDocument(ctx, document); err != nil {
			log.Printf("构建文档发生错误, 文档ID: %s, 错误信息: %v", document.ID, err)

			// 更新文档状态为错误
			now := time.Now()
			s.repo.UpdateDocument(ctx, document.ID, map[string]any{
				"status":     consts.DocumentStatusError,
				"error":      err.Error(),
				"stopped_at": &now,
			})
		}
	}

	return nil
}

// buildSingleDocument 构建单个文档
func (s *IndexingService) buildSingleDocument(ctx context.Context, document *entity.Document) error {
	// 3. 更新当前状态为解析中，并记录开始处理的时间
	now := time.Now().UnixMilli()
	err := s.repo.UpdateDocument(ctx, document.ID, map[string]any{
		"status":                consts.DocumentStatusParsing,
		"processing_started_at": &now,
	})
	if err != nil {
		return err
	}

	// 4. 执行文档加载步骤，并更新文档的状态与时间
	lcDocuments, err := s.parsing(ctx, document)
	if err != nil {
		return fmt.Errorf("文档解析失败: %w", err)
	}

	// 5. 执行文档分割步骤，并更新文档状态与时间，涵盖了片段的信息
	lcSegments, err := s.splitting(ctx, document, lcDocuments)
	if err != nil {
		return fmt.Errorf("文档分割失败: %w", err)
	}

	// 6. 执行文档索引构建，涵盖关键词提取、向量，并更新数据状态
	err = s.indexing(ctx, document, lcSegments)
	if err != nil {
		return fmt.Errorf("文档索引构建失败: %w", err)
	}

	// 7. 存储操作，涵盖文档状态更新，以及向量数据库的存储
	err = s.completed(ctx, document, lcSegments)
	if err != nil {
		return fmt.Errorf("文档存储失败: %w", err)
	}

	return nil
}

// UpdateDocumentEnabled 根据传递的文档id更新文档状态，同时修改向量数据库中的记录
func (s *IndexingService) UpdateDocumentEnabled(ctx context.Context, documentID uuid.UUID) error {
	// 1. 构建缓存键
	cacheKey := fmt.Sprintf(consts.LockDocumentUpdateEnabled, documentID)

	// 2. 根据传递的document_id获取文档记录
	document, err := s.repo.GetDocumentByID(ctx, documentID)
	if err != nil {
		return err
	}
	if document == nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("当前文档不存在"))
	}

	// 3. 查询归属于当前文档的所有片段的节点id
	segments, err := s.repo.GetSegmentsByDocumentID(ctx, documentID)
	if err != nil {
		return err
	}

	var segmentIDs []uuid.UUID
	var nodeIDs []uuid.UUID
	for _, segment := range segments {
		if segment.Status == consts.SegmentStatusCompleted {
			segmentIDs = append(segmentIDs, segment.ID)
			if segment.NodeID != uuid.Nil {
				nodeIDs = append(nodeIDs, segment.NodeID)
			}
		}
	}

	defer func() {
		// 6. 清空缓存键表示异步操作已经执行完成，无论失败还是成功都全部清除
		s.redisClient.Del(ctx, cacheKey)
	}()

	// 4. 执行循环遍历所有node_ids并更新向量数据
	for _, nodeID := range nodeIDs {
		nodeIDColumn := milvusentity.NewColumnString("uuid", []string{nodeID.String()})
		enabledColumn := milvusentity.NewColumnBool("document_enabled", []bool{document.Enabled})
		err := s.vectorDatabaseService.UpdateDocument(ctx, nodeIDColumn, enabledColumn)
		if err != nil {
			// 更新片段状态为错误
			now := time.Now().UnixMilli()
			s.repo.UpdateSegmentByNodeID(ctx, nodeID, map[string]any{
				"error":       err,
				"status":      consts.SegmentStatusError,
				"enabled":     false,
				"disabled_at": &now,
				"stopped_at":  &now,
			})
		}
	}

	// 5. 更新关键词表对应的数据
	if document.Enabled {
		// 6. 从禁用改为启用，需要新增关键词
		var enabledSegmentIDs []uuid.UUID
		for _, segment := range segments {
			if segment.Enabled {
				enabledSegmentIDs = append(enabledSegmentIDs, segment.ID)
			}
		}
		err = s.keywordTableService.AddKeywords(ctx, document.DatasetID, enabledSegmentIDs)
		if err != nil {
			log.Printf("添加关键词表失败: %v", err)
		}
	} else {
		// 7. 从启用改为禁用，需要剔除关键词
		err = s.keywordTableService.RemoveSegmentIDs(ctx, document.DatasetID, segmentIDs)
		if err != nil {
			log.Printf("删除关键词表失败: %v", err)
		}
	}

	return nil
}

// DeleteDocument 根据传递的知识库id+文档id删除文档信息
func (s *IndexingService) DeleteDocument(ctx context.Context, datasetID, documentID uuid.UUID) error {
	// 1. 查找该文档下的所有片段id列表
	segments, err := s.repo.GetSegmentsByDocumentID(ctx, documentID)
	if err != nil {
		return err
	}

	var segmentIDs []uuid.UUID
	for _, segment := range segments {
		segmentIDs = append(segmentIDs, segment.ID)
	}

	// 2. 调用向量数据库删除其关联记录
	err = s.vectorDatabaseService.DeleteDocumentsByID(ctx, "document_id", documentID.String())
	if err != nil {
		log.Printf("删除向量数据库记录失败: %v", err)
	}

	// 3. 删除postgres关联的segment记录
	err = s.repo.DeleteSegmentsByDocumentID(ctx, documentID)
	if err != nil {
		return err
	}

	// 4. 删除片段id对应的关键词记录
	err = s.keywordTableService.RemoveSegmentIDs(ctx, datasetID, segmentIDs)
	if err != nil {
		log.Printf("删除关键词表记录失败: %v", err)
	}

	return nil
}

// DeleteDataset 根据传递的知识库id执行相应的删除操作
func (s *IndexingService) DeleteDataset(ctx context.Context, datasetID uuid.UUID) error {
	// 1. 删除关联的文档记录
	err := s.repo.DeleteDocumentsByDatasetID(ctx, datasetID)
	if err != nil {
		log.Printf("删除文档记录失败: %v", err)
	}

	// 2. 删除关联的片段记录
	err = s.repo.DeleteSegmentsByDatasetID(ctx, datasetID)
	if err != nil {
		log.Printf("删除片段记录失败: %v", err)
	}

	// 3. 删除关联的关键词表记录
	err = s.repo.DeleteKeywordTablesByDatasetID(ctx, datasetID)
	if err != nil {
		log.Printf("删除关键词表记录失败: %v", err)
	}

	// 4. 删除知识库查询记录
	err = s.repo.DeleteDatasetQueriesByDatasetID(ctx, datasetID)
	if err != nil {
		log.Printf("删除知识库查询记录失败: %v", err)
	}

	// 5. 调用向量数据库删除知识库的关联记录
	err = s.vectorDatabaseService.DeleteDocumentsByID(ctx, "dataset_id", datasetID.String())
	if err != nil {
		log.Printf("删除向量数据库知识库记录失败: %v", err)
	}

	return nil
}

// parsing 解析传递的文档为LangChain文档列表
func (s *IndexingService) parsing(ctx context.Context, document *entity.Document) ([]schema.Document, error) {
	// 1. 获取upload_file并加载LangChain文档
	uploadFile, err := s.repo.GetUploadFileByID(ctx, document.UploadFileID)
	if err != nil {
		return nil, err
	}

	lcDocuments, err := s.fileExtractor.Load(ctx, uploadFile, false, true)
	if err != nil {
		return nil, err
	}

	// 2. 循环处理LangChain文档，并删除多余的空白字符串
	for _, lcDocument := range lcDocuments {
		lcDocument.PageContent = s.cleanExtraText(lcDocument.PageContent)
	}

	// 3. 更新文档状态并记录时间
	var characterCount int
	for _, lcDocument := range lcDocuments {
		characterCount += len(lcDocument.PageContent)
	}

	now := time.Now().UnixMilli()
	err = s.repo.UpdateDocument(ctx, document.ID, map[string]any{
		"character_count":      characterCount,
		"status":               consts.DocumentStatusSplitting,
		"parsing_completed_at": &now,
	})
	if err != nil {
		return nil, err
	}

	return lcDocuments, nil
}

// splitting 根据传递的信息进行文档分割，拆分成小块片段
func (s *IndexingService) splitting(ctx context.Context, document *entity.Document, lcDocuments []schema.Document) ([]schema.Document, error) {
	// 1. 根据process_rule获取文本分割器
	processRule, err := s.repo.GetProcessRuleByID(ctx, document.ProcessRuleID)
	if err != nil {
		return nil, err
	}

	textSplitter, err := s.processRuleService.GetTextSplitterByProcessRule(
		ctx,
		processRule,
		s.embeddingsService.CalculateTokenCount,
	)
	if err != nil {
		return nil, err
	}

	// 2. 按照process_rule规则清除多余的字符串
	for _, lcDocument := range lcDocuments {
		cleanedText, err := s.processRuleService.CleanTextByProcessRule(ctx, lcDocument.PageContent, processRule)
		if err != nil {
			return nil, err
		}
		lcDocument.PageContent = cleanedText
	}

	// 3. 分割文档列表为片段列表
	var lcSegments []schema.Document
	for _, lcDocument := range lcDocuments {
		chunks, err := textSplitter.SplitText(lcDocument.PageContent)
		if err != nil {
			return nil, err
		}

		for _, chunk := range chunks {
			lcSegments = append(lcSegments, schema.Document{
				PageContent: chunk,
				Metadata:    lcDocument.Metadata,
			})
		}
	}

	// 4. 获取对应文档下得到最大片段位置
	maxPosition, err := s.repo.GetMaxSegmentPosition(ctx, document.ID)
	if err != nil {
		maxPosition = 0
	}

	// 5. 循环处理片段数据并添加元数据，同时存储到postgres数据库中
	var segments []*entity.Segment
	for _, lcSegment := range lcSegments {
		maxPosition++
		content := lcSegment.PageContent
		nodeID := uuid.New()

		segment := &entity.Segment{
			ID:             uuid.New(),
			AccountID:      document.AccountID,
			DatasetID:      document.DatasetID,
			DocumentID:     document.ID,
			NodeID:         nodeID,
			Position:       maxPosition,
			Content:        content,
			CharacterCount: len(content),
			TokenCount:     s.embeddingsService.CalculateTokenCount(content),
			Hash:           util.GenerateHash(content),
			Status:         consts.SegmentStatusWaiting,
		}

		err = s.repo.CreateSegment(ctx, segment)
		if err != nil {
			return nil, err
		}

		lcSegment.Metadata = map[string]any{
			"account_id":       document.AccountID.String(),
			"dataset_id":       document.DatasetID.String(),
			"document_id":      document.ID.String(),
			"segment_id":       segment.ID.String(),
			"node_id":          nodeID.String(),
			"document_enabled": false,
			"segment_enabled":  false,
		}

		segments = append(segments, segment)
	}

	// 6. 更新文档的数据，涵盖状态、token数等内容
	var totalTokenCount int
	for _, segment := range segments {
		totalTokenCount += segment.TokenCount
	}

	now := time.Now().UnixMilli()
	err = s.repo.UpdateDocument(ctx, document.ID, map[string]any{
		"token_count":            totalTokenCount,
		"status":                 consts.DocumentStatusIndexing,
		"splitting_completed_at": &now,
	})
	if err != nil {
		return nil, err
	}

	return lcSegments, nil
}

// indexing 根据传递的信息构建索引，涵盖关键词提取、词表构建
func (s *IndexingService) indexing(ctx context.Context, document *entity.Document, lcSegments []schema.Document) error {
	for _, lcSegment := range lcSegments {
		// 1. 提取每一个片段对应的关键词，关键词的数量最多不超过10个
		keywords := s.jiebaService.ExtractKeywords(lcSegment.PageContent, 10)

		// 2. 逐条更新文档片段的关键词
		segmentID, _ := uuid.Parse(lcSegment.Metadata["segment_id"].(string))
		now := time.Now().UnixMilli()
		keywordsJSON, _ := sonic.Marshal(keywords)

		err := s.repo.UpdateSegment(ctx, segmentID, map[string]any{
			"keywords":              string(keywordsJSON),
			"status":                consts.SegmentStatusIndexing,
			"indexing_completed_at": &now,
		})
		if err != nil {
			log.Printf("更新片段关键词失败: %v", err)
			continue
		}

		// 3. 获取当前知识库的关键词表
		keywordTableRecord, err := s.keywordTableService.GetKeywordByDateSet(ctx, document.DatasetID)
		if err != nil {
			log.Printf("获取关键词表失败: %v", err)
			continue
		}

		// 4. 解析关键词表
		keywordTable := keywordTableRecord.KeywordMap
		// 5. 循环将新关键词添加到关键词表中
		for _, keyword := range keywords {
			if _, exists := keywordTable[keyword]; !exists {
				keywordTable[keyword] = []string{}
			}

			// 检查是否已存在该片段ID
			segmentIDStr := lcSegment.Metadata["segment_id"].(string)
			found := false
			for _, existingID := range keywordTable[keyword] {
				if existingID == segmentIDStr {
					found = true
					break
				}
			}

			if !found {
				keywordTable[keyword] = append(keywordTable[keyword], segmentIDStr)
			}
		}

		// 6. 更新关键词表
		updatedKeywordTableJSON, _ := sonic.Marshal(keywordTable)
		err = s.repo.UpdateKeywordTable(ctx, keywordTableRecord.ID, map[string]any{
			"keyword_table": string(updatedKeywordTableJSON),
		})
		if err != nil {
			log.Printf("更新关键词表失败: %v", err)
		}
	}

	// 7. 更新文档状态
	now := time.Now().UnixMilli()
	return s.repo.UpdateDocument(ctx, document.ID, map[string]any{
		"indexing_completed_at": &now,
	})
}

// completed 存储文档片段到向量数据库，并完成状态更新
func (s *IndexingService) completed(ctx context.Context, document *entity.Document, lcSegments []schema.Document) error {
	// 1. 循环遍历片段列表数据，将文档状态及片段状态设置成True
	for _, lcSegment := range lcSegments {
		lcSegment.Metadata["document_enabled"] = true
		lcSegment.Metadata["segment_enabled"] = true
	}

	// 2. 调用向量数据库，每次存储10条数据，避免一次传递过多的数据
	batchSize := 10
	for i := 0; i < len(lcSegments); i += batchSize {
		end := i + batchSize
		if end > len(lcSegments) {
			end = len(lcSegments)
		}

		chunks := lcSegments[i:end]
		var nodeIDs []uuid.UUID
		var documents []schema.Document

		for _, chunk := range chunks {
			documents = append(documents, schema.Document{
				PageContent: chunk.PageContent,
				Metadata:    chunk.Metadata,
			})
		}

		// 存储到向量数据库
		_, err := s.vectorDatabaseService.AddDocument(ctx, documents)
		if err != nil {
			log.Printf("构建文档片段索引发生异常, 错误信息: %v", err)

			// 更新片段状态为错误
			now := time.Now().UnixMilli()
			for _, nodeID := range nodeIDs {
				s.repo.UpdateSegmentByNodeID(ctx, nodeID, map[string]any{
					"status":       consts.SegmentStatusError,
					"completed_at": nil,
					"stopped_at":   &now,
					"enabled":      false,
					"error":        err.Error(),
				})
			}
		} else {
			// 更新片段状态为完成
			now := time.Now().UnixMilli()
			for _, nodeID := range nodeIDs {
				s.repo.UpdateSegmentByNodeID(ctx, nodeID, map[string]any{
					"status":       consts.SegmentStatusCompleted,
					"completed_at": &now,
					"enabled":      true,
				})
			}
		}
	}

	// 3. 更新文档的状态数据
	now := time.Now().UnixMilli()
	return s.repo.UpdateDocument(ctx, document.ID, map[string]any{
		"status":       consts.DocumentStatusCompleted,
		"completed_at": &now,
		"enabled":      true,
	})
}

// cleanExtraText 清除过滤传递的多余空白字符串
func (s *IndexingService) cleanExtraText(text string) string {
	text = regexp.MustCompile(`<\|`).ReplaceAllString(text, "<")
	text = regexp.MustCompile(`\|>`).ReplaceAllString(text, ">")
	text = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F\xEF\xBF\xBE]`).ReplaceAllString(text, "")
	text = strings.ReplaceAll(text, "\uFFFE", "") // 删除零宽非标记字符
	return text
}
