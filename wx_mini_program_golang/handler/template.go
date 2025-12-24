package handler

import (
	"context"
	"wxcloudrun-golang/consts"
	"wxcloudrun-golang/util"

	"github.com/cloudwego/hertz/pkg/app"
)

type RecipeTemplate struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Images  []string `json:"images"`
	Content string   `json:"content"`
}

// SystemTemplates 预置的系统模版
// 实际场景中，这里的数据可以从数据库读取，或者配置文件读取
var SystemTemplates = []RecipeTemplate{
	{
		ID:   "tpl_001",
		Name: "红烧肉 (模版)",
		Images: []string{
			consts.TemplateFilePrefix + "/hongshaorou.png",
		},
		Content: "① 五花肉切块焯水去腥；\n② 热锅炒糖色，下肉翻炒上色；\n③ 加葱姜、料酒、生抽、老抽、冰糖和水；\n④ 小火慢炖至熟透，收汁出锅。"},
	{
		ID:   "tpl_002",
		Name: "番茄炒蛋 (模版)",
		Images: []string{
			consts.TemplateFilePrefix + "/fanqie_chaodan.png",
		},
		Content: "1. 番茄切块，鸡蛋打散。\n2. 锅热油，炒熟鸡蛋盛出。\n3. 锅留底油，炒软番茄。\n4. 加入鸡蛋，调味出锅。",
	},
	{
		ID:   "tpl_003",
		Name: "炒西蓝花 (模版)",
		Images: []string{
			consts.TemplateFilePrefix + "/xilanhua.png",
		},
		Content: "① 西蓝花掰小朵，焯水备用；\n② 热锅少油，蒜片爆香；\n③ 下西蓝花快速翻炒；\n④ 加盐调味，翻匀出锅。",
	}, {
		ID:   "tpl_004",
		Name: "红烧鱼 (模版)",
		Images: []string{
			consts.TemplateFilePrefix + "/hongshaoyu.png",
		},
		Content: "① 鱼处理干净，擦干水分；\n② 热锅煎至两面金黄；\n③ 加葱姜、料酒、生抽、老抽、糖和水；\n④ 小火焖熟，收汁装盘。",
	},
}

// GetRecipeTemplates
// 方法/路径：GET /api/recipe/templates
// 鉴权：需 JWT (建议)
// 返回：RecipeTemplate[]
func GetRecipeTemplates(_ context.Context, c *app.RequestContext) {
	_, err := util.GetOpenId(c)
	if err != nil {
		util.ReturnErrorJSON(c, "未登录", err)
		return
	}

	util.ReturnSuccessJSON(c, SystemTemplates)
}
