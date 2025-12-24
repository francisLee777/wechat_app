package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
	"strings"
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

	gormDB, _ := gorm.Open(mysql.Open("root:745620@(127.0.0.1:3306)/francis?charset=utf8mb4&parseTime=True&loc=Local"))
	// reuse your gorm db
	g.UseDB(gormDB)
	// Generate struct `User` based on table `users`
	nameList := []string{"user_info", "family", "message", "recipe", "menu"}
	genList := make([]interface{}, 0)
	for _, table := range nameList {
		genList = append(genList, g.GenerateModelAs(table, SnakeToCamel(table)+"DBModel"))
	}
	g.ApplyBasic(genList...)
	// Generate the code
	g.Execute()
}

// SnakeToCamel 把 snake_case 转成 PascalCase
func SnakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
	}
	return strings.Join(parts, "")
}
