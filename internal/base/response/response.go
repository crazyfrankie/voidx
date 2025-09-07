package response

import (
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/voidx/internal/base/internal/httputil"
)

func InvalidParamRequestResponse(c *gin.Context, err error) {
	httputil.BadRequest(c, err.Error())
}

func InternalServerErrorResponse(c *gin.Context, err error) {
	httputil.InternalError(c, err)
}

func Success(c *gin.Context) {
	httputil.Success(c, nil)
}

func Data(c *gin.Context, data any) {
	httputil.Success(c, data)
}
