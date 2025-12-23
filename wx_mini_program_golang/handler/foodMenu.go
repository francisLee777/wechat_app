package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"wxcloudrun-golang/util"
)

type FoodMenuItem struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type ListFoodMenuReq struct {
	Age  int    `json:"age"`
	Name string `json:"name"`
}

// ListFoodMenu 获取菜单列表接口
func ListFoodMenu(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ListFoodMenu1", r.URL)
	res := &JsonResult{}
	req := &ListFoodMenuReq{}
	err := util.BindJson(r, req)
	if err != nil {
		fmt.Fprint(w, "绑定错误", err)
		return
	}
	res.Data = r.Header
	msg, err := json.Marshal(res)
	if err != nil {
		fmt.Fprint(w, "内部错误", err)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(msg)
}
