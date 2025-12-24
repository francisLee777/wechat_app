package handler

import (
	"context"
	"sort"
	"wxcloudrun-golang/db"
	"wxcloudrun-golang/db/dao"
	"wxcloudrun-golang/db/model"
	"wxcloudrun-golang/util"

	"github.com/cloudwego/hertz/pkg/app"
	"gorm.io/gorm"
)

type AddMessageRequest struct {
	Content         string `json:"content"`
	ParentID        *int64 `json:"parentId"`
	ReplyToUserID   string `json:"replyToUserId"`
	ReplyToUserName string `json:"replyToUserName"`
}

// AddMessage
// 方法/路径：POST /api/message/add
// 鉴权：需 JWT；用户必须已加入家庭
// 请求：{ content: string, parentId?: number, replyToUserId?: string, replyToUserName?: string }
// 返回：MessageDBModel
func AddMessage(_ context.Context, c *app.RequestContext) {
	openId, err := util.GetOpenId(c)
	if err != nil {
		util.ReturnErrorJSON(c, "未登录", err)
		return
	}

	var req AddMessageRequest
	if err := util.BindJson(c, &req); err != nil {
		util.ReturnErrorJSON(c, "参数错误", err)
		return
	}

	var user model.UserInfoDBModel
	if err := db.Get().Where("openId = ?", openId).First(&user).Error; err != nil {
		util.ReturnErrorJSON(c, "用户不存在", err)
		return
	}

	if user.FamilyID == 0 {
		util.ReturnErrorJSON(c, "请先加入家庭", nil)
		return
	}

	msg := model.MessageDBModel{
		FamilyID:        user.FamilyID,
		UserID:          user.OpenID,
		UserName:        user.UserNickName,
		UserAvatar:      user.UserIconURL,
		Content:         req.Content,
		ReplyToUserID:   req.ReplyToUserID,
		ReplyToUserName: req.ReplyToUserName,
	}
	if req.ParentID != nil {
		msg.ParentID = *req.ParentID
	}

	if err := db.Get().Create(&msg).Error; err != nil {
		util.ReturnErrorJSON(c, "发送失败", err)
		return
	}

	util.ReturnSuccessJSON(c, msg)
}

// CreateDefaultMessage 创建默认留言（内部调用）
func CreateDefaultMessage(tx *gorm.DB, familyID int64, userID, userName, userAvatar string) error {
	defaultMsg := model.MessageDBModel{
		FamilyID:   familyID,
		UserID:     userID,
		UserName:   userName,
		UserAvatar: userAvatar,
		Content:    "欢迎来到我们的家庭厨房！在这里可以分享食谱、点餐和留言。",
	}
	return tx.Create(&defaultMsg).Error
}

type MessageResponse struct {
	model.MessageDBModel
	Replies []model.MessageDBModel `json:"replies"`
}

// GetMessages
// 方法/路径：GET /api/message/list
// 鉴权：需 JWT
// 返回：{ ...message, replies: MessageDBModel[] }[]（根留言倒序、回复正序）
func GetMessages(_ context.Context, c *app.RequestContext) {
	openId, err := util.GetOpenId(c)
	if err != nil {
		util.ReturnErrorJSON(c, "未登录", err)
		return
	}

	var user model.UserInfoDBModel
	if err := db.Get().Where("openId = ?", openId).First(&user).Error; err != nil {
		util.ReturnErrorJSON(c, "用户不存在", err)
		return
	}

	if user.FamilyID == 0 {
		util.ReturnSuccessJSON(c, []MessageResponse{})
		return
	}

	var allMessages []model.MessageDBModel
	if err := db.Get().Where("family_id = ?", user.FamilyID).Find(&allMessages).Error; err != nil {
		util.ReturnErrorJSON(c, "查询失败", err)
		return
	}

	// Build Tree
	var roots []model.MessageDBModel
	var repliesMap = make(map[int64][]model.MessageDBModel)

	for _, m := range allMessages {
		if m.ParentID == 0 {
			roots = append(roots, m)
		} else {
			pid := m.ParentID
			repliesMap[pid] = append(repliesMap[pid], m)
		}
	}

	// Sort roots desc (newest first)
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].CreateTime.After(roots[j].CreateTime)
	})

	var result []MessageResponse
	for _, root := range roots {
		replies := repliesMap[root.ID]
		// Sort replies asc (oldest first)
		sort.Slice(replies, func(i, j int) bool {
			return replies[i].CreateTime.Before(replies[j].CreateTime)
		})
		if replies == nil {
			replies = []model.MessageDBModel{}
		}

		result = append(result, MessageResponse{
			MessageDBModel: root,
			Replies:        replies,
		})
	}

	util.ReturnSuccessJSON(c, result)
}

// DeleteMessage
// 方法/路径：POST /api/message/delete?id=xxx
// 鉴权：需 JWT；家长或留言本人可删除
// 返回：字符串提示；若删除根留言，则级联删除其所有回复
func DeleteMessage(_ context.Context, c *app.RequestContext) {
	openId, err := util.GetOpenId(c)
	if err != nil {
		util.ReturnErrorJSON(c, "未登录", err)
		return
	}

	idStr := c.Query("id")
	if idStr == "" {
		util.ReturnErrorJSON(c, "参数错误", nil)
		return
	}

	var user model.UserInfoDBModel
	if err := db.Get().Where("openId = ?", openId).First(&user).Error; err != nil {
		util.ReturnErrorJSON(c, "用户不存在", err)
		return
	}

	var msg model.MessageDBModel
	if err := db.Get().Where("id = ?", idStr).First(&msg).Error; err != nil {
		util.ReturnErrorJSON(c, "留言不存在", err)
		return
	}

	// Permission check
	if user.Role != "admin" && msg.UserID != user.OpenID {
		util.ReturnErrorJSON(c, "无权删除", nil)
		return
	}

	var responseHandled bool
	err = dao.Transaction(func(tx *gorm.DB) error {
		// Delete replies if root
		if msg.ParentID == 0 {
			if err := tx.Where("parent_id = ?", msg.ID).Delete(&model.MessageDBModel{}).Error; err != nil {
				util.ReturnErrorJSON(c, "删除回复失败", err)
				responseHandled = true
				return err
			}
		}

		if err := tx.Delete(&msg).Error; err != nil {
			util.ReturnErrorJSON(c, "删除失败", err)
			responseHandled = true
			return err
		}

		return nil
	})

	if err != nil {
		if !responseHandled {
			util.ReturnErrorJSON(c, "系统错误", err)
		}
		return
	}

	util.ReturnSuccessJSON(c, "删除成功")
}
