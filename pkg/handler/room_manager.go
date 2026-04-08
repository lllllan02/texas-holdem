package handler

import (
	"log"
	"sync"
	"time"
)

// RoomManager 管理所有活跃的房间，并负责自动回收空闲房间
type RoomManager struct {
	mu              sync.RWMutex
	rooms           map[string]*Room
	idleTimers      map[string]*time.Timer
	recycleAfter    time.Duration
	userRoomManager *UserRoomManager
}

func NewRoomManager(recycleAfter time.Duration) *RoomManager {
	return &RoomManager{
		rooms:           make(map[string]*Room),
		idleTimers:      make(map[string]*time.Timer),
		recycleAfter:    recycleAfter,
		userRoomManager: NewUserRoomManager(),
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

	m.userRoomManager.RemoveRoom(roomNumber)

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
		// 计时器已触发，先从 idleTimers 摘除登记。
		// 否则若后续 Execute 内因又有连接而提前 return，映射里仍残留已触发的 timer，
		// OnRoomEmpty 会误判「已有计时器」而不再调度新一轮回收。
		m.mu.Lock()
		delete(m.idleTimers, roomNumber)
		m.mu.Unlock()

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

			m.userRoomManager.RemoveRoom(roomNumber)

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

func (m *RoomManager) RecordUserRoom(userID, roomNumber string, isOwner bool) {
	m.userRoomManager.RecordUserRoom(userID, roomNumber, isOwner)
}

func (m *RoomManager) GetUserActiveRooms(userID string) []map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userRooms := m.userRoomManager.GetUserRooms(userID)
	if userRooms == nil {
		return make([]map[string]interface{}, 0)
	}

	result := make([]map[string]interface{}, 0, len(userRooms))
	for roomNumber, userRoom := range userRooms {
		if room, ok := m.rooms[roomNumber]; ok {
			result = append(result, map[string]interface{}{
				"room_id":     room.ID,
				"room_number": room.RoomNumber,
				"owner_id":    room.OwnerID,
				"is_paused":   room.IsPaused,
				"is_owner":    userRoom.IsOwner,
				"joined_at":   userRoom.JoinedAt.Unix(),
			})
		}
	}

	return result
}
