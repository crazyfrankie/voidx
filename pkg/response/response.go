package response

import (
	"net/http"

	"github.com/crazyfrankie/gem/gerrors"
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/voidx/pkg/errno"
)

type Response struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func Error(c *gin.Context, err error) {
	if bizErr, ok := gerrors.FromBizStatusError(err); ok {
		c.JSON(http.StatusOK, Response{
			Code:    bizErr.BizStatusCode(),
			Message: bizErr.BizMessage(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    50000,
		Message: err.Error(),
	})
}

func Success(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    errno.Success.BizStatusCode(),
		Message: errno.Success.BizMessage(),
	})
}

func SuccessWithData(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    errno.Success.BizStatusCode(),
		Message: errno.Success.BizMessage(),
		Data:    data,
	})
}
