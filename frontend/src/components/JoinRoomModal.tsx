import { useState } from 'react'

interface JoinRoomModalProps {
  show: boolean;
  onClose: () => void;
  onJoin: (roomNumber: string) => void;
}

export function JoinRoomModal({ show, onClose, onJoin }: JoinRoomModalProps) {
  const [joinRoomNumber, setJoinRoomNumber] = useState('')

  if (!show) return null;

  return (
    <div 
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div className="bg-gray-800 border border-gray-700 rounded-xl shadow-2xl w-full max-w-sm flex flex-col p-6 animate-in fade-in zoom-in duration-200">
        <h2 className="text-xl font-bold text-white mb-4">加入房间</h2>
        <input 
          type="text" 
          value={joinRoomNumber}
          onChange={(e) => setJoinRoomNumber(e.target.value)}
          placeholder="请输入 6 位房间号"
          className="bg-gray-900 border border-gray-600 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-blue-500 transition-colors mb-6 font-mono text-center tracking-widest text-lg"
          maxLength={6}
          autoFocus
          onKeyDown={(e) => {
            if (e.key === 'Enter') onJoin(joinRoomNumber);
            if (e.key === 'Escape') onClose();
          }}
        />
        <div className="flex justify-end gap-3">
          <button 
            onClick={onClose}
            className="px-4 py-2 rounded-lg text-sm font-medium text-gray-300 hover:bg-gray-700 transition-colors"
          >
            取消
          </button>
          <button 
            onClick={() => onJoin(joinRoomNumber)}
            disabled={!joinRoomNumber.trim()}
            className="px-6 py-2 rounded-lg text-sm font-medium bg-blue-600 hover:bg-blue-500 disabled:bg-gray-600 text-white transition-colors shadow-md"
          >
            加入
          </button>
        </div>
      </div>
    </div>
  )
}
