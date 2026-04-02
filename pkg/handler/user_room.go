package handler

import (
	"sync"
	"time"
)

type UserRoom struct {
	UserID     string
	RoomNumber string
	JoinedAt   time.Time
	IsOwner    bool
}

type UserRoomManager struct {
	mu       sync.RWMutex
	userRoom map[string]map[string]*UserRoom // userID -> roomNumber -> UserRoom
}

func NewUserRoomManager() *UserRoomManager {
	return &UserRoomManager{
		userRoom: make(map[string]map[string]*UserRoom),
	}
}

func (m *UserRoomManager) RecordUserRoom(userID, roomNumber string, isOwner bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.userRoom[userID]; !ok {
		m.userRoom[userID] = make(map[string]*UserRoom)
	}

	m.userRoom[userID][roomNumber] = &UserRoom{
		UserID:     userID,
		RoomNumber: roomNumber,
		JoinedAt:   time.Now(),
		IsOwner:    isOwner,
	}
}

func (m *UserRoomManager) GetUserRooms(userID string) map[string]*UserRoom {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rooms, ok := m.userRoom[userID]
	if !ok {
		return nil
	}

	result := make(map[string]*UserRoom, len(rooms))
	for roomNumber, userRoom := range rooms {
		result[roomNumber] = userRoom
	}

	return result
}

func (m *UserRoomManager) RemoveRoom(roomNumber string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, rooms := range m.userRoom {
		delete(rooms, roomNumber)
	}
}
