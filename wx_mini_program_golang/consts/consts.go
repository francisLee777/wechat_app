package consts

import (
	"fmt"
	"path/filepath"
)

const (
	AbsolutePath = "/app"

	// DataPath 用户存储的根目录
	DataPath     = "/data/user_data"
	FamilyPrefix = DataPath + "/family_png" // 家庭
	UserPrefix   = DataPath + "/user_png"   // 昵称和头像

	// AssetFilePrefix 公共文件根目录
	AssetFilePrefix    = "/assets"
	TemplateFilePrefix = AssetFilePrefix + "/templates"
)

func FamilyDirPath(familyID int64) string {
	return filepath.Join(FamilyPrefix, fmt.Sprint(familyID))
}

func FamilyFilePathStr(familyIDStr string, fileName string) string {
	return filepath.Join(FamilyPrefix, familyIDStr, fileName)
}

func UserDirPath(openID string) string {
	return filepath.Join(UserPrefix, openID)
}

func UserFilePathStr(openID string, fileName string) string {
	return filepath.Join(UserDirPath(openID), fileName)
}
