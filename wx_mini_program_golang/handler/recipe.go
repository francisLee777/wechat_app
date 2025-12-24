package handler

import (
	"context"
	"encoding/json"
	"wxcloudrun-golang/db"
	"wxcloudrun-golang/db/dao"
	"wxcloudrun-golang/db/model"
	"wxcloudrun-golang/util"

	"github.com/cloudwego/hertz/pkg/app"
	"gorm.io/gorm"
)

type RecipeDTO struct {
	ID            int64    `json:"id"`
	FamilyID      int64    `json:"familyId"`
	Name          string   `json:"name"`
	Images        []string `json:"images"`
	Content       string   `json:"content"`
	CreatorID     string   `json:"creatorId"`
	CreatorName   string   `json:"creatorName"`
	CreatorAvatar string   `json:"creatorAvatar"`
	SortOrder     int64    `json:"sortOrder"`
	CreateTime    int64    `json:"createTime"`
	UpdateTime    int64    `json:"updateTime"`
}

// RecipeDTO 用于前端展示的食谱数据传输结构
func convertRecipeToDTO(r *model.RecipeDBModel) *RecipeDTO {
	var images []string
	if r.Images != "" {
		_ = json.Unmarshal([]byte(r.Images), &images)
	}
	if images == nil {
		images = []string{}
	}
	return &RecipeDTO{
		ID:            r.ID,
		FamilyID:      r.FamilyID,
		Name:          r.Name,
		Images:        images,
		Content:       r.Content,
		CreatorID:     r.CreatorID,
		CreatorName:   r.CreatorName,
		CreatorAvatar: r.CreatorAvatar,
		SortOrder:     r.SortOrder,
		CreateTime:    r.CreateTime.UnixMilli(),
		UpdateTime:    r.UpdateTime.UnixMilli(),
	}
}

type AddRecipeRequest struct {
	Name    string   `json:"name"`
	Images  []string `json:"images"`
	Content string   `json:"content"`
}

// AddRecipe
// 方法/路径：POST /api/recipe/add
// 鉴权：需 JWT；角色必须为 admin
// 请求：{ name: string, images: string[], content: string }
// 返回：RecipeDTO
func AddRecipe(_ context.Context, c *app.RequestContext) {
	openId, err := util.GetOpenId(c)
	if err != nil {
		util.ReturnErrorJSON(c, "未登录", err)
		return
	}

	var req AddRecipeRequest
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

	if user.Role != "admin" {
		util.ReturnErrorJSON(c, "只有家长可以创建食谱", nil)
		return
	}

	imagesJson, _ := json.Marshal(req.Images)
	recipe := model.RecipeDBModel{
		FamilyID:      user.FamilyID,
		Name:          req.Name,
		Images:        string(imagesJson),
		Content:       req.Content,
		CreatorID:     user.OpenID,
		CreatorName:   user.UserNickName,
		CreatorAvatar: user.UserIconURL,
	}

	if err := db.Get().Create(&recipe).Error; err != nil {
		util.ReturnErrorJSON(c, "创建食谱失败", err)
		return
	}

	// Update sort order to match ID or something unique if needed, or just use timestamp
	db.Get().Model(&recipe).Update("sort_order", recipe.CreateTime.Second())
	recipe.SortOrder = int64(recipe.CreateTime.Second())

	util.ReturnSuccessJSON(c, convertRecipeToDTO(&recipe))
}

// CreateDefaultRecipe 创建默认食谱（内部调用）
func CreateDefaultRecipe(tx *gorm.DB, familyID int64, creatorID, creatorName, creatorAvatar string, imagesJSON string) error {
	defaultRecipe := model.RecipeDBModel{
		FamilyID:      familyID,
		Name:          "番茄炒蛋 (示例)",
		Images:        imagesJSON,
		Content:       "1. 番茄切块，鸡蛋打散。\n2. 锅热油，炒熟鸡蛋盛出。\n3. 锅留底油，炒软番茄。\n4. 加入鸡蛋，调味出锅。\n这是一个示例食谱，您可以添加更多美味佳肴！",
		SortOrder:     0,
		CreatorID:     creatorID,
		CreatorName:   creatorName,
		CreatorAvatar: creatorAvatar,
	}
	return tx.Create(&defaultRecipe).Error
}

