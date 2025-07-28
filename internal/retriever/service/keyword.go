package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/retriever/repository"
)

const lockKeyPrefix = "lock:keyword_table:"

type KeywordService struct {
	repo *repository.KeywordRepository
}

// NewKeywordService 创建一个新的关键词表服务
func NewKeywordService(repo *repository.KeywordRepository) *KeywordService {
	return &KeywordService{
		repo: repo,
	}
}

// GetKeywordByDateSet 获取知识库的关键词表
func (s *KeywordService) GetKeywordByDateSet(ctx context.Context, datasetID uuid.UUID) (*entity.Keyword, error) {
	// 从数据库获取关键词表
	keyword, err := s.repo.GetByDatasetID(ctx, datasetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get keyword table: %w", err)
	}

	if keyword == nil {
		keyword = &entity.Keyword{
			DatasetID: datasetID,
		}
		if err := s.repo.Create(ctx, keyword); err != nil {
			return nil, err
		}
	}

	return keyword, nil
}

// AddKeywords 向知识库的关键词表添加关键词
func (s *KeywordService) AddKeywords(ctx context.Context, datasetID uuid.UUID, segmentID []uuid.UUID) error {
	// 1. 获取锁
	lockKey := getLockKey(datasetID)
	lock := s.repo.AcquireLock(ctx, lockKey)
	if lock == "" {
		return fmt.Errorf("failed to acquire lock for dataset %s", datasetID)
	}
	defer s.repo.ReleaseLock(ctx, lockKey, lock)

	// 2. 获取关键词表记录
	record, err := s.repo.GetByDatasetID(ctx, datasetID)
	if err != nil {
		return fmt.Errorf("获取关键词表记录失败: %w", err)
	}

	// 3. 转换keyword_table为map[string]map[string]struct{}
	keywordTable := make(map[string]map[string]struct{})
	for field, keywords := range record.KeywordMap {
		keywordSet := make(map[string]struct{})
		for _, kw := range keywords {
			keywordSet[kw] = struct{}{}
		}
		keywordTable[field] = keywordSet
	}

	// 4. 查询片段信息
	segments, err := s.repo.GetKeywordBySegments(ctx, segmentID)
	if err != nil {
		return fmt.Errorf("查询片段失败: %w", err)
	}

	// 5. 更新关键词表
	for _, seg := range segments {
		for _, keyword := range seg.Keywords {
			if _, exists := keywordTable[keyword]; !exists {
				keywordTable[keyword] = make(map[string]struct{})
			}
			keywordTable[keyword][seg.ID.String()] = struct{}{}
		}
	}

	// 6. 转换回存储格式并更新
	updateMap := make(map[string][]string)
	for field, idSet := range keywordTable {
		ids := make([]string, 0, len(idSet))
		for id := range idSet {
			ids = append(ids, id)
		}
		updateMap[field] = ids
	}

	return s.repo.Update(ctx, &entity.Keyword{
		ID:         record.ID,
		DatasetID:  datasetID,
		KeywordMap: updateMap,
	})
}

// RemoveSegmentIDs 从知识库的关键词表中移除指定的片段ID
func (s *KeywordService) RemoveSegmentIDs(ctx context.Context, datasetID uuid.UUID, segmentIDs []uuid.UUID) error {
	// 获取锁
	lockKey := getLockKey(datasetID)
	lock := s.repo.AcquireLock(ctx, lockKey)
	if lock == "" {
		return fmt.Errorf("failed to acquire lock for dataset %s", datasetID)
	}
	defer s.repo.ReleaseLock(ctx, lockKey, lock)

	// 获取现有的关键词表
	keywordTable, err := s.GetKeywordByDateSet(ctx, datasetID)
	if err != nil {
		return fmt.Errorf("failed to get keyword table: %w", err)
	}

	// 创建segmentID集合，用于快速查找
	segmentIDSet := make(map[string]struct{}, len(segmentIDs))
	for _, id := range segmentIDs {
		segmentIDSet[id.String()] = struct{}{}
	}

	// 从每个关键词的片段ID列表中移除指定的片段ID
	for keyword, ids := range keywordTable.KeywordMap {
		newIDs := make([]string, 0, len(ids))
		for _, id := range ids {
			if _, exists := segmentIDSet[id]; !exists {
				newIDs = append(newIDs, id)
			}
		}

		if len(newIDs) > 0 {
			keywordTable.KeywordMap[keyword] = newIDs
		} else {
			delete(keywordTable.KeywordMap, keyword)
		}
	}

	// 保存关键词表
	return s.repo.Update(ctx, keywordTable)
}

func getLockKey(datasetID uuid.UUID) string {
	return lockKeyPrefix + datasetID.String()
}
