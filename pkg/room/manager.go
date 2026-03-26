package room

import (
	"errors"
	"log"
	"sync"
)

// EngineFactory 引擎工厂函数
type EngineFactory func() GameEngine

// RoomManager 房间管理器
type RoomManager struct {
	mu        sync.RWMutex
	rooms     map[string]*Room
	factories map[string]EngineFactory // 注册的引擎工厂
}

// NewRoomManager 创建一个新的房间管理器
func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms:     make(map[string]*Room),
		factories: make(map[string]EngineFactory),
	}
}

// RegisterEngine 注册一个游戏引擎工厂
func (m *RoomManager) RegisterEngine(gameType string, factory EngineFactory) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.factories[gameType] = factory
}

// CreateRoom 创建一个新房间
func (m *RoomManager) CreateRoom(id string, hostID string, gameType string) (*Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.rooms[id]; exists {
		return nil, errors.New("room already exists")
	}

	factory, ok := m.factories[gameType]
	if !ok {
		return nil, errors.New("unsupported game type")
	}

	engine := factory()
	rm := NewRoom(id, hostID, engine, m)
	m.rooms[id] = rm
	return rm, nil
}

// GetRoom 获取指定 ID 的房间
func (m *RoomManager) GetRoom(id string) (*Room, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	rm, exists := m.rooms[id]
	return rm, exists
}

// DismissRoom 房主解散房间
func (m *RoomManager) DismissRoom(id string, userID string) error {
	m.mu.RLock()
	rm, exists := m.rooms[id]
	m.mu.RUnlock()

	if !exists {
		return errors.New("room not found")
	}

	if rm.GetHostID() != userID {
		return errors.New("only host can dismiss room")
	}

	// 广播解散消息
	rm.Broadcast("", ActionRoomState, "房间已被房主解散")

	// 移除房间
	m.RemoveRoom(id)
	return nil
}
// RemoveRoom 销毁并移除指定 ID 的房间
func (m *RoomManager) RemoveRoom(id string) {
	m.mu.Lock()
	room, exists := m.rooms[id]
	if exists {
		delete(m.rooms, id)
	}
	m.mu.Unlock()

	// 锁外执行销毁，避免阻塞其他房间的创建/查询操作
	if exists {
		room.destroy()
		log.Printf("Room [%s] destroyed", id)
	}
}
