package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/analysis/repository"
	"github.com/crazyfrankie/voidx/internal/app"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/pkg/errno"
)

type AnalysisService struct {
	repo   *repository.AnalysisRepo
	appSvc *app.Service
}

func NewAnalysisService(repo *repository.AnalysisRepo, appSvc *app.Service) *AnalysisService {
	return &AnalysisService{
		repo:   repo,
		appSvc: appSvc,
	}
}

// GetAppAnalysis 根据传递的应用id+账号获取指定应用的分析信息
func (s *AnalysisService) GetAppAnalysis(ctx context.Context, appID, userID uuid.UUID) (*resp.AppAnalysisResp, error) {
	// 1. 验证应用权限
	application, err := s.appSvc.GetApp(ctx, appID, userID)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage("应用不存在")
	}

	if application.AccountID != userID {
		return nil, errno.ErrForbidden.AppendBizMessage("无权限访问该应用")
	}

	// 2. 尝试从缓存获取数据
	if analysis, err := s.repo.GetAppAnalysisFromCache(ctx, appID); err == nil && analysis != nil {
		return analysis, nil
	}

	// 3. 计算时间范围
	now := time.Now()
	todayMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	sevenDaysAgo := todayMidnight.AddDate(0, 0, -7)
	fourteenDaysAgo := todayMidnight.AddDate(0, 0, -14)

	// 4. 查询消息数据
	sevenDaysMessages, err := s.repo.GetMessagesByTimeRange(ctx, application.ID, sevenDaysAgo, todayMidnight)
	if err != nil {
		return nil, err
	}

	fourteenDaysMessages, err := s.repo.GetMessagesByTimeRange(ctx, application.ID, fourteenDaysAgo, sevenDaysAgo)
	if err != nil {
		return nil, err
	}

	// 5. 计算指标
	sevenOverviewIndicators := s.calculateOverviewIndicators(sevenDaysMessages)
	fourteenOverviewIndicators := s.calculateOverviewIndicators(fourteenDaysMessages)

	// 6. 计算环比
	pop := s.calculatePOP(sevenOverviewIndicators, fourteenOverviewIndicators)

	// 7. 计算趋势
	trend := s.calculateTrend(todayMidnight, 7, sevenDaysMessages)

	// 8. 构建响应
	analysis := &resp.AppAnalysisResp{
		TotalMessages: resp.IndicatorData{
			Data: float64(sevenOverviewIndicators.TotalMessages),
			Pop:  pop.TotalMessages,
		},
		ActiveAccounts: resp.IndicatorData{
			Data: float64(sevenOverviewIndicators.ActiveAccounts),
			Pop:  pop.ActiveAccounts,
		},
		AvgOfConversationMessages: resp.IndicatorData{
			Data: sevenOverviewIndicators.AvgOfConversationMessages,
			Pop:  pop.AvgOfConversationMessages,
		},
		TokenOutputRate: resp.IndicatorData{
			Data: sevenOverviewIndicators.TokenOutputRate,
			Pop:  pop.TokenOutputRate,
		},
		CostConsumption: resp.IndicatorData{
			Data: sevenOverviewIndicators.CostConsumption,
			Pop:  pop.CostConsumption,
		},
		TotalMessagesTrend:             trend.TotalMessagesTrend,
		ActiveAccountsTrend:            trend.ActiveAccountsTrend,
		AvgOfConversationMessagesTrend: trend.AvgOfConversationMessagesTrend,
		CostConsumptionTrend:           trend.CostConsumptionTrend,
	}

	// 9. 存储到缓存
	if err := s.repo.SetAppAnalysisToCache(ctx, appID, analysis); err != nil {

	}

	return analysis, nil
}

// calculateOverviewIndicators 计算概览指标
func (s *AnalysisService) calculateOverviewIndicators(messages []entity.Message) resp.OverviewIndicators {
	totalMessages := len(messages)

	// 计算激活用户数
	activeAccountsMap := make(map[uuid.UUID]bool)
	for _, msg := range messages {
		activeAccountsMap[msg.CreatedBy] = true
	}
	activeAccounts := len(activeAccountsMap)

	// 计算平均会话互动数
	conversationMap := make(map[uuid.UUID]bool)
	for _, msg := range messages {
		conversationMap[msg.ConversationID] = true
	}
	conversationCount := len(conversationMap)

	var avgOfConversationMessages float64
	if conversationCount > 0 {
		avgOfConversationMessages = float64(totalMessages) / float64(conversationCount)
	}

	// 计算Token输出速度和费用消耗
	var totalTokens int
	var totalLatency float64
	var costConsumption float64

	for _, msg := range messages {
		totalTokens += msg.TotalTokenCount
		totalLatency += msg.Latency
		costConsumption += msg.TotalPrice
	}

	var tokenOutputRate float64
	if totalLatency > 0 {
		tokenOutputRate = float64(totalTokens) / totalLatency
	}

	return resp.OverviewIndicators{
		TotalMessages:             totalMessages,
		ActiveAccounts:            activeAccounts,
		AvgOfConversationMessages: avgOfConversationMessages,
		TokenOutputRate:           tokenOutputRate,
		CostConsumption:           costConsumption,
	}
}

