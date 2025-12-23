package util

import (
	"encoding/json"
	"fmt"
	"github.com/bytedance/sonic"
	"math/rand"
	"net/http"
	"strings"
)

// JsonResult 返回结构
type JsonResult struct {
	Code     int         `json:"code"`
	ErrorMsg string      `json:"errorMsg,omitempty"`
	Data     interface{} `json:"data"`
}

func BindJson(r *http.Request, ptr interface{}) error {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	return decoder.Decode(ptr)
}

// GetOpenIdFromHeader 从Header里面获取微信 openId
func GetOpenIdFromHeader(r *http.Request) (string, error) {
	openId := r.Header.Get("X-Wx-Openid")
	if openId == "" {
		return "", fmt.Errorf("缺少openId")
	}
	return openId, nil
}

// ReturnSuccessJSON 向resp中注入json返回值
func ReturnSuccessJSON(w http.ResponseWriter, res interface{}) {
	w.Header().Set("content-type", "application/json")
	marshal, _ := json.Marshal(JsonResult{Data: res})
	_, _ = w.Write(marshal)
}

// ReturnErrorJSON 向 resp 中注入 errorMsg 返回值
func ReturnErrorJSON(w http.ResponseWriter, msg string, err error) {
	w.Header().Set("content-type", "application/json")
	marshal, _ := json.Marshal(JsonResult{Code: -1, ErrorMsg: fmt.Sprintf("msg %s, error = %v", msg, err)})
	_, _ = w.Write(marshal)
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
