package core

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"time"
)

func InitMysql() *gorm.DB {

	dsn := "root:123123@tcp(127.0.0.1:3306)/plane_war?charset=utf8mb4&parseTime=True&loc=Local"

	var mysqlLogger logger.Interface
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: mysqlLogger,
	})
	if err != nil {
		log.Fatalf("[%s] mysql连接失败: %v", dsn, err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)               // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)              // 最多可容纳
	sqlDB.SetConnMaxLifetime(time.Hour * 4) // 连接最大复用时间，不能超过mysql的wait_timeout
	return db
}

func InitGorm(MysqlDataSource string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(MysqlDataSource), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 开启 Info 日志
	})
	if err != nil {
		panic("连接mysql数据库失败, error=" + err.Error())
	}
	fmt.Println("连接mysql数据库成功")
	return db
}
