package handler

import (
	"context"
	"encoding/json"
	"time"
	"wxcloudrun-golang/db"
	"wxcloudrun-golang/db/model"
	"wxcloudrun-golang/util"

	"github.com/cloudwego/hertz/pkg/app"
)

type MenuDTO struct {
	ID         int64  `json:"id"`
	FamilyID   int64  `json:"familyId"`
	Date       string `json:"date"`
	MealType   int    `json:"mealType"`
	RecipeID   int64  `json:"recipeId"`
	RecipeName string `json:"recipeName"`
	RecipeImg  string `json:"recipeImg"` // 只取第一张图
	UserID     string `json:"userId"`
	UserAvatar string `json:"userAvatar"`
	Remark     string `json:"remark"`
	CreateTime int64  `json:"createTime"`
}

// 辅助函数：从 model 转换为 DTO，需要传入关联的 Recipe 和 User Map
func convertMenuToDTO(m *model.MenuDBModel, recipeMap map[int64]model.RecipeDBModel, userMap map[string]model.UserInfoDBModel) *MenuDTO {
	recipe, okRecipe := recipeMap[m.RecipeID]
	user, okUser := userMap[m.UserID]

	// 处理图片
	var firstImg string
	var recipeName string
	if okRecipe {
		recipeName = recipe.Name
		if recipe.Images != "" {
			var images []string
			if err := json.Unmarshal([]byte(recipe.Images), &images); err == nil && len(images) > 0 {
				firstImg = images[0]
			}
		}
	}

	var userAvatar string
	if okUser {
		userAvatar = user.UserIconURL
	}

	return &MenuDTO{
		ID:         m.ID,
		FamilyID:   m.FamilyID,
		Date:       m.Date.Format("2006-01-02"), // time.Time -> string
		MealType:   int(m.MealType),             // int32 -> int
		RecipeID:   m.RecipeID,
		RecipeName: recipeName,
		RecipeImg:  firstImg,
		UserID:     m.UserID,
		UserAvatar: userAvatar,
		Remark:     m.Remark,
		CreateTime: m.CreateTime.Unix() * 1000,
	}
}

// GetMenuList
// 方法/路径：GET /api/menu/list
// 鉴权：需 JWT
// Query: date=YYYY-MM-DD
// 返回：{ breakfast: [], lunch: [], dinner: [] }
func GetMenuList(_ context.Context, c *app.RequestContext) {
	openId, err := util.GetOpenId(c)
	if err != nil {
		util.ReturnErrorJSON(c, "未登录", err)
		return
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	var user model.UserInfoDBModel
	if err := db.Get().Where("openId = ?", openId).First(&user).Error; err != nil {
		util.ReturnErrorJSON(c, "用户不存在", err)
		return
	}

	if user.FamilyID == 0 {
		util.ReturnSuccessJSON(c, map[string][]MenuDTO{
			"breakfast": {}, "lunch": {}, "dinner": {},
		})
		return
	}

	var menus []model.MenuDBModel
	// Date 字段在数据库是 date 类型，映射为 time.Time
	// 查询时传入 string 也是可以的，GORM 会尝试转换
	if err := db.Get().
		Where("family_id = ? AND date = ?", user.FamilyID, dateStr).
		Order("create_time asc").
		Find(&menus).Error; err != nil {
		util.ReturnErrorJSON(c, "查询失败", err)
		return
	}

	// 收集 ID
	var recipeIDs []int64
	var userIDs []string
	for _, m := range menus {
		recipeIDs = append(recipeIDs, m.RecipeID)
		userIDs = append(userIDs, m.UserID)
	}

	// 批量查询 Recipe
	recipeMap := make(map[int64]model.RecipeDBModel)
	if len(recipeIDs) > 0 {
		var recipes []model.RecipeDBModel
		db.Get().Where("id IN ?", recipeIDs).Find(&recipes)
		for _, r := range recipes {
			recipeMap[r.ID] = r
		}
	}

	// 批量查询 User
	userMap := make(map[string]model.UserInfoDBModel)
	if len(userIDs) > 0 {
		var users []model.UserInfoDBModel
		db.Get().Where("openId IN ?", userIDs).Find(&users)
		for _, u := range users {
			userMap[u.OpenID] = u
		}
	}

	// 分组
	res := map[string][]MenuDTO{
		"breakfast": {},
		"lunch":     {},
		"dinner":    {},
	}

	for _, m := range menus {
		dto := convertMenuToDTO(&m, recipeMap, userMap)
		switch m.MealType {
		case 1:
			res["breakfast"] = append(res["breakfast"], *dto)
		case 2:
			res["lunch"] = append(res["lunch"], *dto)
		case 3:
			res["dinner"] = append(res["dinner"], *dto)
		}
	}

	util.ReturnSuccessJSON(c, res)
}

type AddMenuRequest struct {
	Date     string `json:"date"`
	MealType int    `json:"mealType"` // 1,2,3
	RecipeID int64  `json:"recipeId"`
	Remark   string `json:"remark"`
}

// AddMenu
// 方法/路径：POST /api/menu/add
// 鉴权：需 JWT
func AddMenu(_ context.Context, c *app.RequestContext) {
	openId, err := util.GetOpenId(c)
	if err != nil {
		util.ReturnErrorJSON(c, "未登录", err)
		return
	}

	var req AddMenuRequest
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

	// 校验食谱
	var recipe model.RecipeDBModel
	if err := db.Get().First(&recipe, req.RecipeID).Error; err != nil {
		util.ReturnErrorJSON(c, "食谱不存在", err)
		return
	}
	if recipe.FamilyID != user.FamilyID {
		util.ReturnErrorJSON(c, "食谱不属于当前家庭", nil)
		return
	}

	// 解析日期
	parseDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		util.ReturnErrorJSON(c, "日期格式错误", err)
		return
	}

	menu := model.MenuDBModel{
		FamilyID: user.FamilyID,
		Date:     parseDate,
		MealType: int32(req.MealType),
		RecipeID: req.RecipeID,
		UserID:   user.OpenID,
		Remark:   req.Remark,
	}

	if err := db.Get().Create(&menu).Error; err != nil {
		util.ReturnErrorJSON(c, "添加失败", err)
		return
	}

	// 构造返回数据
	recipeMap := map[int64]model.RecipeDBModel{recipe.ID: recipe}
	userMap := map[string]model.UserInfoDBModel{user.OpenID: user}

	util.ReturnSuccessJSON(c, convertMenuToDTO(&menu, recipeMap, userMap))
}

// DeleteMenu
// 方法/路径：POST /api/menu/delete?id=xxx
// 鉴权：需 JWT；仅可删除自己家庭的
func DeleteMenu(_ context.Context, c *app.RequestContext) {
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

	var menu model.MenuDBModel
	if err := db.Get().First(&menu, idStr).Error; err != nil {
		util.ReturnErrorJSON(c, "记录不存在", err)
		return
	}

	if menu.FamilyID != user.FamilyID {
		util.ReturnErrorJSON(c, "无权操作", nil)
		return
	}

	if err := db.Get().Delete(&menu).Error; err != nil {
		util.ReturnErrorJSON(c, "删除失败", err)
		return
	}

	util.ReturnSuccessJSON(c, "已删除")
}
