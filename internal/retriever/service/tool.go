package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/pkg/consts"
)

type DatasetRetrievalInput struct {
	Query string `json:"query" description:"知识库搜索query语句，类型为字符串"`
}

// datasetRetrievalTool 实现了tools.Tool接口
type datasetRetrievalTool struct {
	service           *RetrievalService
	userID            uuid.UUID
	datasetIDs        []uuid.UUID
	resource          consts.RetrievalSource
	retrievalConfig   map[string]any
	k                 int
	score             float32
	retrievalStrategy consts.RetrievalStrategy
	inputSchema       interface{}
}

func (t *datasetRetrievalTool) Name() string {
	return "dataset_retrieval_tool"
}

func (t *datasetRetrievalTool) Call(ctx context.Context, input string) (string, error) {
	// 执行搜索
	documents, err := t.service.SearchInDatasets(
		ctx,
		t.userID,
		req.SearchRequest{
			Query:          input,
			DatasetIDs:     t.datasetIDs,
			RetrieverType:  string(t.retrievalStrategy),
			K:              t.k,
			ScoreThreshold: t.score,
		},
	)
	if err != nil {
		return "", fmt.Errorf("search failed: %w", err)
	}

	// 处理结果
	if len(documents) == 0 {
		return "知识库内没有检索到对应内容", nil
	}

	// 合并文档
	combined, err := combineDocuments(documents)
	if err != nil {
		return "", fmt.Errorf("failed to combine documents: %w", err)
	}

	return combined, nil
}

func (t *datasetRetrievalTool) Description() string {
	return "如果需要搜索扩展的知识库内容，当你觉得用户的提问超过你的知识范围时，可以尝试调用该工具，输入为搜索query语句，返回数据为检索内容字符串"
}

func (t *datasetRetrievalTool) Args() interface{} {
	return t.inputSchema
}

// combineDocuments 合并文档为字符串
func combineDocuments(docs []resp.SearchResult) (string, error) {
	var builder strings.Builder
	for _, doc := range docs {
		if _, err := builder.WriteString(doc.Content + "\n\n"); err != nil {
			return "", err
		}
	}
	return builder.String(), nil
}
