# Demo 04: WebSocket 实时通信基础 (WebSocket Basics)

## 目标
脱离纯算法，搭建前后端双向通信的桥梁。验证 Go 服务器能够接收浏览器的消息，并能主动向浏览器推送消息。

## 目录结构建议
```text
poker/
├── server/
│   ├── main.go      # HTTP/WS 服务入口
│   ├── client.go    # 封装单个 WebSocket 连接
│   └── hub.go       # 广播中心，管理所有连接
├── web-demo/
│   └── index.html   # 极简的原生 HTML 测试页
```

## 核心数据结构设计 (Go)
```go
// 单个客户端连接
type Client struct {
    Conn *websocket.Conn
    Send chan []byte // 待发送消息的缓冲通道
}

// 消息中心
type Hub struct {
    Clients    map[*Client]bool
    Broadcast  chan []byte
    Register   chan *Client
    Unregister chan *Client
}
```

## 核心方法签名
- `func (h *Hub) Run()`：在一个独立的 Goroutine 中运行，处理注册、注销和广播事件。
- `func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request)`：将普通的 HTTP 请求升级为 WebSocket 连接。

## 具体测试流程 (手动验证)
1. 启动 Go 服务 `go run server/*.go`，监听 `localhost:8080`。
2. 在浏览器中打开 `web-demo/index.html`（里面写一段原生的 `new WebSocket('ws://localhost:8080/ws')` 代码）。
3. **验证连通性**: 浏览器控制台打印 "Connected"。
4. **验证广播**: 打开两个浏览器标签页 A 和 B。在 A 的输入框发消息，B 的页面上能立刻显示出来（类似聊天室）。
5. **验证断开**: 关闭标签页 A，Go 后端终端打印 "Client disconnected"，且不会引发 panic。