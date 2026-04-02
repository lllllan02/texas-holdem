import { Link } from 'react-router-dom'
import { ArrowLeft, User, History, Settings } from 'lucide-react'

interface HeaderProps {
  userName: string;
  onOpenHistory: () => void;
  onOpenSettings: () => void;
}

export function Header({ userName, onOpenHistory, onOpenSettings }: HeaderProps) {
  return (
    <header className="absolute top-0 left-0 right-0 p-4 flex justify-between items-center z-10">
      <Link to="/" className="flex items-center text-gray-400 hover:text-white transition">
        <ArrowLeft className="w-5 h-5 mr-2" />
        Back to Lobby
      </Link>
      <div className="flex items-center gap-3 sm:gap-4">
        <div className="text-gray-400 text-sm hidden sm:block">
          Room: <span className="text-white font-mono">123456</span>
        </div>
        <button 
          onClick={onOpenHistory}
          className="flex items-center gap-1.5 text-gray-400 hover:text-white transition bg-gray-800/50 px-3 py-1.5 rounded-full border border-gray-700 hover:border-gray-500"
        >
          <History className="w-4 h-4" />
          <span className="text-sm">History</span>
        </button>
        
        <div className="flex items-center gap-2 bg-gray-800/50 pl-1 pr-2 py-1 rounded-full border border-gray-700">
          <div className="w-7 h-7 bg-blue-600 rounded-full flex items-center justify-center shadow-inner">
            <User className="w-4 h-4 text-white" />
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
      </div>
    </header>
  )
}
