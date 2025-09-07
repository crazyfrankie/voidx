package consts

// Dataset相关常量定义

// DefaultDatasetDescriptionFormatter 默认知识库描述格式化文本
const DefaultDatasetDescriptionFormatter = "当你需要回答管理《{name}》的时候可以引用该知识库。"

// ProcessType 文档处理规则类型枚举
type ProcessType string

const (
	ProcessTypeAutomatic ProcessType = "automatic"
	ProcessTypeCustom    ProcessType = "custom"
)

// DefaultProcessRule 默认的处理规则
var DefaultProcessRule = map[string]any{
	"mode": "custom",
	"rule": map[string]any{
		"pre_process_rules": []any{
			map[string]any{"id": "remove_extra_space", "enabled": true},
			map[string]any{"id": "remove_url_and_email", "enabled": true},
		},
		"segment": map[string]any{
			"separators": []any{
				"\n\n",
				"\n",
				"。|！|？",
				"\\.\\s|\\!\\s|\\?\\s", // 英文标点符号后面通常需要加空格
				"；|;\\s",
				"，|,\\s",
				" ",
				"",
			},
			"chunk_size":    500,
			"chunk_overlap": 50,
		},
	},
}

// DocumentStatus 文档状态类型枚举
type DocumentStatus string

const (
	DocumentStatusWaiting   DocumentStatus = "waiting"
	DocumentStatusParsing   DocumentStatus = "parsing"
	DocumentStatusSplitting DocumentStatus = "splitting"
	DocumentStatusIndexing  DocumentStatus = "indexing"
	DocumentStatusCompleted DocumentStatus = "completed"
	DocumentStatusError     DocumentStatus = "error"
)

// SegmentStatus 片段状态类型枚举
type SegmentStatus string

const (
	SegmentStatusWaiting   SegmentStatus = "waiting"
	SegmentStatusIndexing  SegmentStatus = "indexing"
	SegmentStatusCompleted SegmentStatus = "completed"
	SegmentStatusError     SegmentStatus = "error"
)

// RetrievalStrategy 检索策略类型枚举
type RetrievalStrategy string

const (
	RetrievalStrategyFullText RetrievalStrategy = "full_text"
	RetrievalStrategySemantic RetrievalStrategy = "semantic"
	RetrievalStrategyHybrid   RetrievalStrategy = "hybrid"
)

// RetrievalSource 检索来源
type RetrievalSource string

const (
	RetrievalSourceHitTesting RetrievalSource = "hit_testing"
	RetrievalSourceApp        RetrievalSource = "app"
)
