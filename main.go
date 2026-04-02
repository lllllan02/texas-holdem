package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/lllllan02/texas-holdem/pkg/handler"
)

func main() {
	// 设置 Gin 模式 (ReleaseMode / DebugMode)
	gin.SetMode(gin.DebugMode)

	// 初始化 Gin 引擎
	r := gin.Default()

	// 注册跨域中间件（如果是前后端分离项目需要）
	r.Use(corsMiddleware())

	// 配置文件上传目录的静态文件服务
	r.Static("/uploads", "./uploads")

	// 注册业务路由和 WebSocket 接口
	handler.RegisterRoutes(r)

	// 启动 HTTP 服务器
	port := ":8080"
	log.Printf("Starting server on %s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// corsMiddleware 简单的跨域中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
