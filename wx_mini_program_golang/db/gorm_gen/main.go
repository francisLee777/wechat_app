package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath:      "db/dao/gen",
		OutFile:      "query_gorm_gen.go",
		ModelPkgPath: "db/model/",
		Mode:         gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
		//if you want to generate index tags from database, set FieldWithIndexTag true
		FieldWithIndexTag: true,
		//if you want to generate type tags from database, set FieldWithTypeTag true
		FieldWithTypeTag: true,
		WithUnitTest:     true,
	})

	gormDB, _ := gorm.Open(mysql.Open("root:745620@(127.0.0.1:3306)/lihaoyu?charset=utf8mb4&parseTime=True&loc=Local"))
	// reuse your gorm db
	g.UseDB(gormDB)
	// Generate struct `User` based on table `users`
	table1 := g.GenerateModelAs("kitchen_order", "OrderDBModel")
	table2 := g.GenerateModelAs("user_info", "UserInfoDBModel")
	g.ApplyBasic(table1, table2)
	// Generate the code
	g.Execute()
}
