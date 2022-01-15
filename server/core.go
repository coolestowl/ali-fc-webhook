package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrFuncHandler(ctx *gin.Context, f func(*gin.Context) (interface{}, error)) {
	resp, err := f(ctx)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": resp,
	})
}
