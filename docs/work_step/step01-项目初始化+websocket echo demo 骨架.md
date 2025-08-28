### 目录结构

airplane-war/
├── cmd/
│   └── server/
│       └── main.go       # 入口
└── internal/
├── api/
│   └── ws.go         # WebSocket 接口
└── ws/
└── hub.go        # 连接管理（echo用）


### 前置操作
    go mod init
    go get github.com/gin-gonic/gin
    go get github.com/gorilla/websocket


---

# 🌟 阶段一思路概览

阶段目标：

* **建立 WebSocket 连接**
* **能够发送消息**
* **服务端能原样返回消息（Echo）**
* **前端可以看到发送和接收的日志**

核心是 **WebSocket 基础通信**，为后续的匹配和房间功能打基础。

---

## 1️⃣ 后端部分（Go + Gin + WebSocket）

### 1.1 `cmd/server/main.go`

* **作用**：程序入口，启动 Gin HTTP 服务
* **关键点**：

  ```go
  r := gin.Default()
  r.GET("/ws", api.WsHandler)
  r.Run(":8080")
  ```

    * Gin 负责 HTTP 路由
    * `/ws` 路由交给 `api.WsHandler` 处理 WebSocket
    * 启动 HTTP 服务监听 8080 端口

### 1.2 `internal/api/ws.go`

* **作用**：将 HTTP 请求升级为 WebSocket

* **关键点**：

  ```go
  upgrader := websocket.Upgrader{ CheckOrigin: func(r *http.Request) bool { return true } }
  conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)
  ```

    * `Upgrader` 将普通 HTTP 升级为 WebSocket
    * `CheckOrigin` 返回 true 允许跨域（demo 简化处理）

* 创建客户端：

  ```go
  client := ws.NewClient(conn)
  go client.ReadPump()
  go client.WritePump()
  ws.HubInstance.Register <- client
  ```

    * 每个客户端有 **自己的读写 goroutine**
    * 将客户端注册到 **Hub** 管理

### 1.3 `internal/ws/hub.go`

* **作用**：管理所有客户端连接，实现消息广播

* **核心结构**：

  ```go
  type Hub struct {
      Clients map[*Client]bool
      Broadcast chan []byte
      Register chan *Client
      Unregister chan *Client
  }
  ```

    * `Clients`：所有在线客户端
    * `Broadcast`：消息广播管道
    * `Register` / `Unregister`：注册/注销客户端

* **核心逻辑**：

  ```go
  func (h *Hub) Run() {
      for {
          select {
          case client := <-h.Register: ...        // 新连接
          case client := <-h.Unregister: ...      // 断开连接
          case message := <-h.Broadcast: ...      // 广播消息
          }
      }
  }
  ```

    * `Run` 在独立 goroutine 持续运行
    * 接收到消息后，遍历 `Clients` 发给每个客户端

* **客户端读写**：

  ```go
  func (c *Client) ReadPump() { ... }  // 读消息
  func (c *Client) WritePump() { ... } // 写消息
  ```

    * `ReadPump`：监听客户端发来的消息 → 放入 `Hub.Broadcast`
    * `WritePump`：从 `Send` channel 取消息 → 发给客户端

✅ 这样实现了 **echo 功能**：客户端发消息 → 服务器广播 → 收到消息

---

## 2️⃣ 前端部分（HTML + JS）

### 2.1 页面结构

* 输入框 + 发送按钮
* 消息日志区域（显示发送/接收消息）

### 2.2 WebSocket 核心逻辑

```javascript
ws = new WebSocket("ws://localhost:8080/ws");

ws.onopen = () => log("✅ 已连接服务器");
ws.onmessage = (evt) => log("📩 收到: " + evt.data, "msg-recv");
ws.onclose = () => log("❌ 连接已关闭");
ws.onerror = (err) => log("⚠️ 出错: " + err);
```

* 建立连接
* 注册事件回调：

    * `onopen` → 连接成功
    * `onmessage` → 收到服务器消息
    * `onclose` → 连接关闭
    * `onerror` → 出现错误

### 2.3 发送消息

```javascript
function sendMsg() {
    let msg = document.getElementById("msgInput").value;
    ws.send(msg);
    log("📤 发送: " + msg, "msg-send");
}
```

* 通过 `ws.send` 将输入框消息发送给服务端
* 在日志区显示发送消息

### 2.4 日志显示

```javascript
function log(message, type="msg-sys") {
    let p = document.createElement("p");
    p.className = type;
    p.textContent = message;
    logDiv.appendChild(p);
    logDiv.scrollTop = logDiv.scrollHeight;
}
```

* 统一显示发送、接收、系统消息
* 滚动条自动滚动到底部

---

## 3️⃣ 流程总结

1. 前端打开浏览器 → `test.html`
2. JS 自动执行 `new WebSocket("ws://localhost:8080/ws")`
3. Gin 后端 `/ws` 路由升级为 WebSocket
4. 后端创建 `Client` 对象 → 注册到 `Hub`
5. 前端发送消息 → 后端 `ReadPump` 接收 → 放到 `Hub.Broadcast`
6. `Hub.Run` 遍历在线客户端 → 通过 `WritePump` 发回前端
7. 前端 `onmessage` 收到消息 → 日志显示

---

📌 **阶段一核心思路**：

* 先把 **通信通道搭起来**（WebSocket + Hub + Client）
* 前端简单测试 → 确认收发消息正常
* 后续阶段（匹配、房间、游戏循环）都基于这个通信框架


