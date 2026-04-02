import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { User, Settings, Clock, Play, Pause, Crown } from 'lucide-react'
import { useUser } from '../hooks/useUser'
import { SettingsModal } from '../components/SettingsModal'
import { CreateRoomModal } from '../components/CreateRoomModal'
import { JoinRoomModal } from '../components/JoinRoomModal'
import { createRoom, getRoom, getUserActiveRooms } from '../api/room'
import type { UserActiveRoom } from '../api/room'

export default function Home() {
  const { user, loading, updateUserInfo } = useUser()
  const navigate = useNavigate()
  const [showSettings, setShowSettings] = useState(false)
  const [showJoinModal, setShowJoinModal] = useState(false)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [isCreating, setIsCreating] = useState(false)
  const [kickNotice, setKickNotice] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const [activeRooms, setActiveRooms] = useState<UserActiveRoom[]>([])

  useEffect(() => {
    try {
      const kick = sessionStorage.getItem('kick_notice')
      if (kick) {
        sessionStorage.removeItem('kick_notice')
        setKickNotice(kick)
      }

      const home = sessionStorage.getItem('home_notice')
      if (home) {
        sessionStorage.removeItem('home_notice')
        setNotice(home)
      }
    } catch {
      // ignore
    }
  }, [])

  useEffect(() => {
    const fetchActiveRooms = async () => {
      if (!user) return
      try {
        const { rooms } = await getUserActiveRooms(user.id)
        setActiveRooms(rooms.sort((a, b) => b.joined_at - a.joined_at))
      } catch (err: any) {
        console.error('Failed to fetch active rooms:', err)
      }
    }

    fetchActiveRooms()
  }, [user])

  const formatJoinTime = (timestamp: number) => {
    const now = Date.now()
    const diff = now - timestamp * 1000
    const minutes = Math.floor(diff / 60000)
    const hours = Math.floor(diff / 3600000)
    const days = Math.floor(diff / 86400000)

    if (minutes < 1) return '刚刚'
    if (minutes < 60) return `${minutes}分钟前`
    if (hours < 24) return `${hours}小时前`
    return `${days}天前`
  }

  const handleCreateRoom = async (roomOptions: any) => {
    if (!user) return;
    setIsCreating(true);
    try {
      const res = await createRoom(user.id, roomOptions);
      navigate(`/table?room=${res.room_number}`);
    } catch (err: any) {
      alert(err.message || '创建房间失败');
    } finally {
      setIsCreating(false);
      setShowCreateModal(false);
    }
  };

  const handleJoinRoom = async (roomNumber: string) => {
    const rn = roomNumber.trim()
    if (!rn) return;
    try {
      await getRoom(rn)
      setShowJoinModal(false)
      navigate(`/table?room=${rn}`)
    } catch (err: any) {
      setNotice(err?.message || '房间不存在或已被回收')
      setShowJoinModal(false)
    }
  };

  const handleQuickJoinRoom = async (roomNumber: string) => {
    try {
      await getRoom(roomNumber)
      navigate(`/table?room=${roomNumber}`)
    } catch (err: any) {
      setNotice(err?.message || '房间不存在或已被回收')
    }
  };

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gray-900 text-white relative">
      
      {/* 右上角用户信息 */}
      <header className="absolute top-0 right-0 p-4 flex justify-end z-10">
        {!loading && user && (
          <div className="flex items-center gap-2 bg-gray-800/50 pl-1 pr-2 py-1 rounded-full border border-gray-700">
            <div className="w-7 h-7 bg-blue-600 rounded-full flex items-center justify-center shadow-inner overflow-hidden">
              {user.avatar ? (
                <img src={user.avatar} alt="Avatar" className="w-full h-full object-cover" />
              ) : (
                <User className="w-4 h-4 text-white" />
              )}
            </div>
            <span className="text-sm text-white font-medium max-w-[80px] truncate">{user.nickname}</span>
            <button 
              onClick={() => setShowSettings(true)}
              className="text-gray-400 hover:text-white transition p-1 hover:bg-gray-700 rounded-full"
              title="修改用户信息"
            >
              <Settings className="w-4 h-4" />
            </button>
          </div>
        )}
      </header>

      <h1 className="text-5xl font-bold text-green-500 mb-8">
        Texas Hold'em
      </h1>
      
      {kickNotice && (
        <div className="mb-4 w-full max-w-md bg-red-900/20 border border-red-900/40 text-red-200 rounded-xl px-4 py-3 flex items-start justify-between gap-3 shadow-lg">
          <div className="text-sm leading-5">
            {kickNotice}
          </div>
          <button
            onClick={() => setKickNotice(null)}
            className="text-red-200/70 hover:text-red-100 transition"
            aria-label="关闭提示"
          >
            ✕
          </button>
        </div>
      )}

      {notice && (
        <div className="mb-4 w-full max-w-md bg-yellow-900/20 border border-yellow-900/40 text-yellow-100 rounded-xl px-4 py-3 flex items-start justify-between gap-3 shadow-lg">
          <div className="text-sm leading-5">
            {notice}
          </div>
          <button
            onClick={() => setNotice(null)}
            className="text-yellow-100/70 hover:text-yellow-100 transition"
            aria-label="关闭提示"
          >
            ✕
          </button>
        </div>
      )}

      <div className="bg-gray-800 p-8 rounded-xl shadow-2xl border border-gray-700 max-w-md w-full">
        {loading ? (
          <p className="text-gray-400 mb-6 text-center">加载用户信息中...</p>
        ) : (
          <p className="text-gray-400 mb-6 text-center">
            欢迎, <span className="text-blue-400 font-bold">{user?.nickname}</span>! 准备好加入牌桌了吗？
          </p>
        )}
        <div className="flex flex-col gap-4">
          <button 
            onClick={() => setShowCreateModal(true)}
            disabled={loading || !user}
            className="bg-green-600 hover:bg-green-500 disabled:bg-green-800 text-white font-semibold py-3 px-6 rounded-lg transition-colors duration-200"
          >
            创建房间
          </button>
          <button 
            onClick={() => setShowJoinModal(true)}
            className="bg-blue-600 hover:bg-blue-500 text-white font-semibold py-3 px-6 rounded-lg transition-colors duration-200"
          >
            加入房间
          </button>
        </div>

        {activeRooms && activeRooms.length > 0 && (
          <div className="mt-8">
            <div className="flex items-center gap-2 mb-4 text-gray-300">
              <Clock className="w-4 h-4" />
              <span className="text-sm font-medium">最近加入的房间</span>
              <span className="text-xs text-gray-500">({activeRooms.length})</span>
            </div>
            
            <div className="space-y-3">
              {activeRooms.map((room) => (
                <button
                  key={room.room_number}
                  onClick={() => handleQuickJoinRoom(room.room_number)}
                  className={`w-full ${room.is_owner ? 'bg-gradient-to-r from-green-700/30 to-green-800/30 hover:from-green-700/50 hover:to-green-800/50 border border-green-600/40 hover:border-green-500/60' : 'bg-gradient-to-r from-gray-700/60 to-gray-800/60 hover:from-green-700/40 hover:to-green-800/40 border border-gray-600 hover:border-green-500/50'} rounded-xl p-4 text-left transition-all duration-300 group relative overflow-hidden`}
                >
                  <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/5 to-transparent -translate-x-full group-hover:translate-x-full transition-transform duration-700"></div>
                  
                  <div className="flex items-center justify-between relative z-10">
                    <div>
                      <div className="flex items-center gap-2 mb-1">
                        {room.is_owner && <Crown className="w-3 h-3 text-green-400" />}
                        <div className="text-white font-semibold text-lg group-hover:text-green-300 transition-colors">
                          {room.room_number}
                        </div>
                      </div>
                      <div className="flex items-center gap-1.5">
                        {room.is_paused ? (
                          <>
                            <Pause className="w-3 h-3 text-yellow-400" />
                            <span className="text-gray-400 text-xs">已暂停</span>
                          </>
                        ) : (
                          <>
                            <Play className="w-3 h-3 text-green-400" />
                            <span className="text-gray-400 text-xs">进行中</span>
                          </>
                        )}
                        <span className="text-gray-500 text-xs ml-2">• {formatJoinTime(room.joined_at)}</span>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <div className="text-gray-400 group-hover:text-white transition-colors text-sm font-medium">
                        进入
                      </div>
                      <div className={`w-8 h-8 ${room.is_owner ? 'bg-gray-600/50 group-hover:bg-green-600' : 'bg-gray-600/50 group-hover:bg-green-600'} rounded-lg flex items-center justify-center transition-all duration-300`}>
                        <span className="text-gray-300 group-hover:text-white text-lg">→</span>
                      </div>
                    </div>
                  </div>
                </button>
              ))}
            </div>
          </div>
        )}
      </div>

      <SettingsModal 
        show={showSettings} 
        onClose={() => setShowSettings(false)} 
        userName={user?.nickname || ''} 
        userAvatar={user?.avatar || ''}
        setUserInfo={async (name, avatar) => {
          try {
            await updateUserInfo(name, avatar);
          } catch (e) {
            console.error('Failed to update user info', e);
            alert('修改失败');
          }
        }} 
      />

      <JoinRoomModal 
        show={showJoinModal} 
        onClose={() => setShowJoinModal(false)} 
        onJoin={handleJoinRoom} 
      />

      <CreateRoomModal 
        show={showCreateModal} 
        onClose={() => setShowCreateModal(false)} 
        isCreating={isCreating} 
        onCreate={handleCreateRoom} 
      />
    </div>
  )
}
