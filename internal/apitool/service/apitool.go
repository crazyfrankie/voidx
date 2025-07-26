package service

import (
	"context"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/apitool/repository"
	toolsentity "github.com/crazyfrankie/voidx/internal/core/tools/api_tools/entities"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/util"
)

type ApiToolService struct {
	repo *repository.ApiToolRepo
}

func NewApiToolService(repo *repository.ApiToolRepo) *ApiToolService {
	return &ApiToolService{repo: repo}
}

// API工具提供商相关方法

func (s *ApiToolService) GetApiToolProvidersWithPage(ctx context.Context, pageReq req.GetApiToolProvidersWithPageReq) (resp.GetApiToolProvidersWithPageResp, error) {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return resp.GetApiToolProvidersWithPageResp{}, err
	}

	_, err = s.repo.GetApiToolProvidersByAccountID(ctx, userID, pageReq)
	if err != nil {
		return resp.GetApiToolProvidersWithPageResp{}, err
	}

	return resp.GetApiToolProvidersWithPageResp{
		ID:          uuid.UUID{},
		Name:        "",
		Icon:        "",
		Description: "",
		Headers:     nil,
		Tools:       nil,
		Ctime:       0,
	}, nil
}

func (s *ApiToolService) GetApiToolProvider(ctx context.Context, providerID uuid.UUID) (*resp.ApiToolProviderResp, error) {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	provider, err := s.repo.GetApiToolProviderByID(ctx, providerID)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage("API工具提供商不存在")
	}

	if provider.AccountID != userID {
		return nil, errno.ErrForbidden.AppendBizMessage("无权限访问该API工具提供商")
	}

	return &resp.ApiToolProviderResp{
		ID:            provider.ID,
		Name:          provider.Name,
		Icon:          provider.Icon,
		Description:   provider.Description,
		OpenAPISchema: provider.OpenapiSchema,
		Headers:       provider.Headers,
		Ctime:         provider.Ctime,
	}, nil
}

func (s *ApiToolService) UpdateApiToolProvider(ctx context.Context, providerID uuid.UUID, updateReq req.UpdateApiToolProviderReq) error {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return err
	}

	provider, err := s.repo.GetApiToolProviderByID(ctx, providerID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage("API工具提供商不存在")
	}

	if provider.AccountID != userID {
		return errno.ErrForbidden.AppendBizMessage("无权限修改该API工具提供商")
	}

	// 构建更新数据
	updates := make(map[string]any)
	if updateReq.Name != "" {
		// 检查名称是否重复
		exists, err := s.repo.CheckProviderNameExistsExclude(ctx, userID, updateReq.Name, providerID)
		if err != nil {
			return err
		}
		if exists {
			return errno.ErrValidate.AppendBizMessage("提供商名称已存在")
		}
		updates["name"] = updateReq.Name
	}
	if updateReq.Icon != "" {
		updates["icon"] = updateReq.Icon
	}
	if updateReq.Headers != nil {
		updates["headers"] = updateReq.Headers
	}
	if updateReq.OpenAPISchema != "" {
		updates["openapi_schema"] = updateReq.OpenAPISchema
	}

	return s.repo.UpdateApiToolProvider(ctx, providerID, updates)
}

func (s *ApiToolService) DeleteApiToolProvider(ctx context.Context, providerID uuid.UUID) error {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return err
	}

	provider, err := s.repo.GetApiToolProviderByID(ctx, providerID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage("API工具提供商不存在")
	}

	if provider.AccountID != userID {
		return errno.ErrForbidden.AppendBizMessage("无权限删除该API工具提供商")
	}

	return s.repo.DeleteApiToolProvider(ctx, providerID)
}

// API工具相关方法

func (s *ApiToolService) CreateApiTool(ctx context.Context, createReq req.CreateApiToolReq) error {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return err
	}

	openapiSchema, err := s.parseOpenAPISchema(createReq.OpenAPISchema)
	if err != nil {
		return err
	}
	// 验证提供商权限
	provider, err := s.repo.GetApiToolProviderByName(ctx, userID, createReq.Name)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage("API工具提供商不存在")
	}

	if provider != nil {
		return errno.ErrValidate.AppendBizMessage("该API提供商已存在")
	}

	newProvider := &entity.ApiToolProvider{
		AccountID:     userID,
		Name:          createReq.Name,
		Icon:          createReq.Icon,
		OpenapiSchema: createReq.OpenAPISchema,
		Headers:       createReq.Headers,
		Description:   openapiSchema.Description,
	}

	if err := s.repo.CreateApiToolProvider(ctx, newProvider); err != nil {
		return err
	}

	for path, pathItem := range openapiSchema.Paths {
		if pathItem.Post != nil {
			item := pathItem.Post
			param, _ := sonic.Marshal(item.Parameters)
			tool := &entity.ApiTool{
				ProviderID:  newProvider.ID,
				AccountID:   userID,
				Name:        item.OperationID,
				Description: item.Description,
				Method:      "POST",
				URL:         openapiSchema.Server + path,
			}
			if err := sonic.Unmarshal(param, &tool.Parameters); err != nil {
				return err
			}

			err = s.repo.CreateApiTool(ctx, tool)
			if err != nil {
				return err
			}
		}
		if pathItem.Get != nil {
			item := pathItem.Get
			param, _ := sonic.Marshal(item.Parameters)
			tool := &entity.ApiTool{
				ProviderID:  newProvider.ID,
				AccountID:   userID,
				Name:        item.OperationID,
				Description: item.Description,
				Method:      "GET",
				URL:         openapiSchema.Server + path,
			}
			if err := sonic.Unmarshal(param, &tool.Parameters); err != nil {
				return err
			}

			err = s.repo.CreateApiTool(ctx, tool)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *ApiToolService) GetApiTool(ctx context.Context, toolID uuid.UUID) (*resp.ApiToolResp, error) {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	tool, err := s.repo.GetApiToolByID(ctx, toolID)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage("API工具不存在")
	}

	if tool.AccountID != userID {
		return nil, errno.ErrForbidden.AppendBizMessage("无权限访问该API工具")
	}

	provider, err := s.repo.GetApiToolProviderByID(ctx, tool.ProviderID)
	if err != nil {
		return nil, err
	}

	return &resp.ApiToolResp{
		ID:          tool.ID,
		Name:        tool.Name,
		Description: tool.Description,
		Inputs:      tool.Parameters,
		Provider: resp.ApiToolProvider{
			ID:          provider.ID,
			Name:        provider.Name,
			Icon:        provider.Icon,
			Description: provider.Description,
			Headers:     provider.Headers,
		},
	}, nil
}

func (s *ApiToolService) parseOpenAPISchema(openapiSchema string) (*toolsentity.OpenAPISchema, error) {
	// 1. 解析JSON字符串
	var data map[string]any
	if err := sonic.Unmarshal([]byte(strings.TrimSpace(openapiSchema)), &data); err != nil {
		return nil, &toolsentity.ValidateError{Message: "传递数据必须符合OpenAPI规范的JSON字符串"}
	}

	// 2. 提取基本字段
	server, _ := data["server"].(string)
	description, _ := data["description"].(string)
	paths, _ := data["paths"].(map[string]any)

	// 3. 创建并验证OpenAPISchema
	return toolsentity.NewOpenAPISchema(server, description, paths)
}
