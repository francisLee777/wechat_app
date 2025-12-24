package db

import (
	"fmt"
	"os"
	"strings"
	"time"
	"wxcloudrun-golang/db/dao/gen"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var dbInstance *gorm.DB
var DB *gen.Query

// Init 初始化数据库
func Init() error {
	// 构建连接字符串
	source := "%s:%s@tcp(%s)/%s?readTimeout=1500ms&writeTimeout=1500ms&charset=utf8mb4&loc=Local&parseTime=true"
	user := os.Getenv("DB_USER")
	pwd := os.Getenv("DB_PASSWORD")
	addr := os.Getenv("DB_HOST")
	dataBase := os.Getenv("DB_NAME")
	if dataBase == "" {
		dataBase = "francis"
	}
	// 连接到目标数据库
	source = fmt.Sprintf(source, user, pwd, addr, dataBase)
	fmt.Println("开始初始化数据库:", source)

	db, err := gorm.Open(mysql.Open(source), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		}})
	if err != nil {
		fmt.Println("连接到目标数据库失败:", err)
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Println("获取数据库连接池失败:", err)
		return err
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(100)
	sqlDB.SetMaxOpenConns(200)
	sqlDB.SetConnMaxLifetime(time.Hour)

	dbInstance = db
	DB = gen.Use(db)

	// 执行 init.sql 文件
	initSQLFile := "init.sql"
	sqlBytes, err := os.ReadFile(initSQLFile)
	if err != nil {
		fmt.Printf("读取 init.sql 文件失败: %v\n", err)
		return err
	}

	sqlStr := string(sqlBytes)
	// 改进的 SQL 语句分割逻辑
	sqlStatements := splitSQLStatements(sqlStr)

	// 执行每个 SQL 语句
	for i, stmt := range sqlStatements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue // 跳过空语句和注释
		}
		if err := db.Exec(stmt).Error; err != nil {
			fmt.Printf("执行第 %d 条 SQL 语句失败: %s, 错误: %v\n", i+1, stmt, err)
			return err
		}
	}

	fmt.Println("数据库初始化成功")
	return nil
}

// splitSQLStatements 改进的 SQL 语句分割函数
func splitSQLStatements(sqlStr string) []string {
	// 使用正则表达式或更智能的方式分割 SQL 语句
	// 这里使用简单的分割，但处理了一些特殊情况
	var statements []string
	var currentStmt strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	inComment := false

	for i, r := range sqlStr {
		if inComment {
			if r == '\n' {
				inComment = false
			}
			continue
		}

		if r == '-' && i+1 < len(sqlStr) && sqlStr[i+1] == '-' {
			inComment = true
			continue
		}

		if r == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
		} else if r == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
		}

		if r == ';' && !inSingleQuote && !inDoubleQuote && !inComment {
			statements = append(statements, currentStmt.String())
			currentStmt.Reset()
		} else {
			currentStmt.WriteRune(r)
		}
	}

	// 添加最后一个语句
	if stmt := currentStmt.String(); strings.TrimSpace(stmt) != "" {
		statements = append(statements, stmt)
	}

	return statements
}

// Get ...
func Get() *gorm.DB {
	return dbInstance
}
