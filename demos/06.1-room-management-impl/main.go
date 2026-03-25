package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/lllllan02/texas-holdem/demos/06.1-room-management-impl/server"
)

func main() {
	// 1. 初始化全局房间管理器
	manager := server.NewRoomManager()

	// 2. 预先创建一个测试房间，方便前端直接连接
	manager.GetOrCreateRoom("888888")

	// 3. 注册 HTTP 路由
	// 3.1 静态文件服务 (前端页面)
	http.Handle("/", http.FileServer(http.Dir("./static")))
	
	// 3.2 WebSocket 升级接口
	http.HandleFunc("/ws", manager.ServeWS)

	// 4. 启动服务器
	port := ":8081"
	fmt.Printf("Demo 06.1 服务器已启动，请访问 http://localhost%s\n", port)
	fmt.Println("你可以打开多个浏览器标签页，使用不同的玩家 ID 加入同一个房间号，测试房间状态同步。")
	
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("服务器启动失败: ", err)
	}
}
