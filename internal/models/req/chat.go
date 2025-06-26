package req

type ChatReq struct {
	Query string `json:"query" binding:"required,max=2000"`
}
