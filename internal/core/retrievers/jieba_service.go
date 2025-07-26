package retrievers

import (
	"strings"

	"github.com/yanyiwu/gojieba"
)

// JiebaService 提供中文分词和关键词提取功能
type JiebaService struct {
	jieba    *gojieba.Jieba
	stopword map[string]struct{}
}

// NewJiebaService 创建一个新的JiebaService实例
func NewJiebaService(stopwordSet map[string]struct{}) *JiebaService {
	jieba := gojieba.NewJieba()

	// 如果没有提供停用词集合，则使用默认的空集合
	if stopwordSet == nil {
		stopwordSet = make(map[string]struct{})
	}

	return &JiebaService{
		jieba:    jieba,
		stopword: stopwordSet,
	}
}

// Close 释放Jieba资源
func (s *JiebaService) Close() {
	s.jieba.Free()
}

// ExtractKeywords 从文本中提取关键词
func (s *JiebaService) ExtractKeywords(text string, topK int) []string {
	// 使用jieba提取关键词
	words := s.jieba.ExtractWithWeight(text, topK)

	// 过滤停用词并提取关键词
	keywords := make([]string, 0, len(words))
	for _, word := range words {
		// 转换为小写并去除空格
		processedWord := strings.ToLower(strings.TrimSpace(word.Word))

		// 跳过空字符串
		if processedWord == "" {
			continue
		}

		// 跳过停用词
		if _, isStopword := s.stopword[processedWord]; isStopword {
			continue
		}

		keywords = append(keywords, processedWord)
	}

	return keywords
}

// CutForSearch 对文本进行搜索引擎模式分词
func (s *JiebaService) CutForSearch(text string) []string {
	return s.jieba.CutForSearch(text, true)
}

// Cut 对文本进行精确模式分词
func (s *JiebaService) Cut(text string) []string {
	return s.jieba.Cut(text, true)
}

// AddWord 向分词词典中添加新词
func (s *JiebaService) AddWord(word string) {
	s.jieba.AddWord(word)
}

// LoadStopwords 加载停用词集合
func LoadStopwords(stopwordList []string) map[string]struct{} {
	stopwordSet := make(map[string]struct{}, len(stopwordList))
	for _, word := range stopwordList {
		processedWord := strings.ToLower(strings.TrimSpace(word))
		if processedWord != "" {
			stopwordSet[processedWord] = struct{}{}
		}
	}
	return stopwordSet
}
