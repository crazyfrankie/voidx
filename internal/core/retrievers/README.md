# 检索器模块 (Retrievers)

本模块提供了多种检索器实现，用于从知识库中检索相关文档。

## 检索器类型

### 全文检索器 (FullTextRetriever)

基于关键词匹配的检索器，使用Jieba分词提取查询关键词，并在数据库中查找匹配的关键词表和片段ID，然后根据频率排序获取相关文档。

```go
fullTextRetriever := retrievers.NewFullTextRetriever(
    datasetIDs,
    jiebaService,
    options,
)
```

### 语义检索器 (SemanticRetriever)

基于向量相似度的检索器，使用嵌入模型将查询文本转换为向量，然后在向量存储中查找相似的文档。

```go
semanticRetriever := retrievers.NewSemanticRetriever(
    vectorStore,
    embedder,
    datasetIDs,
    options,
)
```

### 混合检索器 (HybridRetriever)

结合全文检索和语义检索的优点，可以提供更全面的检索结果。

```go
hybridRetriever := retrievers.NewHybridRetriever(
    fullTextRetriever,
    semanticRetriever,
    datasetIDs,
    options,
)
```

## 检索器工厂

检索器工厂提供了统一的接口来创建不同类型的检索器。

```go
retrieverFactory := retrievers.NewRetrieverFactory(
    vectorStore,
    embedder,
    jiebaService,
)

// 创建检索器
retriever, err := retrieverFactory.CreateRetriever(
    retrievers.RetrieverTypeHybrid,
    datasetIDs,
    options,
)
```

## 使用示例

```go
// 创建检索服务
retrievalService := service.NewRetrievalService(retrieverFactory)

// 执行检索
response, err := retrievalService.Search(ctx, &req.SearchRequest{
    Query:      "如何使用检索器？",
    DatasetIDs: []string{"dataset-id-1", "dataset-id-2"},
    RetrieverType: "hybrid",
    K: 4,
})
```

## API接口

```
POST /api/v1/retrieval/search
```

请求体：

```json
{
  "query": "如何使用检索器？",
  "dataset_ids": ["dataset-id-1", "dataset-id-2"],
  "retriever_type": "hybrid",
  "k": 4,
  "score_threshold": 0.5,
  "options": {}
}
```

响应体：

```json
{
  "results": [
    {
      "content": "检索器是一种用于从知识库中检索相关文档的工具...",
      "metadata": {
        "account_id": "account-id",
        "dataset_id": "dataset-id-1",
        "document_id": "document-id",
        "segment_id": "segment-id",
        "node_id": "node-id",
        "document_enabled": true,
        "segment_enabled": true,
        "score": 0.85,
        "retrieval_method": "hybrid"
      }
    }
  ],
  "count": 1
}
```