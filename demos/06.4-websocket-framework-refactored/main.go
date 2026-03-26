package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/lllllan02/texas-holdem/demos/06.4-websocket-framework-refactored/games/texas"
	"github.com/lllllan02/texas-holdem/demos/06.4-websocket-framework-refactored/server"
	"github.com/lllllan02/texas-holdem/pkg/room"
)

func main() {
	// 1. 创建全局房间管理器
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	manager := room.NewRoomManager()

	// 注册德州扑克游戏引擎
	manager.RegisterEngine("texas", func() room.GameEngine {
		return texas.NewTexasEngine()
	})

	// 2. 创建 API 处理器
	apiHandler := server.NewAPIHandler(manager)

	// 3. 初始化 Gin 引擎
	r := gin.Default()

	// 4. 设置路由
	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	// API 路由组
	api := r.Group("/api")
	{
		api.POST("/create_room", apiHandler.HandleCreateRoom)
		api.GET("/check_room", apiHandler.HandleCheckRoom)
		api.POST("/dismiss_room", apiHandler.HandleDismissRoom)
	}

	// 提供 WebSocket 接入点
	r.GET("/ws", apiHandler.ServeWS)

	// 5. 启动服务器
	port := ":8082"
	log.Printf("服务器启动成功，监听端口 %s\n", port)
	log.Printf("请在浏览器访问: http://localhost%s\n", port)

	err := r.Run(port)
	if err != nil {
		log.Println("Run: ", err)
	}
}
