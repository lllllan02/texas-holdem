import { useState, useEffect } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'

import { Header } from './table/components/Header'
import { SettingsModal } from '../components/SettingsModal'
import { useUser } from '../hooks/useUser'
import { useWebSocket } from '../hooks/useWebSocket'
import { deleteRoom } from '../api/room'

export default function Table() {
  const { user, loading, updateUserInfo } = useUser()
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const [showSettings, setShowSettings] = useState(false)
  const [isOwner, setIsOwner] = useState(false)
  
  const roomNumber = searchParams.get('room')
  
  // 初始化 WebSocket
  const { lastMessage } = useWebSocket(roomNumber, user?.id)

  useEffect(() => {
    if (!roomNumber) {
      navigate('/')
    }
  }, [roomNumber, navigate])

  // 监听 WebSocket 消息
  useEffect(() => {
    if (!lastMessage) return;

    if (lastMessage.type === 'room.welcome') {
      // 判断当前用户是否是房主
      setIsOwner(lastMessage.payload.owner_id === user?.id);
    } else if (lastMessage.type === 'room.destroyed') {
      alert('房间已被房主解散');
      navigate('/');
    }
  }, [lastMessage, user?.id, navigate]);

  const handleDeleteRoom = async () => {
    if (!roomNumber || !user?.id) return;
    if (!window.confirm('确定要解散房间吗？此操作不可恢复。')) return;

    try {
      await deleteRoom(roomNumber, user.id);
      // 解散成功后，后端会广播 room.destroyed 消息，前端收到后会自动跳转回大厅
    } catch (err: any) {
      alert(err.message || '解散房间失败');
    }
  };

  if (loading) {
    return <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center">Loading user info...</div>
  }

  return (
    <div className="min-h-screen bg-gray-900 text-white flex flex-col lg:flex-row overflow-hidden">
      
      {/* 左侧/上方：游戏主区域 */}
      <div className="flex-1 relative flex flex-col min-h-[60vh] lg:min-h-screen">
        
        <Header 
          userName={user?.nickname || 'Loading...'} 
          userAvatar={user?.avatar}
          roomNumber={roomNumber || 'Unknown'}
          isOwner={isOwner}
          onOpenHistory={() => {}} 
          onOpenSettings={() => setShowSettings(true)} 
          onDeleteRoom={handleDeleteRoom}
        />

        {/* 牌桌区域 */}
        <main className="flex-1 flex items-center justify-center p-4 sm:p-8 mt-16 lg:mt-0 relative">
          {/* 牌桌 (绿色椭圆) */}
          <div className="relative w-full max-w-2xl lg:max-w-4xl aspect-[2/1] bg-green-800 rounded-[1000px] border-[12px] sm:border-[16px] border-gray-800 shadow-2xl flex items-center justify-center">
            
            {/* 牌桌内圈线 */}
            <div className="absolute inset-3 sm:inset-4 rounded-[1000px] border-2 border-green-700 opacity-50 pointer-events-none"></div>

            <div className="text-gray-400 font-mono text-lg animate-pulse">
              Waiting for game data...
            </div>
          </div>
        </main>
      </div>

      {/* 右侧/下方：侧边栏 (聊天与操作) */}
      <aside className="w-full lg:w-80 xl:w-96 bg-gray-800 border-t lg:border-t-0 lg:border-l border-gray-700 flex flex-col z-30 h-auto lg:h-screen p-4 items-center justify-center text-gray-500">
        Action Bar & Chat Placeholder
      </aside>

      <SettingsModal 
        show={showSettings} 
        onClose={() => setShowSettings(false)} 
        userName={user?.nickname || ''} 
        userAvatar={user?.avatar}
        setUserInfo={async (name, avatar) => {
          try {
            await updateUserInfo(name, avatar);
          } catch (e) {
            console.error('Failed to update user info', e);
            alert('修改失败');
          }
        }} 
      />

    </div>
  )
}
