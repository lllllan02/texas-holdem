package core

import (
	"encoding/json"
	"time"
)

// ClientMessage 接收前端发来的 JSON 格式消息
type ClientMessage struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Action   string `json:"action"`
	Content  string `json:"content"`
}

// ServerMessage 后端广播给所有人的 JSON 格式消息
type ServerMessage struct {
	Time     string `json:"time"`
	UserID   string `json:"userId,omitempty"`
	Username string `json:"username,omitempty"`
	Action   string `json:"action"`
	Content  string `json:"content,omitempty"`
}

// 房间级别的通用 Action 定义
const (
	ActionJoin          = "room.join"
	ActionLeave         = "room.leave"
	ActionKick          = "room.kick"
	ActionRoomDismissed = "room.dismissed"
	ActionHostChanged   = "room.host_changed"
	ActionSyncState     = "room.sync_state"
	ActionChat          = "room.chat"
	ActionRoomState     = "room.state"
)

// BuildServerMessage 构建标准的服务端消息字节流
func BuildServerMessage(sender *Client, action string, content string) []byte {
	msg := ServerMessage{
		Time:    time.Now().Format("15:04:05"),
		Action:  action,
		Content: content,
	}

	if sender != nil {
		msg.UserID = sender.GetID()
		msg.Username = sender.GetContext().Username
	} else {
		msg.UserID = "system"
		msg.Username = "系统"
	}

	outBytes, _ := json.Marshal(msg)
	return outBytes
}
