package model

import "time"

const TableNameFamily = "family"

type Family struct {
	ID         int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name       string    `gorm:"column:name" json:"name"`
	OwnerID    string    `gorm:"column:owner_id" json:"ownerId"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
}

func (*Family) TableName() string {
	return TableNameFamily
}
