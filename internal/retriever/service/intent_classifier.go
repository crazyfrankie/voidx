package service

import (
	"regexp"
	"strings"
)

type IntentClassifier struct {
	greetingPatterns []string
	chitchatPatterns []string
	questionWords    []string
}

func NewIntentClassifier() *IntentClassifier {
	return &IntentClassifier{
		greetingPatterns: []string{
			"你好", "hello", "hi", "嗨", "早上好", "下午好", "晚上好",
			"再见", "bye", "拜拜", "谢谢", "thank you", "thanks",
		},
		chitchatPatterns: []string{
			"怎么样", "如何", "好的", "ok", "嗯", "哦", "是的", "不是",
		},
		questionWords: []string{
			"什么", "怎么", "为什么", "哪里", "什么时候", "谁", "如何",
			"what", "how", "why", "where", "when", "who", "which",
		},
	}
}

func (ic *IntentClassifier) ShouldRetrieve(query string) bool {
	query = strings.TrimSpace(strings.ToLower(query))
	
	// 长度过短
	if len(query) < 3 {
		return false
	}
	
	// 问候语检测
	for _, pattern := range ic.greetingPatterns {
		if strings.Contains(query, pattern) && len(query) < 10 {
			return false
		}
	}
	
	// 闲聊检测
	for _, pattern := range ic.chitchatPatterns {
		if query == pattern {
			return false
		}
	}
	
	// 包含疑问词或长度足够，认为需要检索
	for _, word := range ic.questionWords {
		if strings.Contains(query, word) {
			return true
		}
	}
	
	return len(query) > 8
}

func (ic *IntentClassifier) containsTechnicalTerms(query string) bool {
	// 简单的技术词汇检测
	technicalPatterns := []string{
		"api", "数据库", "算法", "配置", "部署", "错误", "异常",
		"database", "config", "deploy", "error", "exception",
	}
	
	query = strings.ToLower(query)
	for _, term := range technicalPatterns {
		if strings.Contains(query, term) {
			return true
		}
	}
	return false
}
