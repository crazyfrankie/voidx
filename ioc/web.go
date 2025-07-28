package ioc

import (
	"github.com/crazyfrankie/voidx/internal/ai"
	"github.com/crazyfrankie/voidx/internal/analysis"
	"github.com/crazyfrankie/voidx/internal/api_key"
	"github.com/crazyfrankie/voidx/internal/apitool"
	"github.com/crazyfrankie/voidx/internal/assistant_agent"
	"github.com/crazyfrankie/voidx/internal/audio"
	"github.com/crazyfrankie/voidx/internal/builtin_app"
	"github.com/crazyfrankie/voidx/internal/builtin_tools"
	"github.com/crazyfrankie/voidx/internal/conversation"
	"github.com/crazyfrankie/voidx/internal/dataset"
	"github.com/crazyfrankie/voidx/internal/document"
	"github.com/crazyfrankie/voidx/internal/oauth"
	"github.com/crazyfrankie/voidx/internal/openapi"
	"github.com/crazyfrankie/voidx/internal/platform"
	"github.com/crazyfrankie/voidx/internal/segment"
	"github.com/crazyfrankie/voidx/internal/upload"
	"github.com/crazyfrankie/voidx/internal/webapp"
	"github.com/crazyfrankie/voidx/internal/wechat"
	"github.com/crazyfrankie/voidx/internal/workflow"
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/voidx/internal/account"
	"github.com/crazyfrankie/voidx/internal/app"
	"github.com/crazyfrankie/voidx/internal/auth"
	"github.com/crazyfrankie/voidx/internal/llm"
	"github.com/crazyfrankie/voidx/pkg/jwt"
	"github.com/crazyfrankie/voidx/pkg/middlewares"
)

func InitWeb(mws []gin.HandlerFunc, account *account.Handler, ai *ai.Handler, analysis *analysis.Handler,
	apiKey *api_key.Handler, apiTool *apitool.Handler, app *app.Handler,
	assistant *assistant_agent.Handler, audio *audio.Handler, auth *auth.Handler,
	builtinApp *builtin_app.Handler, builtinTools *builtin_tools.Handler,
	conversation *conversation.Handler, dataset *dataset.Handler,
	document *document.Handler, llm *llm.Handler, oauth *oauth.Handler,
	openapi *openapi.Handler, platform *platform.Handler, segment *segment.Handler,
	upload *upload.Handler, webapp *webapp.Handler, wechat *wechat.Handler, workflow *workflow.Handler) *gin.Engine {
	srv := gin.Default()
	srv.Use(mws...)

	apiGroup := srv.Group("api")

	account.RegisterRoute(apiGroup)
	ai.RegisterRoute(apiGroup)
	analysis.RegisterRoute(apiGroup)
	apiKey.RegisterRoute(apiGroup)
	apiTool.RegisterRoute(apiGroup)
	app.RegisterRoute(apiGroup)
	assistant.RegisterRoute(apiGroup)
	audio.RegisterRoute(apiGroup)
	auth.RegisterRoute(apiGroup)
	builtinApp.RegisterRoute(apiGroup)
	builtinTools.RegisterRoute(apiGroup)
	conversation.RegisterRoute(apiGroup)
	dataset.RegisterRoute(apiGroup)
	document.RegisterRoute(apiGroup)
	llm.RegisterRoute(apiGroup)
	oauth.RegisterRoute(apiGroup)
	openapi.RegisterRoute(apiGroup)
	platform.RegisterRoute(apiGroup)
	segment.RegisterRoute(apiGroup)
	upload.RegisterRoute(apiGroup)
	webapp.RegisterRoute(apiGroup)
	wechat.RegisterRoute(apiGroup)
	workflow.RegisterRoute(apiGroup)

	return srv
}

func InitMiddlewares(jwt *jwt.TokenService) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middlewares.CORS(),

		middlewares.NewAuthnHandler(jwt).
			IgnorePath("/api/auth/login").
			Auth(),
	}
}
