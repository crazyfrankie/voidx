package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/crazyfrankie/voidx/types/errno"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/process_rule/repository"
)

type ProcessRuleService struct {
	repo *repository.ProcessRuleRepo
}

func NewProcessRuleService(repo *repository.ProcessRuleRepo) *ProcessRuleService {
	return &ProcessRuleService{
		repo: repo,
	}
}

// TextSplitterConfig 文本分割器配置
type TextSplitterConfig struct {
	ChunkSize    int      `json:"chunk_size"`
	ChunkOverlap int      `json:"chunk_overlap"`
	Separators   []string `json:"separators"`
}

// TextSplitter 文本分割器接口
type TextSplitter interface {
	SplitText(text string) ([]string, error)
}

// RecursiveCharacterTextSplitter 递归字符文本分割器
type RecursiveCharacterTextSplitter struct {
	ChunkSize        int
	ChunkOverlap     int
	Separators       []string
	IsSeparatorRegex bool
	LengthFunction   func(string) int
}

// NewRecursiveCharacterTextSplitter 创建递归字符文本分割器
func NewRecursiveCharacterTextSplitter(config TextSplitterConfig, lengthFunction func(string) int) *RecursiveCharacterTextSplitter {
	if lengthFunction == nil {
		lengthFunction = func(s string) int { return len(s) }
	}

	return &RecursiveCharacterTextSplitter{
		ChunkSize:        config.ChunkSize,
		ChunkOverlap:     config.ChunkOverlap,
		Separators:       config.Separators,
		IsSeparatorRegex: true,
		LengthFunction:   lengthFunction,
	}
}

// SplitText 分割文本
func (r *RecursiveCharacterTextSplitter) SplitText(text string) ([]string, error) {
	return r.splitTextRecursive(text, r.Separators)
}

// splitTextRecursive 递归分割文本
func (r *RecursiveCharacterTextSplitter) splitTextRecursive(text string, separators []string) ([]string, error) {
	var finalChunks []string

	// 如果没有分隔符，直接按长度分割
	if len(separators) == 0 {
		return r.splitByLength(text), nil
	}

	separator := separators[0]
	newSeparators := separators[1:]

	var splits []string
	if r.IsSeparatorRegex {
		// 使用正则表达式分割
		re, err := regexp.Compile(separator)
		if err != nil {
			return nil, err
		}
		splits = re.Split(text, -1)
	} else {
		// 使用字符串分割
		splits = strings.Split(text, separator)
	}

	var goodSplits []string
	for _, split := range splits {
		if r.LengthFunction(split) < r.ChunkSize {
			goodSplits = append(goodSplits, split)
		} else {
			if len(goodSplits) > 0 {
				mergedText := r.mergeSplits(goodSplits, separator)
				finalChunks = append(finalChunks, mergedText...)
				goodSplits = nil
			}

			// 递归处理过长的分割
			if len(newSeparators) == 0 {
				finalChunks = append(finalChunks, r.splitByLength(split)...)
			} else {
				otherInfo, err := r.splitTextRecursive(split, newSeparators)
				if err != nil {
					return nil, err
				}
				finalChunks = append(finalChunks, otherInfo...)
			}
		}
	}

	if len(goodSplits) > 0 {
		mergedText := r.mergeSplits(goodSplits, separator)
		finalChunks = append(finalChunks, mergedText...)
	}

	return finalChunks, nil
}

// mergeSplits 合并分割的文本
func (r *RecursiveCharacterTextSplitter) mergeSplits(splits []string, separator string) []string {
	var docs []string
	var currentDoc []string
	total := 0

	for _, split := range splits {
		length := r.LengthFunction(split)
		if total+length+(len(currentDoc)*len(separator)) > r.ChunkSize && len(currentDoc) > 0 {
			if len(currentDoc) > 0 {
				doc := strings.Join(currentDoc, separator)
				if strings.TrimSpace(doc) != "" {
					docs = append(docs, doc)
				}

				// 处理重叠
				for total > r.ChunkOverlap || (total+length+(len(currentDoc)*len(separator)) > r.ChunkSize && total > 0) {
					if len(currentDoc) == 0 {
						break
					}
					total -= r.LengthFunction(currentDoc[0]) + len(separator)
					currentDoc = currentDoc[1:]
				}
			}
		}

		currentDoc = append(currentDoc, split)
		total += length + len(separator)
	}

	if len(currentDoc) > 0 {
		doc := strings.Join(currentDoc, separator)
		if strings.TrimSpace(doc) != "" {
			docs = append(docs, doc)
		}
	}

	return docs
}

