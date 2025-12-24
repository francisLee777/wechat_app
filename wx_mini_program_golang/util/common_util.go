package util

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/app"
)

// JsonResult 返回结构
type JsonResult struct {
	Code     int         `json:"code"`
	ErrorMsg string      `json:"errorMsg,omitempty"`
	Data     interface{} `json:"data"`
}

func BindJson(c *app.RequestContext, ptr interface{}) error {
	return c.Bind(ptr)
}

// GetOpenId 获取用户 OpenID (仅从 JWT 中间件设置的 Context 中获取)
func GetOpenId(c *app.RequestContext) (string, error) {
	// 从 Context 获取 OpenID (由 JWT 中间件设置)
	if value, exists := c.Get("openid"); exists {
		if openId, ok := value.(string); ok {
			return openId, nil
		}
	}
	// 不再支持从 Header 获取 OpenID，仅使用 JWT
	return "", fmt.Errorf("未登录或缺少有效凭证")
}

// ReturnSuccessJSON 向resp中注入json返回值
func ReturnSuccessJSON(c *app.RequestContext, res interface{}) {
	c.JSON(http.StatusOK, JsonResult{Data: res})
}

// ReturnErrorJSON 向 resp 中注入 errorMsg 返回值
func ReturnErrorJSON(c *app.RequestContext, msg string, err error) {
	c.PureJSON(http.StatusInternalServerError, JsonResult{Code: -1, ErrorMsg: fmt.Sprintf("msg %s, error = %v", msg, err)})
}

// GenerateUUID 生成 UUID
func GenerateUUID() string {
	uuid := rand.Intn(1000000000)
	uuidStr := fmt.Sprintf("%d", uuid)
	// 将 UUID 转换为字符串，并去掉其中的空格和连字符
	return strings.ReplaceAll(uuidStr, "-", "")
}

func Convert2JSONString(int interface{}) string {
	str, _ := sonic.MarshalString(int)
	return str
}
