package server

import (
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lllllan02/texas-holdem/pkg/room"
	"github.com/lllllan02/texas-holdem/pkg/wscore"
)

// APIHandler 封装了 HTTP 接口和 WebSocket 接入点
type APIHandler struct {
	manager *room.RoomManager
}

// NewAPIHandler 创建 API 处理器
func NewAPIHandler(manager *room.RoomManager) *APIHandler {
	return &APIHandler{
		manager: manager,
	}
}

// generateRoomID 生成 6 位数字房间号
func generateRoomID() string {
	rand.Seed(time.Now().UnixNano())
	return strconv.Itoa(100000 + rand.Intn(900000))
}

// HandleCreateRoom 处理创建房间的 HTTP 请求
func (h *APIHandler) HandleCreateRoom(c *gin.Context) {
	var req struct {
		UserID   string `json:"userId" binding:"required"`
		GameType string `json:"gameType" binding:"required"` // 指定游戏类型
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	roomID := generateRoomID()
	rm, err := h.manager.CreateRoom(roomID, req.UserID, req.GameType, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"roomId": rm.GetID()})
}

// HandleCheckRoom 处理检查房间是否存在的 HTTP 请求
func (h *APIHandler) HandleCheckRoom(c *gin.Context) {
	roomID := c.Query("roomId")
	rm, exists := h.manager.GetRoom(roomID)

	if exists {
		c.JSON(http.StatusOK, gin.H{
			"exists": true,
			"hostId": rm.GetHostID(),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{"exists": false})
	}
}

// HandleDismissRoom 处理房主解散房间的 HTTP 请求
func (h *APIHandler) HandleDismissRoom(c *gin.Context) {
	var req struct {
		RoomID string `json:"roomId" binding:"required"`
		UserID string `json:"userId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.manager.DismissRoom(req.RoomID, req.UserID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ServeWS 处理 WebSocket 连接请求
func (h *APIHandler) ServeWS(c *gin.Context) {
	// 1. 获取 URL 参数中的 roomID 和 playerID(userId) 以及 username
	roomID := c.Query("room")
	userID := c.Query("userId")
	username := c.Query("username")

	if roomID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing room or userId parameter"})
		return
	}

	if username == "" {
		username = "游客_" + userID[:4]
	}

	// 2. 检查房间是否存在 (必须先通过 HTTP 接口创建)
	rm, exists := h.manager.GetRoom(roomID)
	if !exists {
		log.Printf("玩家 [%s] 尝试加入不存在的房间 [%s]\n", userID, roomID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// 3. 使用 wscore 框架处理 WebSocket 连接
	wscore.ServeWS(rm.GetHub(), rm, userID, nil, c.Writer, c.Request)

	log.Printf("玩家 [%s](%s) 成功连接到房间 [%s] 的 WebSocket\n", username, userID, roomID)
}
