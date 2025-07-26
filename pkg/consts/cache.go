package consts

// Cache相关常量定义

// LockExpireTime 缓存锁的过期时间，单位为秒，默认为600
const LockExpireTime = 600

// 缓存锁键名模板
const (
	// LockDocumentUpdateEnabled 更新文档启用状态缓存锁
	LockDocumentUpdateEnabled = "lock:document:update:enabled_{document_id}"

	// LockKeywordTableUpdateKeywordTable 更新关键词表缓存锁
	LockKeywordTableUpdateKeywordTable = "lock:keyword_table:update:keyword_table_{dataset_id}"

	// LockSegmentUpdateEnabled 更新片段启用状态缓存锁
	LockSegmentUpdateEnabled = "lock:segment:update:enabled_{segment_id}"
)
