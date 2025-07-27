package req

// UploadFileReq 上传文件请求
type UploadFileReq struct {
	File     []byte `json:"file" binding:"required"`
	Filename string `json:"filename" binding:"required"`
	MimeType string `json:"mime_type"`
}

// UploadImageReq 上传图片请求
type UploadImageReq struct {
	File     []byte `json:"file" binding:"required"`
	Filename string `json:"filename" binding:"required"`
	MimeType string `json:"mime_type"`
}
