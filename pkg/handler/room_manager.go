package handler

import (
	"log"
	"sync"
	"time"
)

// RoomManager 管理所有活跃的房间，并负责自动回收空闲房间
type RoomManager struct {
	mu           sync.RWMutex
	rooms        map[string]*Room
	idleTimers   map[string]*time.Timer
	recycleAfter time.Duration
}

func NewRoomManager(recycleAfter time.Duration) *RoomManager {
	return &RoomManager{
		rooms:        make(map[string]*Room),
		idleTimers:   make(map[string]*time.Timer),
		recycleAfter: recycleAfter,
	}
}

func (m *RoomManager) Add(room *Room) {
	m.mu.Lock()
	m.rooms[room.RoomNumber] = room
	m.mu.Unlock()
}

func (m *RoomManager) Get(roomNumber string) (*Room, bool) {
	m.mu.RLock()
	room, ok := m.rooms[roomNumber]
	m.mu.RUnlock()
	return room, ok
}

// Delete 显式删除房间（例如房主解散）。不会阻塞持锁调用 Stop，避免死锁。
func (m *RoomManager) Delete(roomNumber string) (*Room, bool) {
	m.mu.Lock()
	room, ok := m.rooms[roomNumber]
	if !ok {
		m.mu.Unlock()
		return nil, false
	}
	delete(m.rooms, roomNumber)
	m.cancelIdleTimerLocked(roomNumber)
	m.mu.Unlock()
	return room, true
}

// OnRoomActive 表示该房间出现活跃连接（用于取消回收计时器）
func (m *RoomManager) OnRoomActive(roomNumber string) {
	m.mu.Lock()
	m.cancelIdleTimerLocked(roomNumber)
	m.mu.Unlock()
}

// OnRoomEmpty 表示该房间当前无任何连接（用于启动回收计时器）
func (m *RoomManager) OnRoomEmpty(roomNumber string) {
	if m.recycleAfter <= 0 {
		return
	}

	m.mu.Lock()
	// 房间可能已被删除
	room, ok := m.rooms[roomNumber]
	if !ok {
		m.mu.Unlock()
		return
	}
	// 已经有计时器则不重复创建
	if _, exists := m.idleTimers[roomNumber]; exists {
		m.mu.Unlock()
		return
	}

	timer := time.AfterFunc(m.recycleAfter, func() {
		// Timer 回调在独立 goroutine：不持 manager 锁，避免和 WS/HTTP 互相等待
		room.Execute(func() {
			// 在房间 Hub 主循环内确认仍为空，避免误回收
			if len(room.clients) != 0 {
				return
			}

			// 从管理器移除（取消/清理计时器记录），再 Stop 房间
			m.mu.Lock()
			// 再次确认仍在管理器中，且还是同一个 room 实例
			if cur, ok := m.rooms[roomNumber]; !ok || cur != room {
				m.mu.Unlock()
				return
			}
			delete(m.rooms, roomNumber)
			m.cancelIdleTimerLocked(roomNumber)
			m.mu.Unlock()

			log.Printf("Room auto recycled: Number=%s (idle %s)", roomNumber, m.recycleAfter)
			room.Stop()
		})
	})

	m.idleTimers[roomNumber] = timer
	m.mu.Unlock()
}

func (m *RoomManager) cancelIdleTimerLocked(roomNumber string) {
	if t := m.idleTimers[roomNumber]; t != nil {
		t.Stop()
	}
	delete(m.idleTimers, roomNumber)
}

