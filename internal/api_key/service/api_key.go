package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/api_key/repository"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/pkg/errno"
)

type ApiKeyService struct {
	repo *repository.ApiKeyRepo
}

func NewApiKeyService(repo *repository.ApiKeyRepo) *ApiKeyService {
	return &ApiKeyService{repo: repo}
}

// CreateApiKey 根据传递的信息创建API秘钥
func (s *ApiKeyService) CreateApiKey(ctx context.Context, userID uuid.UUID, createReq req.CreateApiKeyReq) error {
	apiKey := &entity.ApiKey{
		AccountID: userID,
		ApiKey:    s.generateApiKey(),
		IsActive:  createReq.IsActive,
		Remark:    createReq.Remark,
	}

	return s.repo.CreateApiKey(ctx, apiKey)
}

// GetApiKey 根据传递的秘钥id+账号信息获取记录
func (s *ApiKeyService) GetApiKey(ctx context.Context, apiKeyID, userID uuid.UUID) (*entity.ApiKey, error) {
	apiKey, err := s.repo.GetApiKeyByID(ctx, apiKeyID)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage("API秘钥不存在")
	}

	if apiKey.AccountID != userID {
		return nil, errno.ErrForbidden.AppendBizMessage("API秘钥不存在或无权限")
	}

	return apiKey, nil
}

// GetApiKeyByCredential 根据传递的凭证信息获取ApiKey记录
func (s *ApiKeyService) GetApiKeyByCredential(ctx context.Context, apiKey string) (*entity.ApiKey, error) {
	return s.repo.GetApiKeyByCredential(ctx, apiKey)
}

// UpdateApiKey 根据传递的信息更新API秘钥
func (s *ApiKeyService) UpdateApiKey(ctx context.Context, apiKeyID, userID uuid.UUID, updateReq req.UpdateApiKeyReq) error {
	apiKey, err := s.GetApiKey(ctx, apiKeyID, userID)
	if err != nil {
		return err
	}

	updates := map[string]any{
		"is_active": updateReq.IsActive,
		"remark":    updateReq.Remark,
	}

	return s.repo.UpdateApiKey(ctx, apiKey.ID, updates)
}

// UpdateApiKeyIsActive 根据传递的信息更新API秘钥激活状态
func (s *ApiKeyService) UpdateApiKeyIsActive(ctx context.Context, apiKeyID, userID uuid.UUID, isActive bool) error {
	apiKey, err := s.GetApiKey(ctx, apiKeyID, userID)
	if err != nil {
		return err
	}

	updates := map[string]any{
		"is_active": isActive,
	}

	return s.repo.UpdateApiKey(ctx, apiKey.ID, updates)
}

// DeleteApiKey 根据传递的id删除API秘钥
func (s *ApiKeyService) DeleteApiKey(ctx context.Context, apiKeyID, userID uuid.UUID) error {
	apiKey, err := s.GetApiKey(ctx, apiKeyID, userID)
	if err != nil {
		return err
	}

	return s.repo.DeleteApiKey(ctx, apiKey.ID)
}

// GetApiKeysWithPage 根据传递的信息获取API秘钥分页列表数据
func (s *ApiKeyService) GetApiKeysWithPage(ctx context.Context, userID uuid.UUID, pageReq req.GetApiKeysWithPageReq) ([]resp.GetApiKeysWithPageResp, resp.Paginator, error) {
	apiKeys, total, err := s.repo.GetApiKeysByAccountID(ctx, userID, pageReq)
	if err != nil {
		return nil, resp.Paginator{}, err
	}

	// 转换为响应格式
	apiKeyResps := make([]resp.GetApiKeysWithPageResp, len(apiKeys))
	for i, apiKey := range apiKeys {
		apiKeyResps[i] = resp.GetApiKeysWithPageResp{
			ID:       apiKey.ID,
			ApiKey:   apiKey.ApiKey,
			IsActive: apiKey.IsActive,
			Remark:   apiKey.Remark,
			Utime:    apiKey.Utime,
			Ctime:    apiKey.Ctime,
		}
	}

	// 计算分页信息
	totalPages := (int(total) + pageReq.PageSize - 1) / pageReq.PageSize
	paginator := resp.Paginator{
		CurrentPage: pageReq.Page,
		PageSize:    pageReq.PageSize,
		TotalPage:   totalPages,
		TotalRecord: int(total),
	}

	return apiKeyResps, paginator, nil
}

// generateApiKey 生成一个长度为48的API秘钥，并携带前缀
func (s *ApiKeyService) generateApiKey(apiKeyPrefix ...string) string {
	prefix := "llmops-v1/"
	if len(apiKeyPrefix) > 0 {
		prefix = apiKeyPrefix[0]
	}

	// 生成48字节的随机数据
	randomBytes := make([]byte, 48)
	rand.Read(randomBytes)

	// 使用base64编码，但替换URL不安全字符
	token := base64.URLEncoding.EncodeToString(randomBytes)

	return prefix + token
}