// calculatePOP 计算环比增长
func (s *AnalysisService) calculatePOP(current, previous resp.OverviewIndicators) resp.POPIndicators {
	calculatePOPValue := func(current, previous float64) float64 {
		if previous != 0 {
			return (current - previous) / previous
		}
		return 0
	}

	return resp.POPIndicators{
		TotalMessages:             calculatePOPValue(float64(current.TotalMessages), float64(previous.TotalMessages)),
		ActiveAccounts:            calculatePOPValue(float64(current.ActiveAccounts), float64(previous.ActiveAccounts)),
		AvgOfConversationMessages: calculatePOPValue(current.AvgOfConversationMessages, previous.AvgOfConversationMessages),
		TokenOutputRate:           calculatePOPValue(current.TokenOutputRate, previous.TokenOutputRate),
		CostConsumption:           calculatePOPValue(current.CostConsumption, previous.CostConsumption),
	}
}

// calculateTrend 计算趋势数据
func (s *AnalysisService) calculateTrend(endAt time.Time, daysAgo int, messages []entity.Message) resp.TrendData {
	endAtMidnight := time.Date(endAt.Year(), endAt.Month(), endAt.Day(), 0, 0, 0, 0, endAt.Location())

	totalMessagesTrend := resp.TrendItem{XAxis: []int64{}, YAxis: []int{}}
	activeAccountsTrend := resp.TrendItem{XAxis: []int64{}, YAxis: []int{}}
	avgOfConversationMessagesTrend := resp.TrendItemFloat{XAxis: []int64{}, YAxis: []float64{}}
	costConsumptionTrend := resp.TrendItemFloat{XAxis: []int64{}, YAxis: []float64{}}

	for day := 0; day < daysAgo; day++ {
		trendStartAt := endAtMidnight.AddDate(0, 0, -(daysAgo - day))
		trendEndAt := endAtMidnight.AddDate(0, 0, -(daysAgo - day - 1))

		// 过滤当天的消息
		var dayMessages []entity.Message
		for _, msg := range messages {
			msgTime := time.Unix(msg.Ctime/1000, 0)
			if msgTime.After(trendStartAt) && msgTime.Before(trendEndAt) {
				dayMessages = append(dayMessages, msg)
			}
		}

		// 计算当天指标
		totalMessagesCount := len(dayMessages)

		activeAccountsMap := make(map[uuid.UUID]bool)
		conversationMap := make(map[uuid.UUID]bool)
		var dayCostConsumption float64

		for _, msg := range dayMessages {
			activeAccountsMap[msg.CreatedBy] = true
			conversationMap[msg.ConversationID] = true
			dayCostConsumption += msg.TotalPrice
		}

		activeAccountsCount := len(activeAccountsMap)
		conversationCount := len(conversationMap)

		var avgOfConversationMessagesCount float64
		if conversationCount > 0 {
			avgOfConversationMessagesCount = float64(totalMessagesCount) / float64(conversationCount)
		}

		timestamp := trendStartAt.Unix()

		totalMessagesTrend.XAxis = append(totalMessagesTrend.XAxis, timestamp)
		totalMessagesTrend.YAxis = append(totalMessagesTrend.YAxis, totalMessagesCount)

		activeAccountsTrend.XAxis = append(activeAccountsTrend.XAxis, timestamp)
		activeAccountsTrend.YAxis = append(activeAccountsTrend.YAxis, activeAccountsCount)

		avgOfConversationMessagesTrend.XAxis = append(avgOfConversationMessagesTrend.XAxis, timestamp)
		avgOfConversationMessagesTrend.YAxis = append(avgOfConversationMessagesTrend.YAxis, avgOfConversationMessagesCount)

		costConsumptionTrend.XAxis = append(costConsumptionTrend.XAxis, timestamp)
		costConsumptionTrend.YAxis = append(costConsumptionTrend.YAxis, dayCostConsumption)
	}

	return resp.TrendData{
		TotalMessagesTrend:             totalMessagesTrend,
		ActiveAccountsTrend:            activeAccountsTrend,
		AvgOfConversationMessagesTrend: avgOfConversationMessagesTrend,
		CostConsumptionTrend:           costConsumptionTrend,
	}
}
