package models

import (
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
)

var ChatDB *gorm.DB
var err error // 不提前定义一个err，连接mysql的err不好处理

func InitDB() {
	dsn := viper.GetString("mysql.dsn")
	ChatDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
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
	return
}
