package httputil

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Success(c *gin.Context, resp any) {
	c.JSON(http.StatusOK, Data{Code: 0, Message: "", Data: resp})
}
