package service

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/internal/platform/repository"
	"github.com/crazyfrankie/voidx/pkg/consts"
	"github.com/crazyfrankie/voidx/pkg/errno"
)

type PlatformService struct {
	repo *repository.PlatformRepo
}

func NewPlatformService(repo *repository.PlatformRepo) *PlatformService {
	return &PlatformService{repo: repo}
}

// GetWechatConfig 根据传递的应用id+账号获取微信发布配置
func (s *PlatformService) GetWechatConfig(ctx context.Context, appID, userID uuid.UUID) (*resp.GetWechatConfigResp, error) {
	// 1. 获取应用信息并校验权限
	app, err := s.repo.GetAppByID(ctx, appID)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("应用不存在"))
	}

	if app.AccountID != userID {
		return nil, errno.ErrForbidden.AppendBizMessage(errors.New("无权限访问该应用"))
	}

	// 2. 获取或创建微信配置
	wechatConfig, err := s.getOrCreateWechatConfig(ctx, appID)
	if err != nil {
		return nil, err
	}

	// 3. 构建响应
	serviceAPIPrefix := os.Getenv("SERVICE_API_PREFIX")
	if serviceAPIPrefix == "" {
		serviceAPIPrefix = "http://localhost:8080/api/v1"
	}

	serviceIP := os.Getenv("SERVICE_IP")
	if serviceIP == "" {
		serviceIP = "127.0.0.1"
	}

	return &resp.GetWechatConfigResp{
		AppID:           wechatConfig.AppID,
		URL:             fmt.Sprintf("%s/wechat/%s", serviceAPIPrefix, appID.String()),
		IP:              serviceIP,
		WechatAppID:     wechatConfig.WechatAppID,
		WechatAppSecret: wechatConfig.WechatAppSecret,
		WechatToken:     wechatConfig.WechatToken,
		Status:          string(wechatConfig.Status),
		Utime:           wechatConfig.Utime,
		Ctime:           wechatConfig.Ctime,
	}, nil
}

// UpdateWechatConfig 根据传递的应用id+账号+配置信息更新应用的微信发布配置
func (s *PlatformService) UpdateWechatConfig(ctx context.Context, appID, userID uuid.UUID, updateReq req.UpdateWechatConfigReq) error {
	// 1. 获取应用信息并校验权限
	app, err := s.repo.GetAppByID(ctx, appID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("应用不存在"))
	}

	if app.AccountID != userID {
		return errno.ErrForbidden.AppendBizMessage(errors.New("无权限访问该应用"))
	}

	// 2. 获取或创建微信配置
	wechatConfig, err := s.getOrCreateWechatConfig(ctx, appID)
	if err != nil {
		return err
	}

	// 3. 根据传递的请求判断app_id/app_secret/token是否齐全并计算状态
	status := consts.WechatConfigStatusUnconfigured
	if updateReq.WechatAppID != "" && updateReq.WechatAppSecret != "" && updateReq.WechatToken != "" {
		status = consts.WechatConfigStatusConfigured
	}

	// 4. 根据应用的发布状态修正状态数据
	if app.Status == consts.AppStatusDraft && status == consts.WechatConfigStatusConfigured {
		status = consts.WechatConfigStatusUnconfigured
	}

	// 5. 更新微信配置
	updates := map[string]any{
		"wechat_app_id":     updateReq.WechatAppID,
		"wechat_app_secret": updateReq.WechatAppSecret,
		"wechat_token":      updateReq.WechatToken,
		"status":            status,
	}

	return s.repo.UpdateWechatConfig(ctx, wechatConfig.ID, updates)
}

// getOrCreateWechatConfig 获取或创建微信配置
func (s *PlatformService) getOrCreateWechatConfig(ctx context.Context, appID uuid.UUID) (*entity.WechatConfig, error) {
	// 尝试获取现有配置
	wechatConfig, err := s.repo.GetWechatConfigByAppID(ctx, appID)
	if err == nil && wechatConfig != nil {
		return wechatConfig, nil
	}

	// 如果不存在，创建新的配置
	newConfig := &entity.WechatConfig{
		AppID:           appID,
		WechatAppID:     "",
		WechatAppSecret: "",
		WechatToken:     "",
		Status:          consts.WechatConfigStatusUnconfigured,
	}

	err = s.repo.CreateWechatConfig(ctx, newConfig)
	if err != nil {
		return nil, err
	}

	return newConfig, nil
}
