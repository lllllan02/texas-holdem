import { useState, useEffect } from 'react'
import { Settings, User } from 'lucide-react'

interface SettingsModalProps {
  show: boolean;
  onClose: () => void;
  userName: string;
  setUserName: (name: string) => void;
}

export function SettingsModal({ show, onClose, userName, setUserName }: SettingsModalProps) {
  const [tempUserName, setTempUserName] = useState('')

  useEffect(() => {
    if (show) setTempUserName(userName);
  }, [show, userName]);

  if (!show) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
      <div className="bg-gray-800 border border-gray-700 rounded-xl shadow-2xl w-full max-w-sm flex flex-col">
        <div className="p-4 border-b border-gray-700 flex justify-between items-center">
          <h2 className="text-lg font-bold text-white flex items-center gap-2">
            <Settings className="w-5 h-5 text-blue-400" />
            修改用户信息
          </h2>
          <button 
            onClick={onClose}
            className="text-gray-400 hover:text-white transition"
          >
            ✕
          </button>
        </div>
        
        <div className="p-6 flex flex-col gap-4">
          <div className="flex flex-col items-center gap-3 mb-2">
            <div className="w-16 h-16 bg-blue-600 rounded-full flex items-center justify-center shadow-inner relative group cursor-pointer">
              <User className="w-8 h-8 text-white" />
              <div className="absolute inset-0 bg-black/50 rounded-full opacity-0 group-hover:opacity-100 flex items-center justify-center transition-opacity">
                <span className="text-white text-xs">更换</span>
              </div>
            </div>
            <span className="text-xs text-gray-500">点击更换头像 (暂未实现)</span>
          </div>

          <div className="flex flex-col gap-2">
            <label className="text-sm text-gray-400 font-medium">昵称</label>
            <input 
              type="text" 
              value={tempUserName}
              onChange={(e) => setTempUserName(e.target.value)}
              placeholder="请输入您的昵称"
              maxLength={12}
              className="bg-gray-900 border border-gray-600 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-blue-500 transition-colors"
            />
          </div>
        </div>

        <div className="p-4 border-t border-gray-700 flex justify-end gap-3 bg-gray-900/30 rounded-b-xl">
          <button 
            onClick={onClose}
            className="px-4 py-2 rounded-lg text-sm font-medium text-gray-300 hover:bg-gray-700 transition-colors"
          >
            取消
          </button>
          <button 
            onClick={() => {
              if (tempUserName.trim()) {
                setUserName(tempUserName.trim());
                onClose();
              }
            }}
            className="px-4 py-2 rounded-lg text-sm font-medium bg-blue-600 hover:bg-blue-500 text-white transition-colors shadow-md"
          >
            保存修改
          </button>
        </div>
      </div>
    </div>
  )
}
