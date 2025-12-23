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
	log.Fatal(http.ListenAndServe(":8080", nil))
}
