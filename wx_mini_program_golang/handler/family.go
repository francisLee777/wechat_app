package handler

import (
	"context"
	"fmt"
	"path/filepath"
	"wxcloudrun-golang/consts"
	"wxcloudrun-golang/db"
	"wxcloudrun-golang/db/dao"
	"wxcloudrun-golang/db/model"
	"wxcloudrun-golang/util"

	"github.com/cloudwego/hertz/pkg/app"
	"gorm.io/gorm"
)

// CreateFamilyRequest 创建家庭请求
type CreateFamilyRequest struct {
	Name string `json:"name"`
}

// CreateFamily
// 方法/路径：POST /api/family/create
// 鉴权：需 JWT；用户未加入家庭
// 请求：{ name: string }
// 返回：FamilyDBModel；同时写入默认留言与示例食谱
func CreateFamily(_ context.Context, c *app.RequestContext) {
	openId, err := util.GetOpenId(c)
	if err != nil {
		util.ReturnErrorJSON(c, "未登录", err)
		return
	}

	var req CreateFamilyRequest
	if err := util.BindJson(c, &req); err != nil {
		util.ReturnErrorJSON(c, "参数错误", err)
		return
	}

	var family model.FamilyDBModel
	var responseHandled bool

	err = dao.Transaction(func(tx *gorm.DB) error {
		// Check user
		var user model.UserInfoDBModel
		if err := tx.Where("openId = ?", openId).First(&user).Error; err != nil {
			util.ReturnErrorJSON(c, "用户不存在", err)
			responseHandled = true
			return err
		}

		if user.FamilyID != 0 {
			// 已加入家庭 直接返回
			responseHandled = true
			util.ReturnSuccessJSON(c, family)
			return nil
		}

		// Create Family
		family = model.FamilyDBModel{
			Name:    req.Name,
			OwnerID: openId,
		}

		if err := tx.Create(&family).Error; err != nil {
			util.ReturnErrorJSON(c, "创建家庭失败", err)
			responseHandled = true
			return err
		}

		// Update User
		user.FamilyID = family.ID
		user.Role = "admin"
		if err := tx.Save(&user).Error; err != nil {
			util.ReturnErrorJSON(c, "更新用户状态失败", err)
			responseHandled = true
			return err
		}

		// Create Default Message
		if err := CreateDefaultMessage(tx, family.ID, openId, user.UserNickName, user.UserIconURL); err != nil {
			util.ReturnErrorJSON(c, "创建默认留言失败", err)
			responseHandled = true
			return err
		}

		// Create Default Recipe
		// Try to write default image from local assets into family directory
		var imagesJSON = "[]"
		{
			fileName := "default_food.png"
			src := filepath.Join(consts.AssetFilePrefix, fileName)
			data, err := util.BizOsReadFile(src)
			if err == nil && len(data) > 0 {
				imageURL := consts.FamilyFilePathStr(fmt.Sprint(family.ID), fileName)
				if err := util.WriteFileSafe(imageURL, data); err == nil {
					imagesJSON = fmt.Sprintf("[\"%s\"]", imageURL)
				}
			} else {
				// Log error but don't fail transaction just for image copy failure?
				// Original code: util.ReturnErrorJSON(c, "默认食谱的图片写入失败", err)
				// Let's keep strict behavior
				util.ReturnErrorJSON(c, "默认食谱的图片写入失败", err)
				responseHandled = true
				return err
			}
		}
		if err := CreateDefaultRecipe(tx, family.ID, openId, user.UserNickName, user.UserIconURL, imagesJSON); err != nil {
			util.ReturnErrorJSON(c, "创建默认食谱失败", err)
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

	util.ReturnSuccessJSON(c, family)
}

// DeleteFamily 解散家庭
// 方法/路径：POST /api/family/delete
// 鉴权：需 JWT；仅家长可操作
// 说明：删除家庭所有数据（留言、食谱、家庭记录），重置所有成员状态，删除物理文件
func DeleteFamily(_ context.Context, c *app.RequestContext) {
	openId, err := util.GetOpenId(c)
	if err != nil {
		util.ReturnErrorJSON(c, "未登录", err)
		return
	}

	var familyID int64
	var responseHandled bool

	err = dao.Transaction(func(tx *gorm.DB) error {
		var user model.UserInfoDBModel
		if err := tx.Where("openId = ?", openId).First(&user).Error; err != nil {
			util.ReturnErrorJSON(c, "用户不存在", err)
			responseHandled = true
			return err
		}

		if user.FamilyID == 0 {
			util.ReturnErrorJSON(c, "您尚未加入家庭", nil)
			responseHandled = true
			return fmt.Errorf("not in family")
		}

		if user.Role != "admin" {
			util.ReturnErrorJSON(c, "只有家长可以解散家庭", nil)
			responseHandled = true
			return fmt.Errorf("permission denied")
		}

		familyID = user.FamilyID

		// 1. 删除该家庭的所有留言
		if err := tx.Where("family_id = ?", familyID).Delete(&model.MessageDBModel{}).Error; err != nil {
			util.ReturnErrorJSON(c, "删除留言失败", err)
			responseHandled = true
			return err
		}

		// 2. 删除该家庭的所有食谱
		if err := tx.Where("family_id = ?", familyID).Delete(&model.RecipeDBModel{}).Error; err != nil {
			util.ReturnErrorJSON(c, "删除食谱失败", err)
			responseHandled = true
			return err
		}

		// 3. 重置该家庭所有成员的状态
		if err := tx.Model(&model.UserInfoDBModel{}).Where("family_id = ?", familyID).
			Updates(map[string]interface{}{"family_id": 0, "role": "member"}).Error; err != nil {
			util.ReturnErrorJSON(c, "重置成员状态失败", err)
			responseHandled = true
			return err
		}

		// 4. 删除家庭记录
		if err := tx.Delete(&model.FamilyDBModel{}, familyID).Error; err != nil {
			util.ReturnErrorJSON(c, "删除家庭失败", err)
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

	// 5. 删除物理文件目录 (忽略错误，不影响业务逻辑)
	// 注意：这里可能会删除失败（如文件被占用），但一般在 Linux/Docker 环境下没问题
	familyDir := consts.FamilyDirPath(familyID)
	if err := util.BizOsMkdirAll(familyDir); err != nil {
		fmt.Printf("Failed to remove family dir %s: %v\n", familyDir, err)
	}

	util.ReturnSuccessJSON(c, "家庭已解散")
}

// JoinFamilyRequest 加入家庭请求
type JoinFamilyRequest struct {
	FamilyID int64 `json:"familyId"`
}

// JoinFamily
// 方法/路径：POST /api/family/join
// 鉴权：需 JWT；家庭最多 4 人
// 请求：{ familyId: number }
// 返回：FamilyDBModel
func JoinFamily(_ context.Context, c *app.RequestContext) {
	openId, err := util.GetOpenId(c)
	if err != nil {
		util.ReturnErrorJSON(c, "未登录", err)
		return
	}

	var req JoinFamilyRequest
	if err := util.BindJson(c, &req); err != nil {
		util.ReturnErrorJSON(c, "参数错误", err)
		return
	}

	var family model.FamilyDBModel
	var responseHandled bool

	err = dao.Transaction(func(tx *gorm.DB) error {
		var user model.UserInfoDBModel
		if err := tx.Where("openId = ?", openId).First(&user).Error; err != nil {
			util.ReturnErrorJSON(c, "用户不存在", err)
			responseHandled = true
			return err
		}

		if user.FamilyID != 0 {
			util.ReturnErrorJSON(c, "您已加入家庭", nil)
			responseHandled = true
			return fmt.Errorf("already in family")
		}

		if err := tx.First(&family, req.FamilyID).Error; err != nil {
			util.ReturnErrorJSON(c, "家庭不存在", err)
			responseHandled = true
			return err
		}

		// Count members
		var count int64
		tx.Model(&model.UserInfoDBModel{}).Where("family_id = ?", req.FamilyID).Count(&count)
		if count >= 4 {
			util.ReturnErrorJSON(c, "家庭成员已满", nil)
			responseHandled = true
			return fmt.Errorf("family full")
		}

		user.FamilyID = req.FamilyID
		user.Role = "member"
		if err := tx.Save(&user).Error; err != nil {
			util.ReturnErrorJSON(c, "加入家庭失败", err)
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

	util.ReturnSuccessJSON(c, family)
}

// GetFamilyMembers
// 方法/路径：GET /api/family/members
// 鉴权：需 JWT
// 返回：UserInfoDBModel[]（当前家庭成员）
func GetFamilyMembers(_ context.Context, c *app.RequestContext) {
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
		util.ReturnSuccessJSON(c, []model.UserInfoDBModel{})
		return
	}

	var members []model.UserInfoDBModel
	if err := db.Get().Where("family_id = ?", user.FamilyID).Find(&members).Error; err != nil {
		util.ReturnErrorJSON(c, "查询失败", err)
		return
	}

	util.ReturnSuccessJSON(c, members)
}

// QuitFamily
// 方法/路径：POST /api/family/quit
// 鉴权：需 JWT；家长不可退出
// 返回：字符串提示
func QuitFamily(_ context.Context, c *app.RequestContext) {
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
		util.ReturnErrorJSON(c, "您尚未加入家庭", nil)
		return
	}

	if user.Role == "admin" {
		util.ReturnErrorJSON(c, "家长不能退出家庭", nil)
		return
	}

	user.FamilyID = 0
	user.Role = "member"
	if err := db.Get().Save(&user).Error; err != nil {
		util.ReturnErrorJSON(c, "退出失败", err)
		return
	}

	util.ReturnSuccessJSON(c, "退出成功")
}

// RemoveMemberRequest 移除成员请求
type RemoveMemberRequest struct {
	MemberOpenID string `json:"memberOpenId"`
}

// RemoveMember
// 方法/路径：POST /api/family/removeMember
// 鉴权：需 JWT；角色必须为 admin
// 请求：{ memberOpenId: string }
// 返回：字符串提示
func RemoveMember(_ context.Context, c *app.RequestContext) {
	openId, err := util.GetOpenId(c)
	if err != nil {
		util.ReturnErrorJSON(c, "未登录", err)
		return
	}

	var req RemoveMemberRequest
	if err := util.BindJson(c, &req); err != nil {
		util.ReturnErrorJSON(c, "参数错误", err)
		return
	}

	var responseHandled bool
	err = dao.Transaction(func(tx *gorm.DB) error {
		var admin model.UserInfoDBModel
		if err := tx.Where("openId = ?", openId).First(&admin).Error; err != nil {
			util.ReturnErrorJSON(c, "用户不存在", err)
			responseHandled = true
			return err
		}

		if admin.Role != "admin" {
			util.ReturnErrorJSON(c, "无权操作", nil)
			responseHandled = true
			return fmt.Errorf("permission denied")
		}

		var member model.UserInfoDBModel
		if err := tx.Where("openId = ?", req.MemberOpenID).First(&member).Error; err != nil {
			util.ReturnErrorJSON(c, "成员不存在", err)
			responseHandled = true
			return err
		}

		if member.FamilyID != admin.FamilyID {
			util.ReturnErrorJSON(c, "该成员不在您的家庭", nil)
			responseHandled = true
			return fmt.Errorf("not in same family")
		}

		member.FamilyID = 0
		member.Role = "member"
		if err := tx.Save(&member).Error; err != nil {
			util.ReturnErrorJSON(c, "移除失败", err)
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

	util.ReturnSuccessJSON(c, "移除成功")
}

// GetFamilyInfo
// 方法/路径：GET /api/family/info?id=xxx
// 鉴权：需 JWT；仅同家庭用户可查看
// 返回：FamilyDBModel
func GetFamilyInfo(_ context.Context, c *app.RequestContext) {
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

	var family model.FamilyDBModel
	if err := db.Get().First(&family, idStr).Error; err != nil {
		util.ReturnErrorJSON(c, "家庭不存在", err)
		return
	}

	if user.FamilyID != family.ID {
		util.ReturnErrorJSON(c, "无权查看或家庭不匹配", nil)
		return
	}

	util.ReturnSuccessJSON(c, family)
}
