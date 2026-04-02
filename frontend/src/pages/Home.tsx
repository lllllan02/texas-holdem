import { useState } from 'react'
import { Link } from 'react-router-dom'
import { User, Settings } from 'lucide-react'
import { useUser } from '../hooks/useUser'
import { SettingsModal } from '../components/SettingsModal'

export default function Home() {
  const { user, loading, updateUserInfo } = useUser()
  const [showSettings, setShowSettings] = useState(false)

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
          <button className="bg-green-600 hover:bg-green-500 text-white font-semibold py-3 px-6 rounded-lg transition-colors duration-200">
            Create Room
          </button>
          <button className="bg-blue-600 hover:bg-blue-500 text-white font-semibold py-3 px-6 rounded-lg transition-colors duration-200">
            Join Room
          </button>
          
          <div className="mt-6 pt-6 border-t border-gray-700 text-center">
            <p className="text-sm text-gray-500 mb-4">UI Design Previews</p>
            <Link 
              to="/table" 
              className="text-green-400 hover:text-green-300 underline text-sm"
            >
              Preview Poker Table UI
            </Link>
          </div>
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
    </div>
  )
}
