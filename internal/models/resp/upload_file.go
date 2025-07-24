package resp

import "github.com/google/uuid"

// UploadFileResp 上传文件响应
type UploadFileResp struct {
	ID        uuid.UUID `json:"id"`
	AccountID uuid.UUID `json:"account_id"`
	Name      string    `json:"name"`
	Key       string    `json:"key"`
	Size      int64     `json:"size"`
	Extension string    `json:"extension"`
	URL       string    `json:"url"`
	Ctime     int64     `json:"ctime"`
}
