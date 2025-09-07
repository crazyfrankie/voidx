package httputil

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	//"github.com/crazyfrankie/voidx/pkg/errorx"
	"github.com/crazyfrankie/gem/gerrors"
	"github.com/crazyfrankie/voidx/pkg/logs"
	"github.com/crazyfrankie/voidx/types/errno"
)

type Data struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func BadRequest(c *gin.Context, errMsg string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, Data{Code: http.StatusBadRequest, Message: errMsg})
}

func InternalError(c *gin.Context, err error) {
	var customErr gerrors.BizErrorIface

	ctx := c.Request.Context()

	if errors.As(err, &customErr) && customErr.BizStatusCode() != 0 {
		if customErr.BizStatusCode() == errno.ErrUnauthorized.BizStatusCode() {
			logs.Infof("user is not login, %v \n", customErr.BizMessage())
		} else {
			logs.CtxWarnf(ctx, "[ErrorX] error:  %v %v \n", customErr.BizStatusCode(), err)
		}
		c.AbortWithStatusJSON(http.StatusOK, Data{Code: customErr.BizStatusCode(), Message: customErr.BizMessage()})
		return
	}

	logs.CtxErrorf(ctx, "[InternalError]  error: %v \n", err)
	c.AbortWithStatusJSON(http.StatusInternalServerError, Data{Code: 500, Message: "internal server error"})
}
