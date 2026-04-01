import { useState } from 'react'
import { Link } from 'react-router-dom'
import { ArrowLeft, User, MessageSquare } from 'lucide-react'

// 这是一个纯 UI 设计页面，没有接入后端数据
export default function Table() {
  const [playerCount, setPlayerCount] = useState(9)
  const [betAmount, setBetAmount] = useState(60)

  // 模拟数据
  const minBet = 60;
  const maxBet = 2336;
  const bb = 20;

  // 使用椭圆方程计算座位分布
  const getSeatPosition = (total: number, index: number) => {
    const angle = Math.PI / 2 + (index * 2 * Math.PI) / total;
    const rx = 54; 
    const ry = 62; 
    const x = 50 + rx * Math.cos(angle);
    const y = 50 + ry * Math.sin(angle);
    return { 
      left: `${x}%`, 
      top: `${y}%`,
      transform: 'translate(-50%, -50%)'
    };
  };

  // 计算底牌在牌桌内圈的位置（相对于头像）
  // 核心思想：底牌应该朝着牌桌中心 (50%, 50%) 的方向偏移
  const getCardsPosition = (total: number, index: number) => {
    if (index === 0) return {}; // 自己 (正下方) 保持原样

    const angle = Math.PI / 2 + (index * 2 * Math.PI) / total;
    // 偏移距离 (px)，可以根据需要调整
    const offset = 95; 
    
    // 朝向中心 (50%, 50%) 的反方向是 -cos 和 -sin
    const dx = -offset * Math.cos(angle);
    const dy = -offset * Math.sin(angle);

    return {
      position: 'absolute' as const,
      top: '50%',
      left: '50%',
      transform: `translate(calc(-50% + ${dx}px), calc(-50% + ${dy}px))`,
      zIndex: 30,
    };
  };

  // 计算下注筹码在牌桌上的位置（比底牌更靠近牌桌中心）
  const getBetPosition = (total: number, index: number) => {
    const angle = Math.PI / 2 + (index * 2 * Math.PI) / total;
    // 筹码的偏移量比底牌(95)更大，意味着推得更深（更靠近中心）
    // 增加偏移量，避免和扑克牌重叠
    const offset = index === 0 ? 110 : 155; 
    
    const dx = -offset * Math.cos(angle);
    const dy = -offset * Math.sin(angle);

    return {
      position: 'absolute' as const,
      top: '50%',
      left: '50%',
      transform: `translate(calc(-50% + ${dx}px), calc(-50% + ${dy}px))`,
      zIndex: 40, // 提高层级，确保在扑克牌之上
    };
  };

  // 模拟获取玩家位置 (BTN, SB, BB, UTG...)
  const getMockPosition = (total: number, index: number) => {
    // 假设 index 1 始终是庄家 (BTN)
    const dealerIndex = 1;
    const offset = (index - dealerIndex + total) % total;
    
    if (total === 2) return offset === 0 ? 'BTN' : 'BB';
    if (total === 3) return ['BTN', 'SB', 'BB'][offset];
    if (total === 4) return ['BTN', 'SB', 'BB', 'UTG'][offset];
    if (total === 5) return ['BTN', 'SB', 'BB', 'UTG', 'CO'][offset];
    if (total === 6) return ['BTN', 'SB', 'BB', 'UTG', 'HJ', 'CO'][offset];
    if (total === 7) return ['BTN', 'SB', 'BB', 'UTG', 'MP', 'HJ', 'CO'][offset];
    if (total === 8) return ['BTN', 'SB', 'BB', 'UTG', 'UTG+1', 'MP', 'HJ', 'CO'][offset];
    if (total === 9) return ['BTN', 'SB', 'BB', 'UTG', 'UTG+1', 'UTG+2', 'MP', 'HJ', 'CO'][offset];
    return '';
  };

  return (
    <div className="min-h-screen bg-gray-900 text-white flex flex-col lg:flex-row overflow-hidden">
      
      {/* 左侧/上方：游戏主区域 */}
      <div className="flex-1 relative flex flex-col min-h-[60vh] lg:min-h-screen">
        {/* 顶部导航栏 */}
        <header className="absolute top-0 left-0 right-0 p-4 flex justify-between items-center z-10">
          <Link to="/" className="flex items-center text-gray-400 hover:text-white transition">
            <ArrowLeft className="w-5 h-5 mr-2" />
            Back to Lobby
          </Link>
          <div className="text-gray-400 text-sm">
            Room: <span className="text-white font-mono">123456</span>
          </div>
        </header>

        {/* 控制面板 (仅用于 UI 预览) */}
        <div className="absolute top-16 right-4 sm:right-8 z-20 bg-gray-800 p-4 rounded-xl border border-gray-700 shadow-xl w-48">
          <label className="block text-sm text-green-400 font-semibold mb-3">
            Table Size: {playerCount} Players
          </label>
          <input
            type="range"
            min="2"
            max="9"
            value={playerCount}
            onChange={(e) => setPlayerCount(Number(e.target.value))}
            className="w-full accent-green-500"
          />
          <div className="flex justify-between text-xs text-gray-500 mt-1">
            <span>2</span>
            <span>9</span>
          </div>
        </div>

        {/* 牌桌区域 */}
        <main className="flex-1 flex items-center justify-center p-4 sm:p-8 mt-16 lg:mt-0 relative">
          {/* 牌桌 (绿色椭圆) */}
          <div className="relative w-full max-w-2xl lg:max-w-4xl aspect-[2/1] bg-green-800 rounded-[1000px] border-[12px] sm:border-[16px] border-gray-800 shadow-2xl flex items-center justify-center">
            
            {/* 牌桌内圈线 */}
            <div className="absolute inset-3 sm:inset-4 rounded-[1000px] border-2 border-green-700 opacity-50 pointer-events-none"></div>

            {/* 公共牌区域 (Board) */}
            <div className="flex gap-1 sm:gap-2 z-10">
              {[1, 2, 3, 4, 5].map((i) => (
                <div key={i} className="w-10 h-14 sm:w-16 sm:h-24 bg-white rounded-md shadow-md border border-gray-300 flex items-center justify-center text-red-500 font-bold text-lg sm:text-xl">
                  A♥
                </div>
              ))}
            </div>

            {/* 底池信息 (Pot) */}
            <div className="absolute top-1/4 left-1/2 -translate-x-1/2 bg-black/60 px-4 py-1 rounded-full text-green-400 font-bold text-xs sm:text-sm border border-green-900/50">
              Pot: $1,500
            </div>

            {/* 动态渲染玩家座位 */}
            {Array.from({ length: playerCount }).map((_, i) => {
              const pos = getSeatPosition(playerCount, i);
              const isMe = i === 0;

              // 模拟不同玩家的状态和动作
              let playerState = 'active';
              let lastAction = ''; // 'bet', 'raise', 'call', 'check', 'allin', 'fold'

              if (i === 0) {
                lastAction = 'raise'; // 自己模拟 raise
              } else if (i === 1) {
                lastAction = 'call';
              } else if (i === 2 && playerCount > 2) {
                playerState = 'folded';
                lastAction = 'fold';
              } else if (i === 3 && playerCount > 3) {
                playerState = 'allin';
                lastAction = 'allin';
              } else if (i === 4 && playerCount > 4) {
                lastAction = 'bet';
              } else {
                lastAction = i % 2 === 0 ? 'check' : 'call';
              }

              const isFolded = playerState === 'folded';
              const isAllIn = playerState === 'allin';

              // 模拟底牌显示状态 (自己或对手1可见，弃牌不显示)
              const showCards = (isMe || i === 1) && !isFolded; 

              // 模拟玩家的当前下注额
              let mockBet = 0;
              if (!isFolded) {
                if (isMe) mockBet = betAmount;
                else if (isAllIn) mockBet = 1000;
                else if (lastAction === 'bet') mockBet = 200;
                else if (lastAction === 'raise') mockBet = 400;
                else if (lastAction === 'call') mockBet = 50;
              }

              // 根据动作决定筹码颜色
              let chipColorClass = 'bg-blue-500 border-blue-300'; // 默认 call 为蓝色
              let textColorClass = 'text-blue-400';
              let bgColorClass = 'bg-blue-900/60 border-blue-500/50';

              if (isAllIn) {
                chipColorClass = 'bg-red-500 border-red-300';
                textColorClass = 'text-red-400';
                bgColorClass = 'bg-red-900/80 border-red-500/50';
              } else if (lastAction === 'raise') {
                chipColorClass = 'bg-orange-500 border-orange-300';
                textColorClass = 'text-orange-400';
                bgColorClass = 'bg-orange-900/60 border-orange-500/50';
              } else if (lastAction === 'bet') {
                chipColorClass = 'bg-yellow-500 border-yellow-300'; // bet 为黄色
                textColorClass = 'text-yellow-400';
                bgColorClass = 'bg-black/60 border-yellow-500/50';
              }

              return (
                <div
                  key={i}
                  className={`absolute flex flex-col items-center justify-center z-20 transition-all duration-500 ease-in-out ${isFolded ? 'opacity-50 grayscale' : ''}`}
                  style={pos}
                >
                  {isMe ? (
                    <>
                      {/* 自己的底牌 (大尺寸，稍微倾斜) */}
                      {!isFolded && (
                        <div className="flex gap-1 mb-1 sm:mb-2 z-10 relative">
                           <div className="w-10 h-14 sm:w-14 sm:h-20 bg-white rounded-md shadow-xl border border-gray-300 flex items-center justify-center text-black font-bold text-lg transform -rotate-6">
                            K♠
                          </div>
                          <div className="w-10 h-14 sm:w-14 sm:h-20 bg-white rounded-md shadow-xl border border-gray-300 flex items-center justify-center text-black font-bold text-lg transform rotate-6">
                            Q♠
                          </div>
                        </div>
                      )}
                      <div className="relative z-20">
                        <div className="bg-blue-600/90 px-3 sm:px-4 py-1 rounded-full text-xs sm:text-sm font-bold border border-blue-400 shadow-lg whitespace-nowrap">
                          You ($2500)
                        </div>
                        {getMockPosition(playerCount, i) && (
                          <div className="absolute -right-2 -top-2 min-w-[20px] px-1 h-5 bg-white rounded-full border border-gray-300 flex items-center justify-center text-black text-[9px] font-bold shadow-sm z-30">
                            {getMockPosition(playerCount, i)}
                          </div>
                        )}
                      </div>
                    </>
                  ) : (
                    <>
                      {/* 头像和信息 */}
                      <div className="flex flex-col items-center z-20 relative">
                        <div className={`w-10 h-10 sm:w-12 sm:h-12 bg-gray-700 rounded-full border-2 ${isAllIn ? 'border-red-500 shadow-[0_0_15px_rgba(239,68,68,0.6)]' : 'border-gray-500'} shadow-lg flex items-center justify-center mb-1 relative`}>
                          <User className="w-5 h-5 sm:w-6 sm:h-6 text-gray-400" />
                          {getMockPosition(playerCount, i) && (
                            <div className="absolute -right-3 -bottom-2 min-w-[20px] px-1 h-5 bg-white rounded-full border border-gray-300 flex items-center justify-center text-black text-[9px] font-bold shadow-sm z-30">
                              {getMockPosition(playerCount, i)}
                            </div>
                          )}
                          {isFolded && (
                            <div className="absolute inset-0 bg-black/60 rounded-full flex items-center justify-center">
                               <span className="text-white text-[10px] font-bold transform -rotate-12">FOLD</span>
                            </div>
                          )}
                          {isAllIn && (
                            <div className="absolute -top-2 bg-red-600 text-white text-[9px] font-bold px-1.5 py-0.5 rounded shadow-sm animate-pulse whitespace-nowrap">
                              ALL IN
                            </div>
                          )}
                        </div>
                        <div className="bg-black/70 px-2 py-1 rounded text-[10px] sm:text-xs whitespace-nowrap border border-gray-700">
                          Player {i} ({isAllIn ? '$0' : '$1000'})
                        </div>
                      </div>

                      {/* 对手的底牌 (推入牌桌内圈) */}
                      {!isFolded && (
                        <div className="flex gap-0.5" style={getCardsPosition(playerCount, i)}>
                          {showCards ? (
                            // 摊牌可见
                            <>
                              <div className="w-6 h-8 sm:w-8 sm:h-11 bg-white rounded shadow border border-gray-300 flex items-center justify-center text-red-500 font-bold text-xs sm:text-sm transform -rotate-3">
                                A♥
                              </div>
                              <div className="w-6 h-8 sm:w-8 sm:h-11 bg-white rounded shadow border border-gray-300 flex items-center justify-center text-red-500 font-bold text-xs sm:text-sm transform rotate-3">
                                K♥
                              </div>
                            </>
                          ) : (
                            // 隐藏 (背面)
                            <>
                              <div className="w-6 h-8 sm:w-8 sm:h-11 bg-blue-800 rounded shadow border border-white/20 flex items-center justify-center transform -rotate-3 overflow-hidden">
                                <div className="w-full h-full border-[2px] border-blue-700 opacity-50 m-0.5 rounded-sm"></div>
                              </div>
                              <div className="w-6 h-8 sm:w-8 sm:h-11 bg-blue-800 rounded shadow border border-white/20 flex items-center justify-center transform rotate-3 overflow-hidden">
                                <div className="w-full h-full border-[2px] border-blue-700 opacity-50 m-0.5 rounded-sm"></div>
                              </div>
                            </>
                          )}
                        </div>
                      )}
                    </>
                  )}

                  {/* 玩家的下注筹码 (推入牌桌内圈) */}
                  {mockBet > 0 && !isFolded && (
                    <div style={getBetPosition(playerCount, i)} className={`flex items-center rounded-full px-2 py-0.5 border shadow-sm ${bgColorClass}`}>
                      <div className={`w-3 h-3 rounded-full border-2 border-dashed mr-1 shadow-inner ${chipColorClass}`}></div>
                      <span className={`text-[10px] sm:text-xs font-bold ${textColorClass}`}>
                        ${mockBet}
                      </span>
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        </main>
      </div>

      {/* 右侧/下方：侧边栏 (聊天与操作) */}
      <aside className="w-full lg:w-80 xl:w-96 bg-gray-800 border-t lg:border-t-0 lg:border-l border-gray-700 flex flex-col z-30 h-auto lg:h-screen">
        
        {/* 聊天/日志区域 (占位) */}
        <div className="flex-1 p-4 overflow-y-auto hidden lg:flex flex-col gap-3">
          <div className="flex items-center gap-2 text-gray-400 mb-2 sticky top-0 bg-gray-800 pb-2 z-10 border-b border-gray-700">
            <MessageSquare className="w-4 h-4" />
            <span className="text-sm font-semibold">Game Log & Chat</span>
          </div>
          
          {/* 系统/游戏日志 (弱化、斜体、无明显边框) */}
          <div className="text-xs text-gray-500 italic px-2">
            <span className="text-green-500 font-semibold not-italic">System:</span> Welcome to Room 123456
          </div>
          <div className="text-xs text-gray-500 italic px-2">
            <span className="text-blue-400 font-semibold not-italic">Player 1</span> sits down at seat 1
          </div>
          
          {/* 玩家聊天 (他人发送，左侧气泡) */}
          <div className="flex flex-col gap-1 mt-1">
            <span className="text-xs text-yellow-400 font-semibold px-2">Player 2</span>
            <div className="text-sm text-gray-200 bg-gray-700/80 p-2.5 rounded-2xl rounded-tl-sm self-start max-w-[90%] shadow-sm">
              Good luck everyone! Let's play.
            </div>
          </div>

          {/* 系统/游戏日志 */}
          <div className="text-xs text-gray-500 italic px-2 mt-1">
            <span className="text-yellow-400 font-semibold not-italic">Player 2</span> raises to $100
          </div>

          {/* 玩家聊天 (自己发送，右侧气泡) */}
          <div className="flex flex-col gap-1 mt-1 items-end">
            <span className="text-xs text-blue-400 font-semibold px-2">You</span>
            <div className="text-sm text-white bg-blue-600/90 p-2.5 rounded-2xl rounded-tr-sm self-end max-w-[90%] shadow-sm">
              Bring it on! 😎
            </div>
          </div>
        </div>

        {/* 操作面板 (Action Bar) */}
        <div className="p-4 sm:p-6 bg-gray-800 border-t border-gray-700 shadow-[0_-10px_30px_rgba(0,0,0,0.3)] lg:shadow-none flex flex-col gap-4">
          
          {/* 顶部下注控制区 */}
          <div className="bg-gray-900/60 rounded-xl p-4 border border-gray-700 flex flex-col gap-3">
            {/* 输入框与 BB 显示 */}
            <div className="flex items-center gap-3">
              <input 
                type="number" 
                value={betAmount}
                onChange={(e) => setBetAmount(Number(e.target.value))}
                className="flex-1 bg-gray-800 border border-gray-600 rounded-md px-3 py-2 text-white font-bold focus:outline-none focus:border-blue-500 transition-colors"
              />
              <span className="text-gray-400 text-sm w-16 whitespace-nowrap">
                ({(betAmount / bb).toFixed(1)} BB)
              </span>
            </div>

            {/* 范围提示 */}
            <div className="text-gray-400 text-xs">
              范围: {minBet} - {maxBet}
            </div>

            {/* 滑动条 */}
            <input 
              type="range" 
              min={minBet} 
              max={maxBet} 
              value={betAmount}
              onChange={(e) => setBetAmount(Number(e.target.value))}
              className="w-full accent-blue-500 h-2 bg-gray-700 rounded-lg appearance-none cursor-pointer"
            />

            {/* 快捷下注按钮 */}
            <div className="grid grid-cols-5 gap-2 mt-1">
              <button onClick={() => setBetAmount(minBet)} className="bg-gray-800 hover:bg-gray-700 text-gray-300 text-xs sm:text-sm py-2 rounded border border-gray-600 transition-colors">
                最小
              </button>
              <button onClick={() => setBetAmount(bb * 3)} className="bg-gray-800 hover:bg-gray-700 text-gray-300 text-xs sm:text-sm py-2 rounded border border-gray-600 transition-colors">
                3BB
              </button>
              <button onClick={() => setBetAmount(bb * 4)} className="bg-gray-800 hover:bg-gray-700 text-gray-300 text-xs sm:text-sm py-2 rounded border border-gray-600 transition-colors">
                4BB
              </button>
              <button onClick={() => setBetAmount(bb * 5)} className="bg-gray-800 hover:bg-gray-700 text-gray-300 text-xs sm:text-sm py-2 rounded border border-gray-600 transition-colors">
                5BB
              </button>
              <button onClick={() => setBetAmount(maxBet)} className="bg-red-900/20 hover:bg-red-900/40 text-red-400 text-xs sm:text-sm py-2 rounded border border-red-800/50 transition-colors">
                全押
              </button>
            </div>
          </div>

          {/* 底部动作按钮区 (2x2 网格) */}
          <div className="grid grid-cols-2 gap-2 sm:gap-3">
            <button className="bg-[#4b5563] hover:bg-[#374151] text-white font-bold py-2.5 sm:py-3 rounded-lg shadow-md transition-colors text-sm sm:text-base">
              弃牌
            </button>
            <button className="bg-[#16a34a] hover:bg-[#15803d] text-white font-bold py-2.5 sm:py-3 rounded-lg shadow-md transition-colors text-sm sm:text-base">
              跟注 20
            </button>
            <button className="bg-[#2563eb] hover:bg-[#1d4ed8] text-white font-bold py-2.5 sm:py-3 rounded-lg shadow-md transition-colors text-sm sm:text-base">
              下注 {betAmount}
            </button>
            <button className="bg-[#dc2626] hover:bg-[#b91c1c] text-white font-bold py-2.5 sm:py-3 rounded-lg shadow-md transition-colors text-sm sm:text-base">
              全下 {maxBet}
            </button>
          </div>
        </div>
      </aside>

    </div>
  )
}