// GetRecipes
// 方法/路径：GET /api/recipe/list
// 鉴权：需 JWT
// 返回：RecipeDTO[]（按 sort_order 降序）
func GetRecipes(_ context.Context, c *app.RequestContext) {
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
		util.ReturnSuccessJSON(c, []RecipeDTO{})
		return
	}

	var recipes []model.RecipeDBModel
	if err := db.Get().Where("family_id = ?", user.FamilyID).Order("sort_order desc").Find(&recipes).Error; err != nil {
		util.ReturnErrorJSON(c, "查询失败", err)
		return
	}

	var dtos []RecipeDTO
	for _, recipe := range recipes {
		dtos = append(dtos, *convertRecipeToDTO(&recipe))
	}

	util.ReturnSuccessJSON(c, dtos)
}

type UpdateRecipeRequest struct {
	ID      int64    `json:"id"`
	Name    string   `json:"name"`
	Images  []string `json:"images"`
	Content string   `json:"content"`
}

// UpdateRecipe
// 方法/路径：POST /api/recipe/update
// 鉴权：需 JWT；角色必须为 admin
// 请求：{ id, name, images[], content }
// 返回：RecipeDTO
func UpdateRecipe(_ context.Context, c *app.RequestContext) {
	openId, err := util.GetOpenId(c)
	if err != nil {
		util.ReturnErrorJSON(c, "未登录", err)
		return
	}

	var req UpdateRecipeRequest
	if err := util.BindJson(c, &req); err != nil {
		util.ReturnErrorJSON(c, "参数错误", err)
		return
	}

	var user model.UserInfoDBModel
	if err := db.Get().Where("openId = ?", openId).First(&user).Error; err != nil {
		util.ReturnErrorJSON(c, "用户不存在", err)
		return
	}

	if user.Role != "admin" {
		util.ReturnErrorJSON(c, "无权操作", nil)
		return
	}

	var recipe model.RecipeDBModel
	if err := db.Get().First(&recipe, req.ID).Error; err != nil {
		util.ReturnErrorJSON(c, "食谱不存在", err)
		return
	}

	if recipe.FamilyID != user.FamilyID {
		util.ReturnErrorJSON(c, "无权操作", nil)
		return
	}

	imagesJson, _ := json.Marshal(req.Images)
	recipe.Name = req.Name
	recipe.Images = string(imagesJson)
	recipe.Content = req.Content

	if err := db.Get().Save(&recipe).Error; err != nil {
		util.ReturnErrorJSON(c, "更新失败", err)
		return
	}

	util.ReturnSuccessJSON(c, convertRecipeToDTO(&recipe))
}

// DeleteRecipe
// 方法/路径：POST /api/recipe/delete?id=xxx
// 鉴权：需 JWT；角色必须为 admin
// 返回：字符串提示
func DeleteRecipe(_ context.Context, c *app.RequestContext) {
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

	if user.Role != "admin" {
		util.ReturnErrorJSON(c, "无权操作", nil)
		return
	}

	var recipe model.RecipeDBModel
	if err := db.Get().Where("id = ?", idStr).First(&recipe).Error; err != nil {
		util.ReturnErrorJSON(c, "食谱不存在", err)
		return
	}

	if recipe.FamilyID != user.FamilyID {
		util.ReturnErrorJSON(c, "无权操作", nil)
		return
	}

	if err := db.Get().Delete(&recipe).Error; err != nil {
		util.ReturnErrorJSON(c, "删除失败", err)
		return
	}

	util.ReturnSuccessJSON(c, "删除成功")
}

type ReorderRecipeRequest struct {
	ID        int64  `json:"id"`
	Direction string `json:"direction"` // up, down
}

