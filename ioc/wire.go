//go:build wireinject

package ioc

import (
	"github.com/crazyfrankie/voidx/internal/account"
	"github.com/crazyfrankie/voidx/internal/ai"
	"github.com/crazyfrankie/voidx/internal/analysis"
	"github.com/crazyfrankie/voidx/internal/api_key"
	"github.com/crazyfrankie/voidx/internal/apitool"
	"github.com/crazyfrankie/voidx/internal/app"
	"github.com/crazyfrankie/voidx/internal/app_config"
	"github.com/crazyfrankie/voidx/internal/assistant_agent"
	"github.com/crazyfrankie/voidx/internal/audio"
	"github.com/crazyfrankie/voidx/internal/auth"
	"github.com/crazyfrankie/voidx/internal/builtin_app"
	"github.com/crazyfrankie/voidx/internal/builtin_tools"
	"github.com/crazyfrankie/voidx/internal/conversation"
	"github.com/crazyfrankie/voidx/internal/dataset"
	"github.com/crazyfrankie/voidx/internal/document"
	"github.com/crazyfrankie/voidx/internal/index"
	"github.com/crazyfrankie/voidx/internal/llm"
	"github.com/crazyfrankie/voidx/internal/oauth"
	"github.com/crazyfrankie/voidx/internal/openapi"
	"github.com/crazyfrankie/voidx/internal/platform"
	"github.com/crazyfrankie/voidx/internal/process_rule"
	"github.com/crazyfrankie/voidx/internal/retriever"
	"github.com/crazyfrankie/voidx/internal/segment"
	"github.com/crazyfrankie/voidx/internal/task"
	"github.com/crazyfrankie/voidx/internal/upload"
	"github.com/crazyfrankie/voidx/internal/vecstore"
	"github.com/crazyfrankie/voidx/internal/webapp"
	"github.com/crazyfrankie/voidx/internal/wechat"
	"github.com/crazyfrankie/voidx/internal/workflow"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var baseSet = wire.NewSet(InitCache, InitDB, InitJWT, InitVectorStore, InitMinIO, InitWechat)
var coreSet = wire.NewSet(InitBuiltinAppManager, InitBuiltinToolsCategories,
	InitEmbeddingService, InitFileExtractor, InitJiebaService, InitLLMCore, InitTokenBufMem, InitApiToolsManager, InitBuiltinToolsManager)

type Application struct {
	Server   *gin.Engine
	Consumer *task.TaskManager
}

func InitApplication() *Application {
	wire.Build(
		baseSet,
		coreSet,

		account.InitAccountModule,
		ai.InitAIModule,
		analysis.InitAnalysisModule,
		api_key.InitApiKeyModule,
		apitool.InitApiToolHandler,
		app.InitAppModule,
		app_config.InitAppConfigModule,
		assistant_agent.InitAssistantModule,
		audio.InitAudioModule,
		auth.InitAuthModule,
		builtin_app.InitBuiltinAppModule,
		builtin_tools.InitBuiltinToolsModule,
		conversation.InitConversationModule,
		dataset.InitDatasetHandler,
		document.InitDocumentModule,
		index.InitIndexModule,
		llm.InitLLMModule,
		oauth.InitOAuthModule,
		openapi.InitOpenAIModule,
		platform.InitPlatformModule,
		process_rule.InitProcessRuleModule,
		retriever.InitRetrieverModule,
		segment.InitSegmentModule,
		upload.InitUploadModule,
		vecstore.NewVecStoreService,
		webapp.InitWebAppModule,
		wechat.InitWechatModule,
		workflow.InitWorkflowModule,

		InitTask,
		InitMiddlewares,
		InitWeb,

		wire.FieldsOf(new(*account.AccountModule), "Handler"),
		wire.FieldsOf(new(*ai.AIModule), "Handler"),
		wire.FieldsOf(new(*analysis.AnalysisModule), "Handler"),
		wire.FieldsOf(new(*api_key.ApiKeyModule), "Handler"),
		wire.FieldsOf(new(*apitool.ApiToolModule), "Handler"),
		wire.FieldsOf(new(*app.AppModule), "Handler"),
		wire.FieldsOf(new(*app.AppModule), "Service"),
		wire.FieldsOf(new(*assistant_agent.AssistantModule), "Handler"),
		wire.FieldsOf(new(*audio.AudioModule), "Handler"),
		wire.FieldsOf(new(*auth.AuthModule), "Handler"),
		wire.FieldsOf(new(*builtin_app.BuiltinModule), "Handler"),
		wire.FieldsOf(new(*builtin_tools.BuiltinToolsModule), "Handler"),
		wire.FieldsOf(new(*conversation.ConversationModule), "Handler"),
		wire.FieldsOf(new(*dataset.DataSetModule), "Handler"),
		wire.FieldsOf(new(*document.DocumentModule), "Handler"),
		wire.FieldsOf(new(*index.IndexModule), "Service"),
		wire.FieldsOf(new(*llm.LLMModule), "Handler"),
		wire.FieldsOf(new(*oauth.OAuthModule), "Handler"),
		wire.FieldsOf(new(*openapi.OpenAPIModule), "Handler"),
		wire.FieldsOf(new(*platform.PlatformModule), "Handler"),
		wire.FieldsOf(new(*segment.SegmentModule), "Handler"),
		wire.FieldsOf(new(*upload.UploadModule), "Handler"),
		wire.FieldsOf(new(*upload.UploadModule), "Service"),
		wire.FieldsOf(new(*webapp.WebAppModule), "Handler"),
		wire.FieldsOf(new(*wechat.WechatModule), "Handler"),
		wire.FieldsOf(new(*workflow.WorkflowModule), "Handler"),

		wire.Struct(new(Application), "*"),
	)

	return new(Application)
}
