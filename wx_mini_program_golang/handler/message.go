package handler

import (
	"net/http"
	"sort"
	"wxcloudrun-golang/db"
	"wxcloudrun-golang/db/model"
	"wxcloudrun-golang/util"
)

type AddMessageRequest struct {
	Content         string `json:"content"`
	ParentID        *int64 `json:"parentId"`
	ReplyToUserID   string `json:"replyToUserId"`
	ReplyToUserName string `json:"replyToUserName"`
}

func AddMessage(w http.ResponseWriter, r *http.Request) {
	openId, err := util.GetOpenIdFromHeader(r)
	if err != nil {
		util.ReturnErrorJSON(w, "未登录", err)
		return
	}

	var req AddMessageRequest
	if err := util.BindJson(r, &req); err != nil {
		util.ReturnErrorJSON(w, "参数错误", err)
		return
	}

	var user model.UserInfoDBModel
	if err := db.Get().Where("openId = ?", openId).First(&user).Error; err != nil {
		util.ReturnErrorJSON(w, "用户不存在", err)
		return
	}

	if user.FamilyID == 0 {
		util.ReturnErrorJSON(w, "请先加入家庭", nil)
		return
	}

	msg := model.Message{
		FamilyID:        user.FamilyID,
		UserID:          user.OpenID,
		UserName:        user.UserNickName,
		UserAvatar:      user.UserIconURL,
		Content:         req.Content,
		ParentID:        req.ParentID,
		ReplyToUserID:   req.ReplyToUserID,
		ReplyToUserName: req.ReplyToUserName,
	}

	if err := db.Get().Create(&msg).Error; err != nil {
		util.ReturnErrorJSON(w, "发送失败", err)
		return
	}

	util.ReturnSuccessJSON(w, msg)
}

type MessageResponse struct {
	model.Message
	Replies []model.Message `json:"replies"`
}

func GetMessages(w http.ResponseWriter, r *http.Request) {
	openId, err := util.GetOpenIdFromHeader(r)
	if err != nil {
		util.ReturnErrorJSON(w, "未登录", err)
		return
	}

	var user model.UserInfoDBModel
	if err := db.Get().Where("openId = ?", openId).First(&user).Error; err != nil {
		util.ReturnErrorJSON(w, "用户不存在", err)
		return
	}

	if user.FamilyID == 0 {
		util.ReturnSuccessJSON(w, []MessageResponse{})
		return
	}

	var allMessages []model.Message
	if err := db.Get().Where("family_id = ?", user.FamilyID).Find(&allMessages).Error; err != nil {
		util.ReturnErrorJSON(w, "查询失败", err)
		return
	}

	// Build Tree
	var roots []model.Message
	var repliesMap = make(map[int64][]model.Message)

	for _, m := range allMessages {
		if m.ParentID == nil {
			roots = append(roots, m)
		} else {
			pid := *m.ParentID
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
			replies = []model.Message{}
		}

		result = append(result, MessageResponse{
			Message: root,
			Replies: replies,
		})
	}

	util.ReturnSuccessJSON(w, result)
}

func DeleteMessage(w http.ResponseWriter, r *http.Request) {
	openId, err := util.GetOpenIdFromHeader(r)
	if err != nil {
		util.ReturnErrorJSON(w, "未登录", err)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		util.ReturnErrorJSON(w, "参数错误", nil)
		return
	}

	var user model.UserInfoDBModel
	if err := db.Get().Where("openId = ?", openId).First(&user).Error; err != nil {
		util.ReturnErrorJSON(w, "用户不存在", err)
		return
	}

	var msg model.Message
	if err := db.Get().Where("id = ?", idStr).First(&msg).Error; err != nil {
		util.ReturnErrorJSON(w, "留言不存在", err)
		return
	}

	// Permission check
	if user.Role != "admin" && msg.UserID != user.OpenID {
		util.ReturnErrorJSON(w, "无权删除", nil)
		return
	}

	tx := db.Get().Begin()
	// Delete replies if root
	if msg.ParentID == nil {
		if err := tx.Where("parent_id = ?", msg.ID).Delete(&model.Message{}).Error; err != nil {
			tx.Rollback()
			util.ReturnErrorJSON(w, "删除回复失败", err)
			return
		}
	}

	if err := tx.Delete(&msg).Error; err != nil {
		tx.Rollback()
		util.ReturnErrorJSON(w, "删除失败", err)
		return
	}

	tx.Commit()
	util.ReturnSuccessJSON(w, "删除成功")
}
