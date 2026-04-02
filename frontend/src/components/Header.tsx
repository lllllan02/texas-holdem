import { useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import { ArrowLeft, User, History, Settings, LogOut, Copy } from 'lucide-react'

interface HeaderProps {
  userName: string;
  userAvatar?: string;
  roomNumber: string;
  isOwner: boolean;
  onOpenHistory: () => void;
  onOpenSettings: () => void;
  onDeleteRoom: () => void;
}

export function Header({ userName, userAvatar, roomNumber, isOwner, onOpenHistory, onOpenSettings, onDeleteRoom }: HeaderProps) {
  const [copied, setCopied] = useState(false)
  const copiedTimerRef = useRef<number | null>(null)

  const copyRoomNumber = async () => {
    try {
      await navigator.clipboard.writeText(roomNumber)
    } catch {
      const input = document.createElement('input')
      input.value = roomNumber
      document.body.appendChild(input)
      input.select()
      document.execCommand('copy')
      document.body.removeChild(input)
    }

    setCopied(true)
    if (copiedTimerRef.current) {
      window.clearTimeout(copiedTimerRef.current)
    }
    copiedTimerRef.current = window.setTimeout(() => setCopied(false), 1200)
  }

  return (
    <header className="absolute top-0 left-0 right-0 p-4 flex justify-between items-center z-10">
      <Link to="/" className="flex items-center text-gray-400 hover:text-white transition">
        <ArrowLeft className="w-5 h-5 mr-2" />
        返回大厅
      </Link>
      <div className="flex items-center gap-3 sm:gap-4">
        <button
          type="button"
          onClick={copyRoomNumber}
          className="hidden sm:flex items-center gap-2 text-gray-400 text-sm bg-gray-800/50 px-3 py-1.5 rounded-full border border-gray-700 hover:border-gray-500 hover:text-white transition"
          title="点击复制房间号"
        >
          <span>
            房间号: <span className="text-white font-mono">{roomNumber}</span>
          </span>
          {copied ? (
            <span className="text-xs text-green-400 font-semibold">已复制</span>
          ) : (
            <Copy className="w-4 h-4 text-gray-500" />
          )}
        </button>
        <button 
          onClick={onOpenHistory}
          className="flex items-center gap-1.5 text-gray-400 hover:text-white transition bg-gray-800/50 px-3 py-1.5 rounded-full border border-gray-700 hover:border-gray-500"
        >
          <History className="w-4 h-4" />
          <span className="text-sm">历史对局</span>
        </button>
        
        <div className="flex items-center gap-2 bg-gray-800/50 pl-1 pr-2 py-1 rounded-full border border-gray-700">
          <div className="w-7 h-7 bg-blue-600 rounded-full flex items-center justify-center shadow-inner overflow-hidden">
            {userAvatar ? (
              <img src={userAvatar} alt="Avatar" className="w-full h-full object-cover" />
            ) : (
              <User className="w-4 h-4 text-white" />
            )}
          </div>
          <span className="text-sm text-white font-medium max-w-[80px] truncate">{userName}</span>
          <button 
            onClick={onOpenSettings}
            className="text-gray-400 hover:text-white transition p-1 hover:bg-gray-700 rounded-full"
            title="修改用户信息"
          >
            <Settings className="w-4 h-4" />
          </button>
        </div>

        {isOwner && (
          <button 
            onClick={onDeleteRoom}
            className="flex items-center gap-1.5 text-red-400 hover:text-red-300 transition bg-red-900/20 hover:bg-red-900/40 px-3 py-1.5 rounded-full border border-red-900/50"
            title="解散房间"
          >
            <LogOut className="w-4 h-4" />
            <span className="text-sm hidden sm:block">解散房间</span>
          </button>
        )}
      </div>
    </header>
  )
}
