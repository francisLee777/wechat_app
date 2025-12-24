package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
)

type JsonResult struct {
	Code     int         `json:"code"`
	ErrorMsg string      `json:"errorMsg,omitempty"`
	Data     interface{} `json:"data"`
}

// SendResponse 发送成功的 JSON 响应
func SendResponse(c *app.RequestContext, data interface{}) {
	c.JSON(http.StatusOK, JsonResult{
		Code: 0,
		Data: data,
	})
}

// SendError 发送错误的 JSON 响应
func SendError(c *app.RequestContext, code int, msg string) {
	c.JSON(http.StatusOK, JsonResult{
		Code:     code,
		ErrorMsg: msg,
	})
}

// HandlerFunc 定义适配 Hertz 的 Handler 函数签名
type HandlerFunc func(ctx context.Context, c *app.RequestContext)
