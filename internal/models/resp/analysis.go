package resp

// AppAnalysisResp 应用分析响应
type AppAnalysisResp struct {
	TotalMessages                  IndicatorData  `json:"total_messages"`
	ActiveAccounts                 IndicatorData  `json:"active_accounts"`
	AvgOfConversationMessages      IndicatorData  `json:"avg_of_conversation_messages"`
	TokenOutputRate                IndicatorData  `json:"token_output_rate"`
	CostConsumption                IndicatorData  `json:"cost_consumption"`
	TotalMessagesTrend             TrendItem      `json:"total_messages_trend"`
	ActiveAccountsTrend            TrendItem      `json:"active_accounts_trend"`
	AvgOfConversationMessagesTrend TrendItemFloat `json:"avg_of_conversation_messages_trend"`
	CostConsumptionTrend           TrendItemFloat `json:"cost_consumption_trend"`
}

// IndicatorData 指标数据
type IndicatorData struct {
	Data float64 `json:"data"`
	Pop  float64 `json:"pop"`
}

// TrendItem 趋势数据项（整数）
type TrendItem struct {
	XAxis []int64 `json:"x_axis"`
	YAxis []int   `json:"y_axis"`
}

// TrendItemFloat 趋势数据项（浮点数）
type TrendItemFloat struct {
	XAxis []int64   `json:"x_axis"`
	YAxis []float64 `json:"y_axis"`
}

// OverviewIndicators 概览指标
type OverviewIndicators struct {
	TotalMessages             int     `json:"total_messages"`
	ActiveAccounts            int     `json:"active_accounts"`
	AvgOfConversationMessages float64 `json:"avg_of_conversation_messages"`
	TokenOutputRate           float64 `json:"token_output_rate"`
	CostConsumption           float64 `json:"cost_consumption"`
}

// POPIndicators 环比指标
type POPIndicators struct {
	TotalMessages             float64 `json:"total_messages"`
	ActiveAccounts            float64 `json:"active_accounts"`
	AvgOfConversationMessages float64 `json:"avg_of_conversation_messages"`
	TokenOutputRate           float64 `json:"token_output_rate"`
	CostConsumption           float64 `json:"cost_consumption"`
}

// TrendData 趋势数据
type TrendData struct {
	TotalMessagesTrend             TrendItem      `json:"total_messages_trend"`
	ActiveAccountsTrend            TrendItem      `json:"active_accounts_trend"`
	AvgOfConversationMessagesTrend TrendItemFloat `json:"avg_of_conversation_messages_trend"`
	CostConsumptionTrend           TrendItemFloat `json:"cost_consumption_trend"`
}
