### ✅ 阶段一目标回顾

功能点
注册用户（写入 MySQL）
    登录（校验 + 返回 token，Redis 缓存 session）
    大厅接口（返回房间列表、排行榜简要信息）
技术点
    GORM 连接 MySQL
    Redis 存储用户 session
    API 路由（HTTP）
    数据模型定义（User）
    登录态校验中间件

### 项目目录
game-server/
├── cmd/server/main.go
├── internal/
│   ├── api/
│   │   ├── ws.go
│   │   ├── user.go      # 注册/登录接口
│   │   └── lobby.go     # 大厅接口
│   ├── model/
│   │   ├── user.go      # 用户表模型
│   │   ├── player.go
│   │   └── room.go
│   ├── service/
│   │   ├── db.go        # GORM 初始化
│   │   ├── redis.go     # Redis 初始化
│   │   ├── match/match.go
│   │   └── game/game.go
│   └── ws/hub.go
└── static/html/test.html
