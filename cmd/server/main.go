// ./cmd/server/main.go
package main

import (
	"log"
	"plane_war/internal/config"
	core2 "plane_war/internal/core"
	"plane_war/internal/global"
	"plane_war/internal/router"
)

type Options struct {
	DB bool
}

func main() {
	// 1. 读取配置
	global.Config = config.InitConf()
	//初始化日志
	global.Log = core2.InitLogger()
	//gorm的连接
	global.DB = core2.InitGorm(global.Config.Mysql.Dsn)
	//redis连接
	global.Redis = core2.InitRedis(global.Config.Redis.Addr, global.Config.Redis.Pwd, global.Config.Redis.DB)
	r := router.InitRouter()
	global.Log.Info(global.Config.Server.Host + global.Config.Server.Port)
	if err := r.Run(global.Config.Server.Port); err != nil {
		log.Fatal("server run failed:", err)
	}
}
