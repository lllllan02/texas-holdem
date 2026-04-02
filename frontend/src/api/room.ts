export interface CreateRoomOptions {
  player_count?: number;
  small_blind?: number;
  big_blind?: number;
  initial_chips?: number;
  action_timeout?: number;
}

export interface CreateRoomResponse {
  room_id: string;
  room_number: string;
}

export const createRoom = async (ownerId: string, options?: CreateRoomOptions): Promise<CreateRoomResponse> => {
  const response = await fetch('/api/v1/rooms', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      owner_id: ownerId,
      options: options || {},
    }),
  });

  if (!response.ok) {
    const err = await response.json().catch(() => ({}));
    throw new Error(err.error || 'Failed to create room');
  }

  return response.json();
};

export const deleteRoom = async (roomNumber: string, ownerId: string): Promise<void> => {
  const response = await fetch(`/api/v1/rooms/${roomNumber}?owner_id=${ownerId}`, {
    method: 'DELETE',
  });

  if (!response.ok) {
    const err = await response.json().catch(() => ({}));
    throw new Error(err.error || 'Failed to delete room');
  }
};

export interface GetRoomResponse {
  room_id: string;
  room_number: string;
  owner_id: string;
}

export const getRoom = async (roomNumber: string): Promise<GetRoomResponse> => {
  const response = await fetch(`/api/v1/rooms/${roomNumber}`, { method: 'GET' });
  if (!response.ok) {
    const err = await response.json().catch(() => ({}));
    throw new Error(err.error || '房间不存在');
  }
  return response.json();
};

export interface UserActiveRoom {
  room_id: string;
  room_number: string;
  owner_id: string;
  is_paused: boolean;
  is_owner: boolean;
  joined_at: number;
}

export interface GetUserActiveRoomsResponse {
  rooms: UserActiveRoom[];
}

export const getUserActiveRooms = async (userId: string): Promise<GetUserActiveRoomsResponse> => {
  const response = await fetch(`/api/v1/users/${userId}/rooms`, { method: 'GET' });
  if (!response.ok) {
    const err = await response.json().catch(() => ({}));
    throw new Error(err.error || '获取房间列表失败');
  }
  return response.json();
};
