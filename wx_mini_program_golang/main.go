package main

import (
	"fmt"
	"wxcloudrun-golang/conf"
	"wxcloudrun-golang/consts"
	"wxcloudrun-golang/db"
	"wxcloudrun-golang/router"
	"wxcloudrun-golang/util"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func main() {
	// Set log level to Info
	hlog.SetLevel(hlog.LevelInfo)

	if err := db.Init(); err != nil {
		panic(fmt.Sprintf("mysql init failed with %+v", err))
	}
	if err := conf.Init(); err != nil {
		panic(fmt.Sprintf("conf init failed with %+v", err))
	}

	// Create uploads directory if not exists
	if err := util.BizOsMkdirAll(consts.DataPath); err != nil {
		panic(fmt.Sprintf("failed to create upload directory: %v", err))
	}

	h := server.Default(server.WithHostPorts(":8080"))

	router.Register(h)

	h.Spin()
}
