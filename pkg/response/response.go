package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 通用 API 响应结构
type Response struct {
	Code    int         `json:"code" example:"0"`           // 业务状态码，0 表示成功
	Message string      `json:"message" example:"success"`  // 响应消息
	Data    interface{} `json:"data,omitempty"`             // 响应数据
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func Error(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}
