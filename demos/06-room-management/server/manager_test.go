package server

import (
	"testing"
)

// 1. 并发创建: 启动 100 个 Goroutine 同时调用 CreateRoom，断言不会发生 map 并发写入 panic，且成功创建 100 个房间。
func TestRoomManager_CreateRoom_Concurrent(t *testing.T) {
	// TODO: 初始化 RoomManager
	// TODO: 使用 sync.WaitGroup 启动 100 个 Goroutine
	// TODO: 每个 Goroutine 调用 CreateRoom
	// TODO: 断言管理器中的房间数量为 100
}

// 2. 隔离性测试: 向 Room A 广播一条消息，验证只有加入 Room A 的 Client 能收到，Room B 的 Client 收不到。
func TestRoomManager_Isolation(t *testing.T) {
	// TODO: 创建 RoomManager，并创建 Room A 和 Room B
	// TODO: 创建 Client 1 加入 Room A，Client 2 加入 Room B
	// TODO: 向 Room A 广播消息
	// TODO: 断言 Client 1 收到了消息，Client 2 没有收到
}

// 3. 满员拦截: 设定房间最大人数为 8。模拟 9 个玩家加入同一个房间，断言第 9 个玩家收到 "Room is full" 错误。
func TestRoomManager_JoinRoom_Full(t *testing.T) {
	// TODO: 创建 RoomManager 和一个房间
	// TODO: 循环 8 次，创建 8 个 Client 并成功加入房间
	// TODO: 创建第 9 个 Client 尝试加入房间
	// TODO: 断言返回的错误信息为 "Room is full"
}
