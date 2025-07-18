package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/util"
)

// PublishDraftAppConfig 发布应用的草稿配置
func (s *AppService) PublishDraftAppConfig(ctx context.Context, appID uuid.UUID) error {
	// 获取当前用户ID
	accountID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return err
	}

	// 获取应用
	app, err := s.repo.GetAppByID(ctx, appID)
	if err != nil {
		return err
	}

	// 检查权限
	if app.AccountID != accountID {
		return errno.ErrForbidden.AppendBizMessage("无权访问该应用")
	}

	// 检查草稿配置ID是否存在
	if app.DraftAppConfigID == nil {
		return errno.ErrNotFound.AppendBizMessage("应用草稿配置不存在")
	}

	// 获取草稿配置
	draftAppConfigVersion, err := s.repo.GetAppConfigVersion(ctx, *app.DraftAppConfigID)
	if err != nil {
		return err
	}

	// 创建运行时配置
	createdAppConfig, err := s.repo.CreateAppConfig(ctx, appID, draftAppConfigVersion.Config)
	if err != nil {
		return err
	}

	// 更新应用状态和配置ID
	app.AppConfigID = &createdAppConfig.ID
	app.Status = "published"
	if err := s.repo.UpdateApp(ctx, app); err != nil {
		return err
	}

	// 获取最大版本号
	maxVersion, err := s.repo.GetMaxPublishedVersion(ctx, appID)
	if err != nil {
		return err
	}

	// 创建发布历史记录
	_, err = s.repo.CreateAppConfigVersion(ctx, appID, maxVersion+1, "published", draftAppConfigVersion.Config)
	if err != nil {
		return err
	}

	return nil
}

// CancelPublishAppConfig 取消发布应用配置
func (s *AppService) CancelPublishAppConfig(ctx context.Context, appID uuid.UUID) error {
	// 获取当前用户ID
	accountID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return err
	}

	// 获取应用
	app, err := s.repo.GetAppByID(ctx, appID)
	if err != nil {
		return err
	}

	// 检查权限
	if app.AccountID != accountID {
		return errno.ErrForbidden.AppendBizMessage("无权访问该应用")
	}

	// 检查应用是否已发布
	if app.Status != "published" {
		return errno.ErrValidate.AppendBizMessage("应用未发布，无法取消发布")
	}

	// 更新应用状态
	app.Status = "draft"
	app.AppConfigID = nil
	if err := s.repo.UpdateApp(ctx, app); err != nil {
		return err
	}

	return nil
}

// FallbackHistoryToDraft 回退历史版本到草稿
func (s *AppService) FallbackHistoryToDraft(ctx context.Context, appID uuid.UUID, versionID uuid.UUID) error {
	// 获取当前用户ID
	accountID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return err
	}

	// 获取应用
	app, err := s.repo.GetAppByID(ctx, appID)
	if err != nil {
		return err
	}

	// 检查权限
	if app.AccountID != accountID {
		return errno.ErrForbidden.AppendBizMessage("无权访问该应用")
	}

	// 获取历史版本配置
	historyVersion, err := s.repo.GetAppConfigVersion(ctx, versionID)
	if err != nil {
		return err
	}

	// 检查历史版本是否属于该应用
	if historyVersion.AppID != appID {
		return errno.ErrValidate.AppendBizMessage("历史版本不属于该应用")
	}

	// 检查草稿配置ID是否存在
	if app.DraftAppConfigID == nil {
		return errno.ErrNotFound.AppendBizMessage("应用草稿配置不存在")
	}

	// 获取草稿配置
	draftAppConfigVersion, err := s.repo.GetAppConfigVersion(ctx, *app.DraftAppConfigID)
	if err != nil {
		return err
	}

	// 更新草稿配置
	draftAppConfigVersion.Config = historyVersion.Config
	if err := s.repo.UpdateAppConfigVersion(ctx, draftAppConfigVersion); err != nil {
		return err
	}

	return nil
}

// GetPublishHistoriesWithPage 获取发布历史分页列表
func (s *AppService) GetPublishHistoriesWithPage(ctx context.Context, appID uuid.UUID, pageReq req.GetPublishHistoriesWithPageReq) ([]*resp.AppConfigVersionResp, *resp.Paginator, error) {
	// 获取当前用户ID
	accountID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return nil, nil, err
	}

	// 获取应用
	app, err := s.repo.GetAppByID(ctx, appID)
	if err != nil {
		return nil, nil, err
	}

	// 检查权限
	if app.AccountID != accountID {
		return nil, nil, errno.ErrForbidden.AppendBizMessage("无权访问该应用")
	}

	// 获取发布历史分页列表
	histories, total, err := s.repo.GetPublishHistoriesWithPage(ctx, appID, pageReq.Page, pageReq.PageSize)
	if err != nil {
		return nil, nil, err
	}

	// 转换为响应
	historyResps := make([]*resp.AppConfigVersionResp, 0, len(histories))
	for _, history := range histories {
		historyResps = append(historyResps, &resp.AppConfigVersionResp{
			ID:         history.ID,
			AppID:      history.AppID,
			Version:    history.Version,
			ConfigType: history.ConfigType,
			Ctime:      history.Ctime,
			Utime:      history.Utime,
		})
	}

	// 构建分页器
	paginator := &resp.Paginator{
		CurrentPage: pageReq.Page,
		PageSize:    pageReq.PageSize,
		TotalPage:   int(total) / pageReq.PageSize,
		TotalRecord: int(total),
	}
	if int(total)%pageReq.PageSize > 0 {
		paginator.TotalPage++
	}

	return historyResps, paginator, nil
}
