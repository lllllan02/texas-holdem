import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { User, Settings } from 'lucide-react'
import { useUser } from '../hooks/useUser'
import { SettingsModal } from '../components/SettingsModal'
import { createRoom } from '../api/room'

export default function Home() {
  const { user, loading, updateUserInfo } = useUser()
  const navigate = useNavigate()
  const [showSettings, setShowSettings] = useState(false)
  const [showJoinModal, setShowJoinModal] = useState(false)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [joinRoomNumber, setJoinRoomNumber] = useState('')
  const [isCreating, setIsCreating] = useState(false)
  const [roomOptions, setRoomOptions] = useState({
    player_count: 9,
    small_blind: 10,
    big_blind: 20,
    initial_chips: 2000,
    action_timeout: 30
  })

  const handleCreateRoom = async () => {
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

  const handleJoinRoom = () => {
    if (!joinRoomNumber.trim()) return;
    navigate(`/table?room=${joinRoomNumber.trim()}`);
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
      
      <div className="bg-gray-800 p-8 rounded-xl shadow-2xl border border-gray-700 max-w-md w-full">
        {loading ? (
          <p className="text-gray-400 mb-6 text-center">Loading user info...</p>
        ) : (
          <p className="text-gray-400 mb-6 text-center">
            Welcome, <span className="text-blue-400 font-bold">{user?.nickname}</span>! Ready to join the table?
          </p>
        )}
        <div className="flex flex-col gap-4">
          <button 
            onClick={() => setShowCreateModal(true)}
            disabled={loading || !user}
            className="bg-green-600 hover:bg-green-500 disabled:bg-green-800 text-white font-semibold py-3 px-6 rounded-lg transition-colors duration-200"
          >
            Create Room
          </button>
          <button 
            onClick={() => setShowJoinModal(true)}
            className="bg-blue-600 hover:bg-blue-500 text-white font-semibold py-3 px-6 rounded-lg transition-colors duration-200"
          >
            Join Room
          </button>
        </div>
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

      {/* Join Room Modal */}
      {showJoinModal && (
        <div 
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
          onClick={(e) => {
            if (e.target === e.currentTarget) setShowJoinModal(false);
          }}
        >
          <div className="bg-gray-800 border border-gray-700 rounded-xl shadow-2xl w-full max-w-sm flex flex-col p-6 animate-in fade-in zoom-in duration-200">
            <h2 className="text-xl font-bold text-white mb-4">Join Room</h2>
            <input 
              type="text" 
              value={joinRoomNumber}
              onChange={(e) => setJoinRoomNumber(e.target.value)}
              placeholder="Enter 6-digit room number"
              className="bg-gray-900 border border-gray-600 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-blue-500 transition-colors mb-6 font-mono text-center tracking-widest text-lg"
              maxLength={6}
              autoFocus
              onKeyDown={(e) => {
                if (e.key === 'Enter') handleJoinRoom();
                if (e.key === 'Escape') setShowJoinModal(false);
              }}
            />
            <div className="flex justify-end gap-3">
              <button 
                onClick={() => setShowJoinModal(false)}
                className="px-4 py-2 rounded-lg text-sm font-medium text-gray-300 hover:bg-gray-700 transition-colors"
              >
                Cancel
              </button>
              <button 
                onClick={handleJoinRoom}
                disabled={!joinRoomNumber.trim()}
                className="px-6 py-2 rounded-lg text-sm font-medium bg-blue-600 hover:bg-blue-500 disabled:bg-gray-600 text-white transition-colors shadow-md"
              >
                Join
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Create Room Modal */}
      {showCreateModal && (
        <div 
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
          onClick={(e) => {
            if (e.target === e.currentTarget && !isCreating) setShowCreateModal(false);
          }}
        >
          <div className="bg-gray-800 border border-gray-700 rounded-xl shadow-2xl w-full max-w-sm flex flex-col p-6 animate-in fade-in zoom-in duration-200">
            <h2 className="text-xl font-bold text-white mb-6">Create Room</h2>
            
            <div className="flex flex-col gap-4 mb-6">
              <div className="flex justify-between items-center">
                <label className="text-sm text-gray-400">Player Count</label>
                <select 
                  value={roomOptions.player_count}
                  onChange={(e) => setRoomOptions({...roomOptions, player_count: Number(e.target.value)})}
                  className="bg-gray-900 border border-gray-600 rounded px-2 py-1 text-white text-sm focus:outline-none focus:border-blue-500"
                >
                  {[2,3,4,5,6,7,8,9].map(n => <option key={n} value={n}>{n} Players</option>)}
                </select>
              </div>
              
              <div className="flex justify-between items-center">
                <label className="text-sm text-gray-400">Small Blind</label>
                <input 
                  type="number" 
                  value={roomOptions.small_blind}
                  onChange={(e) => setRoomOptions({...roomOptions, small_blind: Number(e.target.value)})}
                  className="bg-gray-900 border border-gray-600 rounded px-2 py-1 text-white text-sm w-24 text-right focus:outline-none focus:border-blue-500"
                />
              </div>

              <div className="flex justify-between items-center">
                <label className="text-sm text-gray-400">Big Blind</label>
                <input 
                  type="number" 
                  value={roomOptions.big_blind}
                  onChange={(e) => setRoomOptions({...roomOptions, big_blind: Number(e.target.value)})}
                  className="bg-gray-900 border border-gray-600 rounded px-2 py-1 text-white text-sm w-24 text-right focus:outline-none focus:border-blue-500"
                />
              </div>

              <div className="flex justify-between items-center">
                <label className="text-sm text-gray-400">Initial Chips</label>
                <input 
                  type="number" 
                  value={roomOptions.initial_chips}
                  onChange={(e) => setRoomOptions({...roomOptions, initial_chips: Number(e.target.value)})}
                  className="bg-gray-900 border border-gray-600 rounded px-2 py-1 text-white text-sm w-24 text-right focus:outline-none focus:border-blue-500"
                />
              </div>

              <div className="flex justify-between items-center">
                <label className="text-sm text-gray-400">Action Timeout (s)</label>
                <input 
                  type="number" 
                  value={roomOptions.action_timeout}
                  onChange={(e) => setRoomOptions({...roomOptions, action_timeout: Number(e.target.value)})}
                  className="bg-gray-900 border border-gray-600 rounded px-2 py-1 text-white text-sm w-24 text-right focus:outline-none focus:border-blue-500"
                />
              </div>
            </div>

            <div className="flex justify-end gap-3">
              <button 
                onClick={() => setShowCreateModal(false)}
                disabled={isCreating}
                className="px-4 py-2 rounded-lg text-sm font-medium text-gray-300 hover:bg-gray-700 transition-colors"
              >
                Cancel
              </button>
              <button 
                onClick={handleCreateRoom}
                disabled={isCreating}
                className="px-6 py-2 rounded-lg text-sm font-medium bg-green-600 hover:bg-green-500 disabled:bg-green-800 text-white transition-colors shadow-md"
              >
                {isCreating ? 'Creating...' : 'Create'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
