package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/lllllan02/texas-holdem/demos/06.2-room-management-decoupled/core"
	"github.com/lllllan02/texas-holdem/demos/06.2-room-management-decoupled/games/texas"
)

func main() {
	// 1. 创建全局房间管理器
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	manager := core.NewRoomManager()

	// 注册德州扑克游戏引擎
	manager.RegisterEngine("texas", func() core.GameEngine {
		return texas.NewTexasEngine()
	})

	// 未来如果开发了其他游戏，可以在这里继续注册：
	// manager.RegisterEngine("uno", func(roomID string) core.GameEngine {
	// 	return uno.NewUnoEngine()
	// })

	// 2. 初始化 Gin 引擎
	r := gin.Default()

	// 3. 设置路由
	// 提供静态文件服务 (前端页面)
	// 注意：在 Gin 中，如果使用 r.Static("/", "./static")，它会捕获所有根路径下的请求，
	// 这会导致与后面的 /api 和 /ws 路由产生冲突。
	// 解决方案：使用 StaticFS 并指定一个具体的静态目录，或者手动处理根路径。
	// 这里我们手动处理根路径返回 index.html，并将 static 目录下的其他文件挂载到 /static
	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})
	// 如果 static 目录下有其他资源文件（如 css/js），可以通过这个路由访问
	// r.Static("/static", "./static")

	// API 路由组
	api := r.Group("/api")
	{
		api.POST("/create_room", manager.HandleCreateRoom)
		api.GET("/check_room", manager.HandleCheckRoom)
		api.POST("/dismiss_room", manager.HandleDismissRoom)
	}

	// 提供 WebSocket 接入点
	r.GET("/ws", manager.ServeWS)

	// 4. 启动服务器
	port := ":8080"
	log.Printf("服务器启动成功，监听端口 %s\n", port)
	log.Printf("请在浏览器访问: http://localhost%s\n", port)

	err := r.Run(port)
	if err != nil {
		log.Println("Run: ", err)
	}
}
