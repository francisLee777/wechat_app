package handler

import (
	"fmt"
	"github.com/bytedance/sonic"
	"gorm.io/gorm/clause"
	"net/http"
	"wxcloudrun-golang/db"
	"wxcloudrun-golang/db/model"
	"wxcloudrun-golang/util"
)

type FoodOrderItem struct {
	FoodMenuItem
	Count int `json:"count"` // 数量
}

type FoodOrderItemReq struct {
	FoodList     []*FoodOrderItem `json:"food_list"`
	OpenId       string           `json:"open_id"`       // 用户信息
	TargetPeriod string           `json:"target_period"` // 订餐的目标时间，取值只有9个，前端已校验， 明天中午、后天早上、今天晚上
	Extra        string           `json:"extra"`         // 订单备注
}

type OrderQueryResp struct {
	OrderList []*OrderItem `json:"order_list"`
}

type OrderItem struct {
	FoodList     []*FoodOrderItem `json:"food_list"`
	TargetPeriod string           `json:"target_period"`
	Extra        string           `json:"extra"` // 订单备注
}

// OrderFood 提交点餐订单
func OrderFood(w http.ResponseWriter, r *http.Request) {
	req := &FoodOrderItemReq{}
	err := util.BindJson(r, req)
	if err != nil {
		_, _ = fmt.Fprint(w, "绑定错误", err)
		return
	}
	openId, err := util.GetOpenIdFromHeader(r)
	if err != nil {
		_, _ = fmt.Fprint(w, "没有登录", err)
		return
	}
	req.OpenId = openId
	q1 := db.DB.OrderDBModel
	// 冲突的话只更新订单内容
	if err = q1.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{q1.Info.ColumnName().String(), q1.Extra.ColumnName().String()})}).
		Create(&model.OrderDBModel{
			OpenID:       openId,
			UniqueCode:   util.GenerateUUID(),
			TargetPeriod: req.TargetPeriod,
			Info:         util.Convert2JSONString(req.FoodList),
			Extra:        req.Extra,
		}); err != nil {
		_, _ = fmt.Fprint(w, "数据库错误", err)
		return
	}
	util.ReturnSuccessJSON(w, "成功保存/覆写订单")
}

// GetMyOrder 获取已有订单
func GetMyOrder(w http.ResponseWriter, r *http.Request) {
	openId, err := util.GetOpenIdFromHeader(r)
	if err != nil {
		_, _ = fmt.Fprint(w, "没有登录", err)
		return
	}
	q1 := db.DB.OrderDBModel
	orderModelList, err := q1.Where(q1.OpenID.Eq(openId)).Find()
	if err != nil {
		return
	}
	resp := OrderQueryResp{}
	for _, pojo := range orderModelList {
		var itemList []*FoodOrderItem
		_ = sonic.UnmarshalString(pojo.Info, &itemList)
		resp.OrderList = append(resp.OrderList, &OrderItem{
			FoodList:     itemList,
			TargetPeriod: pojo.TargetPeriod,
			Extra:        pojo.Extra,
		})
	}
	util.ReturnSuccessJSON(w, resp)
}
