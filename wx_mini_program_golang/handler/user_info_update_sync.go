package handler

import (
	"fmt"
	"wxcloudrun-golang/db"
	"wxcloudrun-golang/db/model"
)

// SyncType 定义同步类型
type SyncType int

const (
	SyncTypeNickName SyncType = iota
	SyncTypeAvatar
)

// SyncUserProfile 同步用户资料到其他关联表（冗余数据更新）
// 注意：此操作应异步进行或不阻塞主流程，且不应触发被更新表的 update_time
func SyncUserProfile(openId string, syncType SyncType, value string) {
	var err error

	// 1. 同步更新留言板 (Message)
	msgQuery := db.Get().Model(&model.MessageDBModel{}).Where("user_id = ?", openId)

	switch syncType {
	case SyncTypeNickName:
		err = msgQuery.UpdateColumn("user_name", value).Error
	case SyncTypeAvatar:
		err = msgQuery.UpdateColumn("user_avatar", value).Error
	}

	if err != nil {
		fmt.Printf("[SyncUserProfile] Failed to sync message for user %s: %v\n", openId, err)
	}

	// 2. 同步更新食谱表 (Recipe)
	recipeQuery := db.Get().Model(&model.RecipeDBModel{}).Where("creator_id = ?", openId)

	switch syncType {
	case SyncTypeNickName:
		err = recipeQuery.UpdateColumn("creator_name", value).Error
	case SyncTypeAvatar:
		err = recipeQuery.UpdateColumn("creator_avatar", value).Error
	}

	if err != nil {
		fmt.Printf("[SyncUserProfile] Failed to sync recipe for user %s: %v\n", openId, err)
	}
}
