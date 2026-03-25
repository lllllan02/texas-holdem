# Demo 06: 房间管理与联机大厅 (Room Management)

## 目标
支持多桌并发游戏。玩家可以通过房间号加入特定的桌子，不同桌子的游戏状态互不干扰。

## 目录结构建议
```text
poker/
├── server/
│   ├── room.go        # 单个房间管理 (包含一个 Table 和一个 Hub)
│   ├── manager.go     # 全局房间管理器
│   └── manager_test.go
```

## 核心数据结构设计 (Go)
```go
// 房间
type Room struct {
    ID    string
    Table *game.Table // 游戏逻辑状态机
    Hub   *Hub        // 该房间专属的 WebSocket 广播中心
}

// 全局管理器
type RoomManager struct {
    Rooms map[string]*Room
    mu    sync.RWMutex // 保护并发读写
}
```

## 核心方法签名
- `func (m *RoomManager) CreateRoom() *Room`：生成一个随机的 6 位数房间号，初始化 Room 并存入 map。
- `func (m *RoomManager) JoinRoom(roomID string, client *Client) error`：将玩家的 WS 连接加入到指定房间的 Hub 中。
- `func (m *RoomManager) RemoveRoom(roomID string)`：当房间空了之后，清理资源。

## 具体测试用例 (manager_test.go)
1. **并发创建**: 启动 100 个 Goroutine 同时调用 `CreateRoom`，断言不会发生 map 并发写入 panic，且成功创建 100 个房间。
2. **隔离性测试**: 向 Room A 广播一条消息，验证只有加入 Room A 的 Client 能收到，Room B 的 Client 收不到。
3. **满员拦截**: 设定房间最大人数为 8。模拟 9 个玩家加入同一个房间，断言第 9 个玩家收到 "Room is full" 错误。