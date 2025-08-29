package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"plane_war/core"
	"plane_war/global"
	"plane_war/internal/api"
	"plane_war/internal/config"
)

type Options struct {
	DB bool
}

func main() {
	// 1. 读取配置
	global.Config = config.InitConf()
	//初始化日志
	global.Log = core.InitLogger()
	//gorm的连接
	global.DB = core.InitGorm(global.Config.Mysql.Dsn)
	//redis连接
	global.Redis = core.InitRedis(global.Config.Redis.Addr, global.Config.Redis.Pwd, global.Config.Redis.DB)
	r := gin.Default()
	// 静态资源
	r.Static("/static", "./static")
	// 默认页面
	r.StaticFile("/", "./static/html/test.html")

	//websocket 路由
	r.GET("/ws", api.WsHandler)

	log.Println("Listening on localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("server run failed:", err)
	}
}
