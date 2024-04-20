package models

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
)

var ChatDB = InitDB()
var err error // 不提前定义一个err，连接mysql的err不好处理

func InitDB() *gorm.DB {
	//dsn := viper.GetString("mysql.dsn")
	DB, err := gorm.Open(mysql.Open("root:123456@tcp(127.0.0.1:3306)/go_gin_chat?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // 日志记录器
			logger.Config{
				LogLevel: logger.Info, // 日志级别为 Info，输出每一句 SQL 语句
			},
		),
	})
	if err != nil {
		log.Fatal("connect database err:", err)
	}
	return DB
}