// ReorderRecipe
// 方法/路径：POST /api/recipe/reorder
// 鉴权：需 JWT；角色必须为 admin
// 请求：{ id: number, direction: 'up'|'down' }
// 返回：字符串提示
func ReorderRecipe(_ context.Context, c *app.RequestContext) {
	openId, err := util.GetOpenId(c)
	if err != nil {
		util.ReturnErrorJSON(c, "未登录", err)
		return
	}

	var req ReorderRecipeRequest
	if err := util.BindJson(c, &req); err != nil {
		util.ReturnErrorJSON(c, "参数错误", err)
		return
	}

	var user model.UserInfoDBModel
	if err := db.Get().Where("openId = ?", openId).First(&user).Error; err != nil {
		util.ReturnErrorJSON(c, "用户不存在", err)
		return
	}

	if user.Role != "admin" {
		util.ReturnErrorJSON(c, "无权操作", nil)
		return
	}

	var current model.RecipeDBModel
	if err := db.Get().First(&current, req.ID).Error; err != nil {
		util.ReturnErrorJSON(c, "食谱不存在", err)
		return
	}

	if current.FamilyID != user.FamilyID {
		util.ReturnErrorJSON(c, "无权操作", nil)
		return
	}

	var target model.RecipeDBModel
	var errTarget error

	// sort_order desc.
	// Up means moving to earlier position (visual top), so HIGHER sort_order.
	// Down means moving to later position (visual bottom), so LOWER sort_order.
	if req.Direction == "up" {
		// Find the one with smallest SortOrder that is > current.SortOrder
		errTarget = db.Get().Where("family_id = ? AND sort_order > ?", user.FamilyID, current.SortOrder).Order("sort_order asc").First(&target).Error
	} else {
		// Find the one with largest SortOrder that is < current.SortOrder
		errTarget = db.Get().Where("family_id = ? AND sort_order < ?", user.FamilyID, current.SortOrder).Order("sort_order desc").First(&target).Error
	}

	if errTarget != nil {
		// No target found, meaning already at top or bottom
		util.ReturnSuccessJSON(c, "No change")
		return
	}

	// Swap
	tempOrder := current.SortOrder
	current.SortOrder = target.SortOrder
	target.SortOrder = tempOrder

	var responseHandled bool
	err = dao.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&current).Error; err != nil {
			util.ReturnErrorJSON(c, "更新失败", err)
			responseHandled = true
			return err
		}
		if err := tx.Save(&target).Error; err != nil {
			util.ReturnErrorJSON(c, "更新失败", err)
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

	util.ReturnSuccessJSON(c, "排序完成")
}

type BatchUpdateRecipesRequest struct {
	Recipes []struct {
		ID        int64 `json:"id"`
		SortOrder int64 `json:"sortOrder"`
	} `json:"recipes"`
}

// BatchUpdateRecipes
// 方法/路径：POST /api/recipe/batchUpdate
// 鉴权：需 JWT；角色必须为 admin
// 请求：{ recipes: { id, sortOrder }[] }
// 返回：字符串提示
func BatchUpdateRecipes(_ context.Context, c *app.RequestContext) {
	openId, err := util.GetOpenId(c)
	if err != nil {
		util.ReturnErrorJSON(c, "未登录", err)
		return
	}

	var req BatchUpdateRecipesRequest
	if err := util.BindJson(c, &req); err != nil {
		util.ReturnErrorJSON(c, "参数错误", err)
		return
	}

	var user model.UserInfoDBModel
	if err := db.Get().Where("openId = ?", openId).First(&user).Error; err != nil {
		util.ReturnErrorJSON(c, "用户不存在", err)
		return
	}

	if user.Role != "admin" {
		util.ReturnErrorJSON(c, "无权操作", nil)
		return
	}

	var responseHandled bool
	err = dao.Transaction(func(tx *gorm.DB) error {
		for _, item := range req.Recipes {
			// Only update sort_order and verify family_id
			result := tx.Model(&model.RecipeDBModel{}).
				Where("id = ? AND family_id = ?", item.ID, user.FamilyID).
				Update("sort_order", item.SortOrder)

			if result.Error != nil {
				util.ReturnErrorJSON(c, "更新失败", result.Error)
				responseHandled = true
				return result.Error
			}
		}
		return nil
	})

	if err != nil {
		if !responseHandled {
			util.ReturnErrorJSON(c, "系统错误", err)
		}
		return
	}

	util.ReturnSuccessJSON(c, "批量更新成功")
}

// GetRecipe
// 方法/路径：GET /api/recipe/info?id=xxx
// 鉴权：需 JWT；仅同家庭用户可查看
// 返回：RecipeDTO
func GetRecipe(_ context.Context, c *app.RequestContext) {
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

	var recipe model.RecipeDBModel
	if err := db.Get().First(&recipe, idStr).Error; err != nil {
		util.ReturnErrorJSON(c, "食谱不存在", err)
		return
	}

	if recipe.FamilyID != user.FamilyID {
		util.ReturnErrorJSON(c, "无权查看", nil)
		return
	}

	util.ReturnSuccessJSON(c, convertRecipeToDTO(&recipe))
}
