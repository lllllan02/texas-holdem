package wscore

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// Upgrader 是默认的 WebSocket 升级器。
// 业务层可以根据需要修改其属性（例如 CheckOrigin）。
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // 默认允许跨域
}

// ServeWS 处理 HTTP 请求，将其升级为 WebSocket 连接，并注册到 Hub。
// clientID: 客户端的唯一标识（由业务层解析和提供）
// context: 附加的强类型业务上下文数据
func ServeWS(hub *Hub, handler MessageHandler, clientID string, context any, w http.ResponseWriter, r *http.Request) {
	// 1. 升级 HTTP 连接为 WebSocket
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket Upgrade error:", err)
		return
	}

	// 2. 创建通用的 ws.Client
	c := NewClient(clientID, conn, hub, handler)
	c.context = context

	// 3. 注册到 Hub
	select {
	case hub.register <- c:
	case <-hub.destroy:
		// 如果 Hub 已经停止，直接关闭连接
		conn.Close()
		return
	}

	// 4. 启动读写 Goroutine
	go c.WritePump()
	go c.ReadPump()
}
