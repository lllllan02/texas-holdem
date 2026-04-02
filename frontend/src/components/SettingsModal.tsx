import { useState, useEffect } from 'react'
import { Settings, User } from 'lucide-react'
import { uploadImage } from '../api/user'

interface SettingsModalProps {
  show: boolean;
  onClose: () => void;
  userName: string;
  userAvatar?: string;
  setUserInfo: (name: string, avatar: string) => Promise<void> | void;
}

export function SettingsModal({ show, onClose, userName, userAvatar, setUserInfo }: SettingsModalProps) {
  const [tempUserName, setTempUserName] = useState('')
  const [tempAvatar, setTempAvatar] = useState('')
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [isSaving, setIsSaving] = useState(false)

  useEffect(() => {
    if (show) {
      setTempUserName(userName);
      setTempAvatar(userAvatar || '');
      setSelectedFile(null);
    }
  }, [show, userName, userAvatar]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && show && !isSaving) {
        onClose();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [show, isSaving, onClose]);

  const handleImageUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setSelectedFile(file);
      const reader = new FileReader();
      reader.onloadend = () => {
        setTempAvatar(reader.result as string);
      };
      reader.readAsDataURL(file);
    }
  };

  if (!show) return null;

  return (
    <div 
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
      onClick={(e) => {
        if (e.target === e.currentTarget && !isSaving) {
          onClose();
        }
      }}
    >
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
            <div className="relative group rounded-full">
              <input 
                type="file" 
                accept="image/*" 
                onChange={handleImageUpload}
                className="absolute inset-0 w-full h-full opacity-0 cursor-pointer z-10 rounded-full"
                title="点击更换头像"
              />
              <div className="w-16 h-16 bg-blue-600 rounded-full flex items-center justify-center shadow-inner overflow-hidden border-2 border-gray-700">
                {tempAvatar ? (
                  <img src={tempAvatar} alt="Avatar" className="w-full h-full object-cover" />
                ) : (
                  <User className="w-8 h-8 text-white" />
                )}
                <div className="absolute inset-0 bg-black/50 rounded-full opacity-0 group-hover:opacity-100 flex items-center justify-center transition-opacity pointer-events-none">
                  <span className="text-white text-xs">更换</span>
                </div>
              </div>
            </div>
            <span className="text-xs text-gray-500">点击更换头像</span>
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
            onClick={async () => {
              if (tempUserName.trim() && !isSaving) {
                setIsSaving(true);
                try {
                  let finalAvatar = tempAvatar;
                  if (selectedFile) {
                    finalAvatar = await uploadImage(selectedFile);
                  }
                  await setUserInfo(tempUserName.trim(), finalAvatar);
                  onClose();
                } catch (err) {
                  console.error('Failed to save settings:', err);
                  alert('保存失败，请稍后重试');
                } finally {
                  setIsSaving(false);
                }
              }
            }}
            disabled={isSaving}
            className={`px-4 py-2 rounded-lg text-sm font-medium text-white transition-colors shadow-md ${isSaving ? 'bg-blue-800 cursor-not-allowed' : 'bg-blue-600 hover:bg-blue-500'}`}
          >
            {isSaving ? '保存中...' : '保存修改'}
          </button>
        </div>
      </div>
    </div>
  )
}
