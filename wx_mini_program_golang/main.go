package main

import (
	"fmt"
	"log"
	"net/http"
	"wxcloudrun-golang/conf"
	"wxcloudrun-golang/db"
	"wxcloudrun-golang/handler"
)

func main() {
	if err := db.Init(); err != nil {
		panic(fmt.Sprintf("mysql init failed with %+v", err))
	}
	if err := conf.Init(); err != nil {
		panic(fmt.Sprintf("conf init failed with %+v", err))
	}
	http.HandleFunc("/api/listFoodMenu", handler.ListFoodMenu)
	http.HandleFunc("/api/order/orderFood", handler.OrderFood)
	http.HandleFunc("/api/order/GetMyOrder", handler.GetMyOrder)
	http.HandleFunc("/api/user/getUserInfo", handler.GetUserInfo)
	http.HandleFunc("/api/user/saveNickName", handler.SaveNickName)
	http.HandleFunc("/api/user/saveIconURL", handler.SaveIconURL)
	http.HandleFunc("/api/user/login", handler.WechatLogin)

	// Family
	http.HandleFunc("/api/family/create", handler.CreateFamily)
	http.HandleFunc("/api/family/join", handler.JoinFamily)
	http.HandleFunc("/api/family/members", handler.GetFamilyMembers)
	http.HandleFunc("/api/family/quit", handler.QuitFamily)
	http.HandleFunc("/api/family/removeMember", handler.RemoveMember)
	http.HandleFunc("/api/family/info", handler.GetFamilyInfo)

	// Recipe
	http.HandleFunc("/api/recipe/add", handler.AddRecipe)
	http.HandleFunc("/api/recipe/list", handler.GetRecipes)
	http.HandleFunc("/api/recipe/info", handler.GetRecipe)
	http.HandleFunc("/api/recipe/update", handler.UpdateRecipe)
	http.HandleFunc("/api/recipe/delete", handler.DeleteRecipe)
	http.HandleFunc("/api/recipe/reorder", handler.ReorderRecipe)
	http.HandleFunc("/api/recipe/batchUpdate", handler.BatchUpdateRecipes)

	// Message
	http.HandleFunc("/api/message/add", handler.AddMessage)
	http.HandleFunc("/api/message/list", handler.GetMessages)
	http.HandleFunc("/api/message/delete", handler.DeleteMessage)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
