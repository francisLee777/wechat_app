package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"wxcloudrun-golang/db"
	"wxcloudrun-golang/db/model"
	"wxcloudrun-golang/util"

	"github.com/cloudwego/hertz/pkg/app"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetUserInfo
// 方法/路径：GET /api/user/getUserInfo
// 鉴权：需 JWT
// 返回：UserInfoDBModel（包含 openId、user_nickName、user_icon_url、family_id、role 等）
func GetUserInfo(_ context.Context, c *app.RequestContext) {
	openId, err := util.GetOpenId(c)
	if err != nil {
		c.String(http.StatusOK, fmt.Sprintf("没有登录%v", err))
		return
	}
	q1 := db.DB.UserInfoDBModel
	userInfo, err := q1.Where(q1.OpenID.Eq(openId)).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.String(http.StatusOK, fmt.Sprintf("数据库错误%v", err))
		return
	}
	// 查不到时 userInfo 为空
	util.ReturnSuccessJSON(c, userInfo)
}

// SaveNickName
// 方法/路径：POST /api/user/saveNickName
// 鉴权：需 JWT
// 请求：{ nickName: string }
// 返回：{ nickName: string }
func SaveNickName(_ context.Context, c *app.RequestContext) {
	// 获取用户 OpenID
	openId, err := util.GetOpenId(c)
	if err != nil {
		util.ReturnErrorJSON(c, "未登录", err)
		return
	}

	// 解析请求体
	var req struct {
		NickName string `json:"nickName"`
	}
	if err = util.BindJson(c, &req); err != nil {
		util.ReturnErrorJSON(c, "请求参数解析失败", err)
		return
	}

	// 参数验证
	if req.NickName == "" {
		util.ReturnErrorJSON(c, "昵称不能为空", nil)
		return
	}

	// 数据库操作
	q1 := db.DB.UserInfoDBModel
	userInfo := &model.UserInfoDBModel{
		OpenID:       openId,
		UserNickName: req.NickName,
	}

	err = q1.Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{q1.UserNickName.ColumnName().String()}),
	}).Create(userInfo)

	if err != nil {
		util.ReturnErrorJSON(c, "保存昵称失败", err)
		return
	}

	// 同步更新其他模块
	SyncUserProfile(openId, SyncTypeNickName, req.NickName)

	util.ReturnSuccessJSON(c, map[string]interface{}{
		"nickName": req.NickName,
	})
}

// SaveIconURL
// 方法/路径：POST /api/user/saveIconURL
// 鉴权：需 JWT
// 请求：Query 参数 iconURL
// 返回：string（保存的头像 URL）
func SaveIconURL(_ context.Context, c *app.RequestContext) {
	openId, err := util.GetOpenId(c)
	if err != nil {
		c.String(http.StatusOK, fmt.Sprintf("没有登录%v", err))
		return
	}
	iconURL := c.Query("iconURL")
	if iconURL == "" {
		c.String(http.StatusOK, "入参iconURL缺失")
		return
	}

	q1 := db.DB.UserInfoDBModel
	if err = q1.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{q1.UserIconURL.ColumnName().String()})}).
		Create(&model.UserInfoDBModel{
			OpenID:      openId,
			UserIconURL: iconURL,
		}); err != nil {
		c.String(http.StatusOK, fmt.Sprintf("数据库错误%v", err))
		return
	}

	// 同步更新其他模块
	SyncUserProfile(openId, SyncTypeAvatar, iconURL)

	util.ReturnSuccessJSON(c, iconURL)
}
