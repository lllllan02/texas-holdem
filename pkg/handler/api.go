package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/lllllan02/texas-holdem/pkg/game/texas"
	"github.com/lllllan02/texas-holdem/pkg/user"
	"github.com/lllllan02/texas-holdem/pkg/wscore"
)

// RoomManager 管理所有活跃的房间
type RoomManager struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

var globalRoomManager = &RoomManager{
	rooms: make(map[string]*Room),
}

// RegisterRoutes 注册所有的 HTTP 路由
func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		api.POST("/rooms", createRoom)
		
		// 用户相关接口
		api.GET("/users/:id", getUser)
		api.PUT("/users/:id", updateUser)
	}

	// WebSocket 升级接口
	r.GET("/ws/:room_number", serveWs)
}

// CreateRoomRequest 创建房间的请求体
type CreateRoomRequest struct {
	OwnerID string `json:"owner_id" binding:"required"`
	Options any    `json:"options"` // 前端直接传入引擎需要的参数，透传给引擎
}

// CreateRoomResponse 创建房间的响应体
type CreateRoomResponse struct {
	RoomID     string `json:"room_id"`
	RoomNumber string `json:"room_number"`
}

// createRoom 处理创建房间的 HTTP 请求
func createRoom(c *gin.Context) {
	var req CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 生成唯一的 RoomID 和简单的 6 位数字房号（为了演示，这里直接截取 UUID 前 6 位）
	roomID := uuid.New().String()
	roomNumber := roomID[:6]

	// 创建德州扑克引擎实例
	engine := texas.NewTable()

	// 将 Options 转换为 []byte 透传给引擎
	var optionsBytes []byte
	if req.Options != nil {
		var err error
		optionsBytes, err = json.Marshal(req.Options)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid options format"})
			return
		}
	} else {
		optionsBytes = []byte(`{}`)
	}

	// 创建房间
	room, err := NewRoom(roomID, roomNumber, req.OwnerID, engine, optionsBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create room: " + err.Error()})
		return
	}

	// 保存到全局管理器
	globalRoomManager.mu.Lock()
	globalRoomManager.rooms[roomNumber] = room
	globalRoomManager.mu.Unlock()

	log.Printf("Room created: [%s] Number: %s, Owner: %s", roomID, roomNumber, req.OwnerID)

	c.JSON(http.StatusOK, CreateRoomResponse{
		RoomID:     roomID,
		RoomNumber: roomNumber,
	})
}

// 用户相关接口

// getUser 获取用户信息
func getUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id is required"})
		return
	}

	u := user.GetUserByID(userID)
	c.JSON(http.StatusOK, u)
}

// UpdateUserRequest 更新用户信息的请求体
type UpdateUserRequest struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// updateUser 更新用户信息
func updateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id is required"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u := user.UpdateUser(userID, req.Nickname, req.Avatar)

	// TODO: 如果用户正在某个房间内，可以考虑在这里触发一条房间内的状态更新广播
	// 目前简单起见，仅更新数据库（内存缓存），下次用户重连或刷新时生效

	c.JSON(http.StatusOK, u)
}

// WebSocket 接口

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许跨域
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// serveWs 处理 WebSocket 升级请求
func serveWs(c *gin.Context) {
	roomNumber := c.Param("room_number")
	userID := c.Query("user_id") // 简单起见，从 query 参数获取 userID

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	// 查找房间
	globalRoomManager.mu.RLock()
	room, exists := globalRoomManager.rooms[roomNumber]
	globalRoomManager.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}

	// 升级连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	// 创建 WebSocket 客户端，并将 room 作为 MessageHandler 注入
	client := wscore.NewClient(userID, conn, room.hub, room)

	// 注册到 Hub
	room.hub.Register(client)

	// 启动读写 Goroutine
	go client.WritePump()
	go client.ReadPump()
}
