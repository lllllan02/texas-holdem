import { MessageSquare } from 'lucide-react'

export function ChatLog() {
  return (
    <div className="flex-1 p-4 overflow-y-auto hidden lg:flex flex-col gap-3">
      <div className="flex items-center gap-2 text-gray-400 mb-2 sticky top-0 bg-gray-800 pb-2 z-10 border-b border-gray-700">
        <MessageSquare className="w-4 h-4" />
        <span className="text-sm font-semibold">游戏日志与聊天</span>
      </div>
      
      <div className="text-xs text-gray-500 italic px-2">
        <span className="text-green-500 font-semibold not-italic">系统:</span> 欢迎来到房间 123456
      </div>
      <div className="text-xs text-gray-500 italic px-2">
        <span className="text-blue-400 font-semibold not-italic">玩家 1</span> 坐下了 (座位 1)
      </div>
      
      <div className="flex flex-col gap-1 mt-1">
        <span className="text-xs text-yellow-400 font-semibold px-2">玩家 2</span>
        <div className="text-sm text-gray-200 bg-gray-700/80 p-2.5 rounded-2xl rounded-tl-sm self-start max-w-[90%] shadow-sm">
          祝大家好运！开始吧。
        </div>
      </div>

      <div className="text-xs text-gray-500 italic px-2 mt-1">
        <span className="text-yellow-400 font-semibold not-italic">玩家 2</span> 加注到 $100
      </div>

      <div className="flex flex-col gap-1 mt-1 items-end">
        <span className="text-xs text-blue-400 font-semibold px-2">你</span>
        <div className="text-sm text-white bg-blue-600/90 p-2.5 rounded-2xl rounded-tr-sm self-end max-w-[90%] shadow-sm">
          放马过来！😎
        </div>
      </div>
    </div>
  )
}
