package handler

import (
	"github.com/lllllan02/texas-holdem/pkg/core"
	"github.com/lllllan02/texas-holdem/pkg/user"
)

// Room 通用房间：最外层的容器，负责网络连接、聊天、用户进出
// 房间本身不关心具体玩什么游戏，它只负责挂载一个具体的 GameEngine
type Room struct {
	ID         string                `json:"id"`          // 房间的唯一内部标识（如 UUID，随机生成的一串字符）
	RoomNumber string                `json:"room_number"` // 供玩家输入的短房间号（如 "888888"），作为进入凭据
	Name       string                `json:"name"`
	IsPaused   bool                  `json:"is_paused"` // 房间是否处于全局暂停状态
	Users      map[string]*user.User `json:"users"`     // 当前在房间内的所有用户（包括玩家和旁观者）
	GameEngine core.GameEngine       `json:"-"`         // 挂载的具体游戏引擎实例
}

// 确保 Room 实现了 core.Messenger 接口
var _ core.Messenger = (*Room)(nil)

// Broadcast 广播给游戏内的所有玩家
func (r *Room) Broadcast(msgType string, reason string, payload any) {
	// TODO: 实现广播逻辑，例如将参数组装成 core.Message 并通过 wscore 发送给所有在房间内的用户
}

// SendTo 私发给游戏内的特定玩家
func (r *Room) SendTo(userID string, msgType string, reason string, payload any) {
	// TODO: 实现私发逻辑，通过 wscore 找到特定用户的连接并发送
}
