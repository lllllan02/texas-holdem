import { useState } from 'react'

interface CreateRoomModalProps {
  show: boolean;
  onClose: () => void;
  isCreating: boolean;
  onCreate: (options: any) => void;
}

export function CreateRoomModal({ show, onClose, isCreating, onCreate }: CreateRoomModalProps) {
  const [roomOptions, setRoomOptions] = useState({
    player_count: 9,
    small_blind: 10,
    big_blind: 20,
    initial_chips: 2000,
    action_timeout: 30
  })
  
  const [bbMultiplier, setBbMultiplier] = useState(100)

  const blindOptions = [
    { sb: 5, bb: 10 },
    { sb: 10, bb: 20 },
    { sb: 25, bb: 50 },
    { sb: 50, bb: 100 },
    { sb: 100, bb: 200 }
  ]

  const bbMultiplierOptions = [50, 100, 200, 500]

  if (!show) return null;

  return (
    <div 
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
      onClick={(e) => {
        if (e.target === e.currentTarget && !isCreating) onClose();
      }}
    >
      <div className="bg-gray-800 border border-gray-700 rounded-xl shadow-2xl w-full max-w-md flex flex-col p-6 animate-in fade-in zoom-in duration-200">
        <h2 className="text-xl font-bold text-white mb-6">创建房间</h2>
        
        <div className="flex flex-col gap-6 mb-8">
          <div className="flex flex-col gap-2">
            <label className="text-sm text-gray-400">玩家人数</label>
            <div className="flex w-full bg-gray-900/80 p-1 rounded-lg border border-gray-700">
              {[2,3,4,5,6,7,8,9].map(n => (
                <button
                  key={n}
                  onClick={() => setRoomOptions({...roomOptions, player_count: n})}
                  className={`flex-1 py-1.5 rounded-md text-sm font-medium transition-all ${
                    roomOptions.player_count === n 
                      ? 'bg-blue-600 text-white shadow-sm' 
                      : 'text-gray-400 hover:text-gray-200 hover:bg-gray-800'
                  }`}
                >
                  {n}
                </button>
              ))}
            </div>
          </div>
          
          <div className="flex flex-col gap-2">
            <label className="text-sm text-gray-400">盲注 (小盲/大盲)</label>
            <div className="flex w-full bg-gray-900/80 p-1 rounded-lg border border-gray-700">
              {blindOptions.map(opt => {
                const isSelected = roomOptions.small_blind === opt.sb && roomOptions.big_blind === opt.bb;
                return (
                  <button
                    key={`${opt.sb}/${opt.bb}`}
                    onClick={() => setRoomOptions({
                      ...roomOptions, 
                      small_blind: opt.sb, 
                      big_blind: opt.bb,
                      initial_chips: opt.bb * bbMultiplier
                    })}
                    className={`flex-1 py-1.5 rounded-md text-sm font-medium transition-all ${
                      isSelected 
                        ? 'bg-blue-600 text-white shadow-sm' 
                        : 'text-gray-400 hover:text-gray-200 hover:bg-gray-800'
                    }`}
                  >
                    {opt.sb}/{opt.bb}
                  </button>
                );
              })}
            </div>
          </div>

          <div className="flex flex-col gap-2">
            <div className="flex justify-between items-center">
              <label className="text-sm text-gray-400">初始筹码</label>
              <span className="text-xs font-mono text-green-400 bg-green-900/20 px-2 py-0.5 rounded border border-green-900/50">
                总计: {roomOptions.initial_chips}
              </span>
            </div>
            <div className="flex w-full bg-gray-900/80 p-1 rounded-lg border border-gray-700">
              {bbMultiplierOptions.map(m => (
                <button
                  key={m}
                  onClick={() => {
                    setBbMultiplier(m);
                    setRoomOptions({
                      ...roomOptions,
                      initial_chips: roomOptions.big_blind * m
                    });
                  }}
                  className={`flex-1 py-1.5 rounded-md text-sm font-medium transition-all ${
                    bbMultiplier === m 
                      ? 'bg-blue-600 text-white shadow-sm' 
                      : 'text-gray-400 hover:text-gray-200 hover:bg-gray-800'
                  }`}
                >
                  {m} BB
                </button>
              ))}
            </div>
          </div>

          <div className="flex flex-col gap-2">
            <label className="text-sm text-gray-400">操作超时</label>
            <div className="flex w-full bg-gray-900/80 p-1 rounded-lg border border-gray-700">
              {[15, 30, 60, 120].map(t => (
                <button
                  key={t}
                  onClick={() => setRoomOptions({...roomOptions, action_timeout: t})}
                  className={`flex-1 py-1.5 rounded-md text-sm font-medium transition-all ${
                    roomOptions.action_timeout === t 
                      ? 'bg-blue-600 text-white shadow-sm' 
                      : 'text-gray-400 hover:text-gray-200 hover:bg-gray-800'
                  }`}
                >
                  {t}s
                </button>
              ))}
            </div>
          </div>
        </div>

        <div className="flex justify-end gap-3">
          <button 
            onClick={onClose}
            disabled={isCreating}
            className="px-4 py-2 rounded-lg text-sm font-medium text-gray-300 hover:bg-gray-700 transition-colors"
          >
            取消
          </button>
          <button 
            onClick={() => onCreate(roomOptions)}
            disabled={isCreating}
            className="px-6 py-2 rounded-lg text-sm font-medium bg-green-600 hover:bg-green-500 disabled:bg-green-800 text-white transition-colors shadow-md"
          >
            {isCreating ? '创建中...' : '创建'}
          </button>
        </div>
      </div>
    </div>
  )
}
