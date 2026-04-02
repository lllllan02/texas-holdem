import { useState } from 'react'

import { Header } from './table/components/Header'
import { MockControlPanel } from './table/components/MockControlPanel'
import { GameInfoPanel } from './table/components/GameInfoPanel'
import { Board } from './table/components/Board'
import { SettlementOverlay } from './table/components/SettlementOverlay'
import { PlayerSeat } from './table/components/PlayerSeat'
import { ChatLog } from './table/components/ChatLog'
import { ActionBar } from './table/components/ActionBar'
import { HistoryModal } from './table/components/HistoryModal'
import { SettingsModal } from './table/components/SettingsModal'

// 这是一个纯 UI 设计页面，没有接入后端数据
export default function Table() {
  const [playerCount, setPlayerCount] = useState(9)
  const [betAmount, setBetAmount] = useState(60)
  const [showHistory, setShowHistory] = useState(false)
  const [showSettings, setShowSettings] = useState(false)
  const [hideSettlement, setHideSettlement] = useState(false)
  
  // 游戏核心状态模拟
  const [gameState, setGameState] = useState<'waiting' | 'playing' | 'settling'>('playing')
  const [isReady, setIsReady] = useState(false)
  const [currentPlayerIndex] = useState(4) // 模拟轮到玩家4
  const [countdown] = useState(15) // 模拟倒计时 15 秒
  const [canCheck] = useState(false) // 模拟当前是否可以过牌 (Check)
  
  // 模拟用户信息状态
  const [userName, setUserName] = useState('MyNickname')

  // 模拟数据
  const minBet = 60;
  const maxBet = 2336;
  const bb = 20;

  return (
    <div className="min-h-screen bg-gray-900 text-white flex flex-col lg:flex-row overflow-hidden">
      
      {/* 左侧/上方：游戏主区域 */}
      <div className="flex-1 relative flex flex-col min-h-[60vh] lg:min-h-screen">
        
        <Header 
          userName={userName} 
          onOpenHistory={() => setShowHistory(true)} 
          onOpenSettings={() => setShowSettings(true)} 
        />

        <MockControlPanel 
          playerCount={playerCount} 
          setPlayerCount={setPlayerCount} 
          gameState={gameState} 
          setGameState={setGameState} 
          setHideSettlement={setHideSettlement} 
        />

        <GameInfoPanel />

        {/* 牌桌区域 */}
        <main className="flex-1 flex items-center justify-center p-4 sm:p-8 mt-16 lg:mt-0 relative">
          {/* 牌桌 (绿色椭圆) */}
          <div className="relative w-full max-w-2xl lg:max-w-4xl aspect-[2/1] bg-green-800 rounded-[1000px] border-[12px] sm:border-[16px] border-gray-800 shadow-2xl flex items-center justify-center">
            
            {/* 牌桌内圈线 */}
            <div className="absolute inset-3 sm:inset-4 rounded-[1000px] border-2 border-green-700 opacity-50 pointer-events-none"></div>

            <Board gameState={gameState} />

            <SettlementOverlay 
              gameState={gameState} 
              hideSettlement={hideSettlement} 
              setHideSettlement={setHideSettlement} 
              userName={userName} 
            />

            {/* 动态渲染玩家座位 */}
            {Array.from({ length: playerCount }).map((_, i) => (
              <PlayerSeat 
                key={i} 
                index={i} 
                playerCount={playerCount} 
                gameState={gameState} 
                currentPlayerIndex={currentPlayerIndex} 
                countdown={countdown} 
                betAmount={betAmount} 
              />
            ))}
          </div>
        </main>
      </div>

      {/* 右侧/下方：侧边栏 (聊天与操作) */}
      <aside className="w-full lg:w-80 xl:w-96 bg-gray-800 border-t lg:border-t-0 lg:border-l border-gray-700 flex flex-col z-30 h-auto lg:h-screen">
        <ChatLog />
        <ActionBar 
          gameState={gameState} 
          isReady={isReady} 
          setIsReady={setIsReady} 
          canCheck={canCheck} 
          betAmount={betAmount} 
          setBetAmount={setBetAmount} 
          minBet={minBet} 
          maxBet={maxBet} 
          bb={bb} 
        />
      </aside>

      <HistoryModal 
        show={showHistory} 
        onClose={() => setShowHistory(false)} 
        userName={userName} 
      />

      <SettingsModal 
        show={showSettings} 
        onClose={() => setShowSettings(false)} 
        userName={userName} 
        setUserName={setUserName} 
      />

    </div>
  )
}
