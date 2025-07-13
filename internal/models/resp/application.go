package resp

type GetAppConfResp struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	Icon                 string    `json:"icon"`
	Description          string    `json:"description"`
	PublishedAppConfigID string    `json:"publishedAppConfigID"` // uuid
	DraftedAppConfigID   string    `json:"draftedAppConfigID"`   // uuid
	DebugConversationID  string    `json:"debugConversationID"`  // uuid
	PublishedAppConfig   AppConfig `json:"publishedAppConfig"`
	DraftedAppConfig     AppConfig `json:"draftedAppConfig"`
	Utime                int64     `json:"utime"`
	Ctime                int64     `json:"ctime"`
}

type AppConfig struct {
	ID          string `json:"id"`
	ModelConfig struct {
		DialogRound int `json:"dialogRound"`
	} `json:"modelConfig"`
	MemoryMode string `json:"memoryMode"` // longTermMemory | none
	Status     string `json:"status"`     // drafted | published
	Utime      int64  `json:"utime"`
	Ctime      int64  `json:"ctime"`
}

type GetAppLTMResp struct {
	Summary string `json:"summary"`
}

type AppDebugChatResp struct {
	Content string `json:"content"`
}

type GetAppDebugHisListResp struct {
	List []struct {
		Id              string `json:"id"`
		ConversationId  string `json:"conversationId"`
		Query           string `json:"query"`
		Answer          string `json:"answer"`
		AnswerTokens    int    `json:"answerTokens"`
		ResponseLatency int    `json:"responseLatency"`
		Utime           int64  `json:"utime"`
		Ctime           int64  `json:"ctime"`
	} `json:"list"`
	Paginator struct {
		CurrentPage int `json:"currentPage"`
		PageSize    int `json:"pageSize"`
		TotalPage   int `json:"totalPage"`
		TotalRecord int `json:"totalRecord"`
	} `json:"paginator"`
}
