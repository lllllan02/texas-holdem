package core

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleCreateRoom 处理创建房间的 HTTP 请求
func (m *RoomManager) HandleCreateRoom(c *gin.Context) {
	var req struct {
		UserID   string `json:"userId" binding:"required"`
		GameType string `json:"gameType" binding:"required"` // 新增：指定游戏类型
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	room, err := m.CreateRoom(req.UserID, req.GameType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"roomId": room.ID})
}

// HandleCheckRoom 处理检查房间是否存在的 HTTP 请求
func (m *RoomManager) HandleCheckRoom(c *gin.Context) {
	roomID := c.Query("roomId")
	room := m.GetRoom(roomID)

	if room != nil {
		c.JSON(http.StatusOK, gin.H{
			"exists": true,
			"hostId": room.HostID,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{"exists": false})
	}
}

// HandleDismissRoom 处理房主解散房间的 HTTP 请求
func (m *RoomManager) HandleDismissRoom(c *gin.Context) {
	var req struct {
		RoomID string `json:"roomId" binding:"required"`
		UserID string `json:"userId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := m.DismissRoom(req.RoomID, req.UserID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
