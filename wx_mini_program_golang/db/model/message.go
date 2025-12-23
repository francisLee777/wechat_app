package model

import "time"

const TableNameMessage = "message"

type Message struct {
	ID              int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	FamilyID        int64     `gorm:"column:family_id" json:"familyId"`
	UserID          string    `gorm:"column:user_id" json:"userId"`
	UserName        string    `gorm:"column:user_name" json:"userName"`
	UserAvatar      string    `gorm:"column:user_avatar" json:"userAvatar"`
	Content         string    `gorm:"column:content" json:"content"`
	ParentID        *int64    `gorm:"column:parent_id" json:"parentId"`
	ReplyToUserID   string    `gorm:"column:reply_to_user_id" json:"replyToUserId"`
	ReplyToUserName string    `gorm:"column:reply_to_user_name" json:"replyToUserName"`
	CreateTime      time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
}

func (*Message) TableName() string {
	return TableNameMessage
}
