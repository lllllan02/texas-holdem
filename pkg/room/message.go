package room

import (
	"encoding/json"
)

// 系统级 Action 定义
const (
	// ActionJoin 玩家加入房间的广播消息
	ActionJoin = "room.join"
	// ActionLeave 玩家离开房间的广播消息，或客户端主动发送的离开请求
	ActionLeave = "room.leave"
	// ActionChat 房间内的公共聊天消息
	ActionChat = "room.chat"
	// ActionRoomState 房间整体状态更新（如游戏开始、结束等）
	ActionRoomState = "room.state"
	// ActionSyncState 玩家断线重连或初次加入时，同步当前游戏状态的单播消息
	ActionSyncState = "room.sync_state"
	// ActionKick 玩家被踢出房间（如顶号登录）的单播消息
	ActionKick = "room.kick"
	// ActionHostChanged 房主发生转移时的广播消息
	ActionHostChanged = "room.host_changed"
)

// ClientMessage 客户端发送的消息结构
type ClientMessage struct {
	Action  string `json:"action"`
	Content string `json:"content"`
}

// ServerMessage 服务端下发的消息结构
type ServerMessage struct {
	Action   string `json:"action"`
	Content  string `json:"content"`
	SenderID string `json:"senderId,omitempty"` // 如果是系统消息，则为空
}

// BuildServerMessage 辅助函数：构建标准服务端消息
func BuildServerMessage(senderID string, action string, content string) []byte {
	msg := ServerMessage{
		Action:   action,
		Content:  content,
		SenderID: senderID,
	}
	bytes, _ := json.Marshal(msg)
	return bytes
}
