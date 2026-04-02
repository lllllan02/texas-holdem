package handler

import "github.com/lllllan02/texas-holdem/pkg/user"

// 房间级消息类型 (Room MsgType)
const (
	MsgTypePauseGame   = "room.pause"     // 房主暂停游戏
	MsgTypeResumeGame  = "room.resume"    // 房主恢复游戏
	MsgTypeWelcome     = "room.welcome"   // 欢迎加入房间（返回分配的 UserID 等）
	MsgTypeGamePaused  = "room.paused"    // 游戏已暂停
	MsgTypeRoomDestroy = "room.destroyed" // 房间已解散
)

// 房间级消息 Payload 结构体

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
