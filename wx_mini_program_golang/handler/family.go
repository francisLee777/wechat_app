package handler

import (
	"net/http"
	"wxcloudrun-golang/db"
	"wxcloudrun-golang/db/model"
	"wxcloudrun-golang/util"
)

// CreateFamilyRequest 创建家庭请求
type CreateFamilyRequest struct {
	Name string `json:"name"`
}

// CreateFamily 创建家庭
func CreateFamily(w http.ResponseWriter, r *http.Request) {
	openId, err := util.GetOpenIdFromHeader(r)
	if err != nil {
		util.ReturnErrorJSON(w, "未登录", err)
		return
	}

	var req CreateFamilyRequest
	if err := util.BindJson(r, &req); err != nil {
		util.ReturnErrorJSON(w, "参数错误", err)
		return
	}

	tx := db.Get().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check user
	var user model.UserInfoDBModel
	if err := tx.Where("openId = ?", openId).First(&user).Error; err != nil {
		tx.Rollback()
		util.ReturnErrorJSON(w, "用户不存在", err)
		return
	}

	if user.FamilyID != 0 {
		tx.Rollback()
		util.ReturnErrorJSON(w, "您已加入家庭", nil)
		return
	}

	// Create Family
	family := model.Family{
		Name:    req.Name,
		OwnerID: openId,
	}

	if err := tx.Create(&family).Error; err != nil {
		tx.Rollback()
		util.ReturnErrorJSON(w, "创建家庭失败", err)
		return
	}

	// Update User
	user.FamilyID = family.ID
	user.Role = "admin"
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		util.ReturnErrorJSON(w, "更新用户状态失败", err)
		return
	}

	tx.Commit()
	util.ReturnSuccessJSON(w, family)
}

// JoinFamilyRequest 加入家庭请求
type JoinFamilyRequest struct {
	FamilyID int64 `json:"familyId"`
}

// JoinFamily 加入家庭
func JoinFamily(w http.ResponseWriter, r *http.Request) {
	openId, err := util.GetOpenIdFromHeader(r)
	if err != nil {
		util.ReturnErrorJSON(w, "未登录", err)
		return
	}

	var req JoinFamilyRequest
	if err := util.BindJson(r, &req); err != nil {
		util.ReturnErrorJSON(w, "参数错误", err)
		return
	}

	tx := db.Get().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var user model.UserInfoDBModel
	if err := tx.Where("openId = ?", openId).First(&user).Error; err != nil {
		tx.Rollback()
		util.ReturnErrorJSON(w, "用户不存在", err)
		return
	}

	if user.FamilyID != 0 {
		tx.Rollback()
		util.ReturnErrorJSON(w, "您已加入家庭", nil)
		return
	}

	var family model.Family
	if err := tx.First(&family, req.FamilyID).Error; err != nil {
		tx.Rollback()
		util.ReturnErrorJSON(w, "家庭不存在", err)
		return
	}

	// Count members
	var count int64
	tx.Model(&model.UserInfoDBModel{}).Where("family_id = ?", req.FamilyID).Count(&count)
	if count >= 4 {
		tx.Rollback()
		util.ReturnErrorJSON(w, "家庭成员已满", nil)
		return
	}

	user.FamilyID = req.FamilyID
	user.Role = "member"
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		util.ReturnErrorJSON(w, "加入家庭失败", err)
		return
	}

	tx.Commit()
	util.ReturnSuccessJSON(w, family)
}

// GetFamilyMembers 获取家庭成员
func GetFamilyMembers(w http.ResponseWriter, r *http.Request) {
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
		util.ReturnSuccessJSON(w, []model.UserInfoDBModel{})
		return
	}

	var members []model.UserInfoDBModel
	if err := db.Get().Where("family_id = ?", user.FamilyID).Find(&members).Error; err != nil {
		util.ReturnErrorJSON(w, "查询失败", err)
		return
	}

	util.ReturnSuccessJSON(w, members)
}

// QuitFamily 退出家庭
func QuitFamily(w http.ResponseWriter, r *http.Request) {
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
		util.ReturnErrorJSON(w, "您尚未加入家庭", nil)
		return
	}

	if user.Role == "admin" {
		util.ReturnErrorJSON(w, "家长不能退出家庭", nil)
		return
	}

	user.FamilyID = 0
	user.Role = "member"
	if err := db.Get().Save(&user).Error; err != nil {
		util.ReturnErrorJSON(w, "退出失败", err)
		return
	}

	util.ReturnSuccessJSON(w, "退出成功")
}

// RemoveMemberRequest 移除成员请求
type RemoveMemberRequest struct {
	MemberOpenID string `json:"memberOpenId"`
}

// RemoveMember 移除成员
func RemoveMember(w http.ResponseWriter, r *http.Request) {
	openId, err := util.GetOpenIdFromHeader(r)
	if err != nil {
		util.ReturnErrorJSON(w, "未登录", err)
		return
	}

	var req RemoveMemberRequest
	if err := util.BindJson(r, &req); err != nil {
		util.ReturnErrorJSON(w, "参数错误", err)
		return
	}

	tx := db.Get().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var admin model.UserInfoDBModel
	if err := tx.Where("openId = ?", openId).First(&admin).Error; err != nil {
		tx.Rollback()
		util.ReturnErrorJSON(w, "用户不存在", err)
		return
	}

	if admin.Role != "admin" {
		tx.Rollback()
		util.ReturnErrorJSON(w, "无权操作", nil)
		return
	}

	var member model.UserInfoDBModel
	if err := tx.Where("openId = ?", req.MemberOpenID).First(&member).Error; err != nil {
		tx.Rollback()
		util.ReturnErrorJSON(w, "成员不存在", err)
		return
	}

	if member.FamilyID != admin.FamilyID {
		tx.Rollback()
		util.ReturnErrorJSON(w, "该成员不在您的家庭", nil)
		return
	}

	member.FamilyID = 0
	member.Role = "member"
	if err := tx.Save(&member).Error; err != nil {
		tx.Rollback()
		util.ReturnErrorJSON(w, "移除失败", err)
		return
	}

	tx.Commit()
	util.ReturnSuccessJSON(w, "移除成功")
}

// GetFamilyInfo 获取家庭信息
func GetFamilyInfo(w http.ResponseWriter, r *http.Request) {
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

	var family model.Family
	if err := db.Get().First(&family, idStr).Error; err != nil {
		util.ReturnErrorJSON(w, "家庭不存在", err)
		return
	}

	// Permission check?
	// If the user is in the family, they can see it.
	// Or if the user is joining, maybe they can see basic info (name)?
	// For simplicity, allow if user is in family OR if we want to allow "preview before join".
	// But `manage.ts` uses it when user is already in family.
	// Let's allow it for now. Strict check: user.FamilyID == family.ID

	if user.FamilyID != family.ID {
		// Maybe user is not in THIS family.
		// If we want to support "Search Family by ID to Join", we might allow it.
		// But currently frontend `join.ts` does `joinFamily(id)` directly, doesn't preview.
		// `manage.ts` calls it for current family.
		// So strict check is fine for `manage.ts`.
		// But what if `join.ts` wants to show name? It relies on `joinFamily` return.
		util.ReturnErrorJSON(w, "无权查看或家庭不匹配", nil)
		return
	}

	util.ReturnSuccessJSON(w, family)
}