// splitByLength 按长度分割文本
func (r *RecursiveCharacterTextSplitter) splitByLength(text string) []string {
	var chunks []string

	for len(text) > r.ChunkSize {
		chunks = append(chunks, text[:r.ChunkSize])
		start := r.ChunkSize - r.ChunkOverlap
		if start < 0 {
			start = 0
		}
		text = text[start:]
	}

	if len(text) > 0 {
		chunks = append(chunks, text)
	}

	return chunks
}

// GetTextSplitterByProcessRule 根据传递的处理规则+长度计算函数，获取相应的文本分割器
func (s *ProcessRuleService) GetTextSplitterByProcessRule(
	ctx context.Context,
	processRule *entity.ProcessRule,
	lengthFunction func(string) int,
) (TextSplitter, error) {
	// 解析规则JSON
	rule := processRule.Rule

	// 提取分段配置
	segment, ok := rule["segment"].(map[string]any)
	if !ok {
		return nil, errno.ErrInternalServer.AppendBizMessage(errors.New("获取应用的调试会话消息列表"))
	}

	config := TextSplitterConfig{}

	if chunkSize, ok := segment["chunk_size"].(float64); ok {
		config.ChunkSize = int(chunkSize)
	} else {
		config.ChunkSize = 1000 // 默认值
	}

	if chunkOverlap, ok := segment["chunk_overlap"].(float64); ok {
		config.ChunkOverlap = int(chunkOverlap)
	} else {
		config.ChunkOverlap = 200 // 默认值
	}

	if separators, ok := segment["separators"].([]any); ok {
		for _, sep := range separators {
			if sepStr, ok := sep.(string); ok {
				config.Separators = append(config.Separators, sepStr)
			}
		}
	} else {
		// 默认分隔符
		config.Separators = []string{"\n\n", "\n", " ", ""}
	}

	return NewRecursiveCharacterTextSplitter(config, lengthFunction), nil
}

// CleanTextByProcessRule 根据传递的处理规则清除多余的字符串
func (s *ProcessRuleService) CleanTextByProcessRule(ctx context.Context, text string, processRule *entity.ProcessRule) (string, error) {
	rule := processRule.Rule

	// 获取预处理规则
	preProcessRules, ok := rule["pre_process_rules"].([]any)
	if !ok {
		return text, nil
	}

	// 1. 循环遍历所有预处理规则
	for _, ruleItem := range preProcessRules {
		preProcessRule, ok := ruleItem.(map[string]any)
		if !ok {
			continue
		}

		id, ok := preProcessRule["id"].(string)
		if !ok {
			continue
		}

		enabled, ok := preProcessRule["enabled"].(bool)
		if !ok || !enabled {
			continue
		}

		// 2. 删除多余空格
		if id == "remove_extra_space" {
			// 删除3个或更多连续换行符，替换为2个换行符
			pattern := regexp.MustCompile(`\n{3,}`)
			text = pattern.ReplaceAllString(text, "\n\n")

			// 删除2个或更多连续的空白字符，替换为1个空格
			pattern = regexp.MustCompile(`[\t\f\r\x20\u00a0\u1680\u180e\u2000-\u200a\u202f\u205f\u3000]{2,}`)
			text = pattern.ReplaceAllString(text, " ")
		}

		// 3. 删除多余的URL链接及邮箱
		if id == "remove_url_and_email" {
			// 删除邮箱地址
			pattern := regexp.MustCompile(`([a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+)`)
			text = pattern.ReplaceAllString(text, "")

			// 删除HTTP/HTTPS链接
			pattern = regexp.MustCompile(`https?://[^\s]+`)
			text = pattern.ReplaceAllString(text, "")
		}
	}

	return text, nil
}

// CreateProcessRule 创建处理规则
func (s *ProcessRuleService) CreateProcessRule(ctx context.Context, processRule *entity.ProcessRule) error {
	return s.repo.CreateProcessRule(ctx, processRule)
}

// GetProcessRuleByID 根据ID获取处理规则
func (s *ProcessRuleService) GetProcessRuleByID(ctx context.Context, id uuid.UUID) (*entity.ProcessRule, error) {
	return s.repo.GetProcessRuleByID(ctx, id)
}

// GetProcessRuleByDatasetID 根据数据集ID获取处理规则
func (s *ProcessRuleService) GetProcessRuleByDatasetID(ctx context.Context, datasetID uuid.UUID) (*entity.ProcessRule, error) {
	return s.repo.GetProcessRuleByDatasetID(ctx, datasetID)
}

// UpdateProcessRule 更新处理规则
func (s *ProcessRuleService) UpdateProcessRule(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return s.repo.UpdateProcessRule(ctx, id, updates)
}

// DeleteProcessRule 删除处理规则
func (s *ProcessRuleService) DeleteProcessRule(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteProcessRule(ctx, id)
}
