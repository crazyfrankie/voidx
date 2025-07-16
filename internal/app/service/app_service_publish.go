package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/pkg/errno"
)

// PublishDraftAppConfig 发布/更新特定的草稿配置信息
func (s *AppService) PublishDraftAppConfig(ctx context.Context, appID uuid.UUID) error {
	// 获取当前用户ID
	accountID, err := getCurrentUserID(ctx)
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

	// 创建应用配置
	appConfig, err := s.repo.CreateAppConfig(ctx, appID, draftAppConfigVersion.Config)
	if err != nil {
		return err
	}

	// 更新应用状态
	app.AppConfigID = &appConfig.ID
	app.Status = "published"
	if err := s.repo.UpdateApp(ctx, app); err != nil {
		return err
	}

	// 删除原有的知识库关联记录
	if err := s.repo.DeleteAppDatasetJoins(ctx, appID); err != nil {
		return err
	}

	// 添加新的知识库关联记录
	datasets, ok := draftAppConfigVersion.Config["datasets"].([]interface{})
	if ok {
		for _, dataset := range datasets {
			datasetMap, ok := dataset.(map[string]interface{})
			if !ok {
				continue
			}

			datasetIDStr, ok := datasetMap["id"].(string)
			if !ok {
				continue
			}

			datasetID, err := uuid.Parse(datasetIDStr)
			if err != nil {
				continue
			}

			if err := s.repo.CreateAppDatasetJoin(ctx, appID, datasetID); err != nil {
				return err
			}
		}
	}

	// 获取当前最大的发布版本
	maxVersion, err := s.repo.GetMaxPublishedVersion(ctx, appID)
	if err != nil {
		return err
	}

	// 创建发布历史配置
	_, err = s.repo.CreateAppConfigVersion(ctx, appID, maxVersion+1, "published", draftAppConfigVersion.Config)
	if err != nil {
		return err
	}

	return nil
}

// CancelPublishAppConfig 取消发布指定的应用配置信息
func (s *AppService) CancelPublishAppConfig(ctx context.Context, appID uuid.UUID) error {
	// 获取当前用户ID
	accountID, err := getCurrentUserID(ctx)
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

	// 检查应用状态
	if app.Status != "published" {
		return errno.ErrValidate.AppendBizMessage("当前应用未发布，无法取消发布")
	}

	// 更新应用状态
	app.AppConfigID = nil
	app.Status = "draft"
	if err := s.repo.UpdateApp(ctx, app); err != nil {
		return err
	}

	// 删除知识库关联记录
	if err := s.repo.DeleteAppDatasetJoins(ctx, appID); err != nil {
		return err
	}

	return nil
}

// GetPublishHistoriesWithPage 获取应用发布历史列表
func (s *AppService) GetPublishHistoriesWithPage(ctx context.Context, appID uuid.UUID, pageReq req.GetPublishHistoriesWithPageReq) ([]*resp.AppConfigVersionResp, *resp.Paginator, error) {
	// 获取当前用户ID
	accountID, err := getCurrentUserID(ctx)
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
	appConfigVersions, total, err := s.repo.GetPublishHistoriesWithPage(ctx, appID, pageReq.Page, pageReq.PageSize)
	if err != nil {
		return nil, nil, err
	}

	// 转换为响应
	appConfigVersionResps := make([]*resp.AppConfigVersionResp, 0, len(appConfigVersions))
	for _, appConfigVersion := range appConfigVersions {
		appConfigVersionResps = append(appConfigVersionResps, &resp.AppConfigVersionResp{
			ID:         appConfigVersion.ID,
			AppID:      appConfigVersion.AppID,
			Version:    appConfigVersion.Version,
			ConfigType: appConfigVersion.ConfigType,
			Ctime:      appConfigVersion.Ctime,
			Utime:      appConfigVersion.Utime,
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

	return appConfigVersionResps, paginator, nil
}

// FallbackHistoryToDraft 退回指定版本到草稿中
func (s *AppService) FallbackHistoryToDraft(ctx context.Context, appID uuid.UUID, appConfigVersionID uuid.UUID) error {
	// 获取当前用户ID
	accountID, err := getCurrentUserID(ctx)
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

	// 获取历史配置版本
	historyAppConfigVersion, err := s.repo.GetAppConfigVersion(ctx, appConfigVersionID)
	if err != nil {
		return err
	}

	// 检查历史配置版本是否属于该应用
	if historyAppConfigVersion.AppID != appID {
		return errno.ErrForbidden.AppendBizMessage("历史配置版本不属于该应用")
	}

	// 校验历史版本配置
	if err := validateDraftAppConfig(historyAppConfigVersion.Config, accountID); err != nil {
		return err
	}

	// 获取草稿配置
	draftAppConfigVersion, err := s.repo.GetAppConfigVersion(ctx, *app.DraftAppConfigID)
	if err != nil {
		return err
	}

	// 更新草稿配置
	draftAppConfigVersion.Config = historyAppConfigVersion.Config
	if err := s.repo.UpdateAppConfigVersion(ctx, draftAppConfigVersion); err != nil {
		return err
	}

	return nil
}

// GetPublishedConfig 获取应用的发布配置信息
func (s *AppService) GetPublishedConfig(ctx context.Context, appID uuid.UUID) (*resp.PublishedConfigResp, error) {
	// 获取当前用户ID
	accountID, err := getCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	// 获取应用
	app, err := s.repo.GetAppByID(ctx, appID)
	if err != nil {
		return nil, err
	}

	// 检查权限
	if app.AccountID != accountID {
		return nil, errno.ErrForbidden.AppendBizMessage("无权访问该应用")
	}

	// 构建响应
	resp := &resp.PublishedConfigResp{}
	resp.WebApp.Status = app.Status
	resp.WebApp.Token = app.Token
	if app.Token == "" {
		resp.WebApp.Token = generateRandomString(16)
	}

	return resp, nil
}

// RegenerateWebAppToken 重新生成WebApp凭证标识
func (s *AppService) RegenerateWebAppToken(ctx context.Context, appID uuid.UUID) (string, error) {
	// 获取当前用户ID
	accountID, err := getCurrentUserID(ctx)
	if err != nil {
		return "", err
	}

	// 获取应用
	app, err := s.repo.GetAppByID(ctx, appID)
	if err != nil {
		return "", err
	}

	// 检查权限
	if app.AccountID != accountID {
		return "", errno.ErrForbidden.AppendBizMessage("无权访问该应用")
	}

	// 检查应用状态
	if app.Status != "published" {
		return "", errno.ErrValidate.AppendBizMessage("应用未发布，无法生成WebApp凭证标识")
	}

	// 生成新的token
	token := generateRandomString(16)
	app.Token = token
	if err := s.repo.UpdateApp(ctx, app); err != nil {
		return "", err
	}

	return token, nil
}
