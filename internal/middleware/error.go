package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/NanoBoom/asethub/internal/errors"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		switch e := err.(type) {
		case *errors.AppError:
			c.JSON(e.Code, gin.H{
				"error": e.Message,
				"code":  e.Code,
			})
		default:
			// 临时添加详细错误日志
			println("ERROR:", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"code":  500,
				"debug": err.Error(), // 临时添加调试信息
			})
		}
	}
}
