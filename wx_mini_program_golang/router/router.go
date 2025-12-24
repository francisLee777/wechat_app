package router

import (
	"wxcloudrun-golang/consts"
	"wxcloudrun-golang/handler"
	"wxcloudrun-golang/middleware"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func Register(h *server.Hertz) {

	// Register Middleware
	h.Use(middleware.AccessLog())
	h.Use(middleware.CharsetUTF8())

	// 用途：微信登录
	h.Any("/api/user/login", handler.WechatLogin)

	protected := h.Group("/api")
	protected.Use(middleware.JWTAuth())

	// 用途：访问家庭图片资源
	// 鉴权：JWT 或 ?token=；家庭匹配
	// 返回：图片二进制
	protected.GET(consts.FamilyPrefix+"/:familyId/:fileName", handler.ServeFamilyPNG)

	// GET /api/media/user_png/:openId/:fileName
	// 用途：访问个人头像资源（未加入家庭前）
	// 鉴权：JWT 或 ?token=；仅本人可访问
	// 返回：图片二进制
	protected.GET(consts.UserPrefix+"/:openId/:fileName", handler.ServeUserPNG)

	// 原始客户端url是 /assets/templates/fileName
	// 匹配到的是 /templates/fileName
	protected.GET(consts.AssetFilePrefix+"/*filepath", handler.GetAssetImage)

	// POST /api/upload
	// 用途：上传图片到家庭目录
	// 鉴权：JWT；需已加入家庭
	// 请求：multipart/form-data file
	// 返回：{ url }
	protected.POST("/upload", handler.UploadFile)

	// [已移除] 订单相关接口

	// GET /api/user/getUserInfo
	// 用途：获取当前用户信息
	// 鉴权：JWT
	// 返回：UserInfoDBModel
	protected.Any("/user/getUserInfo", handler.GetUserInfo)

	// POST /api/user/saveNickName
	// 用途：保存昵称
	// 鉴权：JWT
	// 请求：{ nickName }
	// 返回：{ nickName }
	protected.Any("/user/saveNickName", handler.SaveNickName)

	// POST /api/user/saveIconURL
	// 用途：保存头像 URL
	// 鉴权：JWT
	// 请求：Query iconURL
	// 返回：string
	protected.Any("/user/saveIconURL", handler.SaveIconURL)

	// POST /api/family/create
	// 用途：创建家庭（设为家长，写默认留言/示例食谱）
	// 鉴权：JWT；未加入家庭
	// 请求：{ name }
	// 返回：FamilyDBModel
	protected.Any("/family/create", handler.CreateFamily)

	// POST /api/family/join
	// 用途：加入家庭
	// 鉴权：JWT；最多 4 人
	// 请求：{ familyId }
	// 返回：FamilyDBModel
	protected.Any("/family/join", handler.JoinFamily)

	// GET /api/family/members
	// 用途：获取家庭成员
	// 鉴权：JWT
	// 返回：UserInfoDBModel[]
	protected.Any("/family/members", handler.GetFamilyMembers)

	// POST /api/family/quit
	// 用途：退出家庭
	// 鉴权：JWT；家长不可退出
	// 返回：字符串
	protected.Any("/family/quit", handler.QuitFamily)

	// POST /api/family/removeMember
	// 用途：家长移除成员
	// 鉴权：JWT；admin
	// 请求：{ memberOpenId }
	// 返回：字符串
	protected.Any("/family/removeMember", handler.RemoveMember)

	// POST /api/family/delete
	// 用途：解散家庭（删除所有数据）
	// 鉴权：JWT；admin
	// 返回：字符串
	protected.Any("/family/delete", handler.DeleteFamily)

	// GET /api/family/info?id=xxx
	// 用途：获取家庭信息
	// 鉴权：JWT；同家庭
	// 返回：FamilyDBModel
	protected.Any("/family/info", handler.GetFamilyInfo)

	// POST /api/recipe/add
	// 用途：新增食谱
	// 鉴权：JWT；admin
	// 请求：{ name, images[], content }
	// 返回：RecipeDTO
	protected.Any("/recipe/add", handler.AddRecipe)

	// GET /api/recipe/list
	// 用途：食谱列表（按 sort_order 降序）
	// 鉴权：JWT
	// 返回：RecipeDTO[]
	protected.Any("/recipe/list", handler.GetRecipes)

	// GET /api/recipe/info?id=xxx
	// 用途：食谱详情
	// 鉴权：JWT；同家庭
	// 返回：RecipeDTO
	protected.Any("/recipe/info", handler.GetRecipe)

	// POST /api/recipe/update
	// 用途：更新食谱
	// 鉴权：JWT；admin
	// 请求：{ id, name, images[], content }
	// 返回：RecipeDTO
	protected.Any("/recipe/update", handler.UpdateRecipe)

	// POST /api/recipe/delete?id=xxx
	// 用途：删除食谱
	// 鉴权：JWT；admin
	// 返回：字符串
	protected.Any("/recipe/delete", handler.DeleteRecipe)

	// POST /api/recipe/reorder
	// 用途：调整顺序（上下移动）
	// 鉴权：JWT；admin
	// 请求：{ id, direction }
	// 返回：字符串
	protected.Any("/recipe/reorder", handler.ReorderRecipe)

	// POST /api/recipe/batchUpdate
	// 用途：批量更新排序
	// 鉴权：JWT；admin
	// 请求：{ recipes: [{ id, sortOrder }] }
	// 返回：字符串
	protected.Any("/recipe/batchUpdate", handler.BatchUpdateRecipes)

	// GET /api/recipe/templates
	// 用途：获取系统推荐菜谱模版
	// 鉴权：JWT
	// 返回：RecipeTemplate[]
	protected.Any("/recipe/templates", handler.GetRecipeTemplates)

	// POST /api/message/add
	// 用途：新增留言/回复
	// 鉴权：JWT；已加入家庭
	// 请求：{ content, parentId?, replyToUserId?, replyToUserName? }
	// 返回：MessageDBModel
	protected.Any("/message/add", handler.AddMessage)

	// GET /api/message/list
	// 用途：留言列表（根倒序、回复正序）
	// 鉴权：JWT
	// 返回：{ ...message, replies: [] }[]
	protected.Any("/message/list", handler.GetMessages)

	// POST /api/message/delete?id=xxx
	// 用途：删除留言（根删除级联）
	// 鉴权：JWT；家长或本人
	// 返回：字符串
	protected.Any("/message/delete", handler.DeleteMessage)

	// GET /api/menu/list?date=YYYY-MM-DD
	// 用途：获取某日菜单
	// 鉴权：JWT
	// 返回：{ breakfast: [], lunch: [], dinner: [] }
	protected.Any("/menu/list", handler.GetMenuList)

	// POST /api/menu/add
	// 用途：添加菜品
	// 鉴权：JWT
	// 请求：{ date, mealType, recipeId, remark }
	// 返回：MenuDTO
	protected.Any("/menu/add", handler.AddMenu)

	// POST /api/menu/delete?id=xxx
	// 用途：删除菜品
	// 鉴权：JWT
	// 返回：字符串
	protected.Any("/menu/delete", handler.DeleteMenu)
}
