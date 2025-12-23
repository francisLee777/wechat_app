package model

import "time"

const TableNameRecipe = "recipe"

type Recipe struct {
	ID         int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	FamilyID   int64     `gorm:"column:family_id" json:"familyId"`
	Name       string    `gorm:"column:name" json:"name"`
	Images     string    `gorm:"column:images" json:"images"` // Stored as JSON string
	Content    string    `gorm:"column:content" json:"content"`
	SortOrder  int64     `gorm:"column:sort_order" json:"sortOrder"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
}

func (*Recipe) TableName() string {
	return TableNameRecipe
}
