package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"wxcloudrun-golang/conf"
	"wxcloudrun-golang/db"
	"wxcloudrun-golang/db/model"
	"wxcloudrun-golang/util"

	"github.com/cloudwego/hertz/pkg/app"
)

// WechatLoginResponse 微信登录接口的响应结构
type WechatLoginResponse struct {
	OpenID     string `json:"openid"` // 用户id，和 appID 一一对应。
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"` // 微信开放平台下用户全局id，跨小程序、网页、小游戏等。
	ErrCode    int    `json:"errcode"` // 不要改这里的 json tag, 微信就是这么返回的
	ErrMsg     string `json:"errmsg"`
}

type LoginRequest struct {
	Code string `json:"code"`
}

// WechatLogin
// 方法/路径：POST /api/user/login
// 鉴权：无（公开接口）
// 请求：JSON { code: string } 或 Query ?code=
// 返回：{ token: string, openid: string }
// 说明：调用微信 jscode2session 校验 code，返回 token
func WechatLogin(_ context.Context, c *app.RequestContext) {
	var req LoginRequest
	// Try to bind JSON body
	_ = c.Bind(&req)

	// If code is not in body, try query param
	if req.Code == "" {
		req.Code = c.Query("code")
	}

	if req.Code == "" {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing code"})
		return
	}

	// 构造请求微信服务器的 URL
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		conf.Conf.AppID, conf.Conf.AppSecret, req.Code)

	// 发送请求到微信服务器
	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to request Wechat server"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to read response body"})
		return
	}

	var wechatResp WechatLoginResponse
	err = json.Unmarshal(body, &wechatResp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse Wechat response"})
		return
	}

	if wechatResp.ErrCode != 0 {
		c.JSON(http.StatusBadRequest, map[string]string{"error": wechatResp.ErrMsg})
		return
	}

	// 准备用户数据
	userInfo := model.UserInfoDBModel{
		OpenID:       wechatResp.OpenID,
		UserNickName: "微信用户", // 默认值
		Role:         "member",
	}

	// Just ensure exist
	err = db.Get().Where(model.UserInfoDBModel{OpenID: wechatResp.OpenID}).FirstOrCreate(&userInfo).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
		return
	}

	// Generate JWT Token
	token, err := util.GenerateToken(wechatResp.OpenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
		return
	}

	// 构造响应
	response := map[string]string{
		"token":  token,
		"openid": wechatResp.OpenID,
	}

	c.JSON(http.StatusOK, response)
}
