package req

type UpdateAppConfReq struct {
	ModelConfig struct {
		DialogRound int `json:"dialogRound"`
	} `json:"modelConfig"`
	MemoryMode string `json:"memoryMode"`
}

type UpdateAppDebugLTMReq struct {
	Summary string `json:"summary"`
}

type ChatReq struct {
	Query string `json:"query" binding:"required,max=2000"`
}
