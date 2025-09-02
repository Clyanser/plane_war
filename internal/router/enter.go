package router

import (
	"github.com/gin-gonic/gin"
	"plane_war/internal/api"
	"plane_war/internal/global"
)

type RouterGroup struct {
	*gin.RouterGroup
}

func InitRouter() *gin.Engine {
	gin.SetMode(global.Config.Server.Env)
	r := gin.Default()
	// 静态资源
	r.Static("/static", "./static")
	// 默认页面
	r.StaticFile("/", "./static/html/test.html")
	// 路由分组
	apiRouterGroup := r.Group("api")
	//路由分成
	routerGroupApp := RouterGroup{apiRouterGroup}
	routerGroupApp.AuthRouter()
	// WebSocket 路由
	r.GET("/ws", api.WsHandler) // WebSocket 路由
	return r
}
