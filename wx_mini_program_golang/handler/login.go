package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"wxcloudrun-golang/conf"
	"wxcloudrun-golang/db"
	"wxcloudrun-golang/db/model"
)

// WechatLoginResponse 微信登录接口的响应结构
type WechatLoginResponse struct {
	OpenID     string `json:"openid"` // 用户id，和 appID 一一对应。
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"` // 微信开放平台下用户全局id，跨小程序、网页、小游戏等。
	ErrCode    int    `json:"errcode"` // 不要改这里的 json tag, 微信就是这么返回的
	ErrMsg     string `json:"errmsg"`
}

// WechatLogin 处理微信小程序登录
func WechatLogin(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 获取 code 参数
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, `{"error": "Missing code"}`, http.StatusBadRequest)
		return
	}

	// 构造请求微信服务器的 URL
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		conf.Conf.AppID, conf.Conf.AppSecret, code)

	// 发送请求到微信服务器
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, `{"error": "Failed to request Wechat server"}`, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, `{"error": "Failed to read response body"}`, http.StatusInternalServerError)
		return
	}

	var wechatResp WechatLoginResponse
	err = json.Unmarshal(body, &wechatResp)
	if err != nil {
		http.Error(w, `{"error": "Failed to parse Wechat response"}`, http.StatusInternalServerError)
		return
	}

	if wechatResp.ErrCode != 0 {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, wechatResp.ErrMsg), http.StatusBadRequest)
		return
	}

	// 自动注册用户
	userInfo := model.UserInfoDBModel{
		OpenID:       wechatResp.OpenID,
		UserNickName: "微信用户", // 默认昵称
		Role:         "member",
	}

	// 如果不存在则创建
	if err := db.Get().Where(model.UserInfoDBModel{OpenID: wechatResp.OpenID}).FirstOrCreate(&userInfo).Error; err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	// 构造响应
	response := map[string]string{
		"openid": wechatResp.OpenID,
		//"session_key": wechatResp.SessionKey,  开发者不应该把 session_key 传到小程序客户端等服务器外的环境
	}

	// 将响应转换为 JSON 并写入响应
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, `{"error": "Failed to generate response"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
