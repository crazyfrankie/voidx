package resp

import "github.com/google/uuid"

// GetApiKeysWithPageResp 获取API秘钥分页列表数据响应
type GetApiKeysWithPageResp struct {
	ID       uuid.UUID `json:"id"`
	ApiKey   string    `json:"api_key"`
	IsActive bool      `json:"is_active"`
	Remark   string    `json:"remark"`
	Utime    int64     `json:"utime"`
	Ctime    int64     `json:"ctime"`
}
