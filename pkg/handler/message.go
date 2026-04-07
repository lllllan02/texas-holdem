package handler

import (
	"github.com/lllllan02/texas-holdem/pkg/core"
	"github.com/lllllan02/texas-holdem/pkg/user"
)

// 房间级消息类型 (Room MsgType)
const (
	MsgTypePauseGame   = "room.pause"     // 房主暂停游戏
	MsgTypeResumeGame  = "room.resume"    // 房主恢复游戏
	MsgTypeWelcome     = "room.welcome"   // 欢迎加入房间（返回分配的 UserID 等）
	MsgTypeGamePaused  = "room.paused"    // 游戏已暂停
	MsgTypeRoomDestroy = "room.destroyed" // 房间已解散
	MsgTypeChat        = "room.chat"      // 房间内聊天消息
	MsgTypePlayerJoin  = "room.player_join"  // 玩家加入房间
	MsgTypePlayerLeave = "room.player_leave" // 玩家离开房间
	MsgTypeHistory     = "room.history"      // 房间历史记录同步
)

// 房间级消息 Payload 结构体

// PlayerEventPayload 玩家进出房间的消息体
type PlayerEventPayload struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}

// ChatPayload 聊天消息体
type ChatPayload struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	Avatar   string `json:"avatar"`
	Message  string `json:"message"`
}

// WelcomePayload 欢迎加入房间的消息体
type WelcomePayload struct {
	RoomID     string     `json:"room_id"`
	RoomNumber string     `json:"room_number"`
	OwnerID    string     `json:"owner_id"`
	User       *user.User `json:"user"`
	GameType   string     `json:"game_type"`
}

// GamePausedPayload 游戏暂停/恢复状态同步的消息体
type GamePausedPayload struct {
	IsPaused bool   `json:"is_paused"`
	UserID   string `json:"user_id"` // 触发暂停/恢复的用户
}

// HistoryEntry 单条历史记录
type HistoryEntry struct {
	Message core.Message `json:"message"`
	Time    int64        `json:"time"` // 毫秒级时间戳
}

// HistoryPayload 房间历史记录同步的消息体
type HistoryPayload struct {
	ChatHistory []HistoryEntry `json:"chat_history"`
	GameHistory []HistoryEntry `json:"game_history"`
}
