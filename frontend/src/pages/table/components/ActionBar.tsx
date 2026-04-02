interface ActionBarProps {
  gameState: string;
  isReady: boolean;
  setIsReady: (b: boolean) => void;
  canCheck: boolean;
  betAmount: number;
  setBetAmount: (n: number) => void;
  minBet: number;
  maxBet: number;
  bb: number;
}

export function ActionBar({ gameState, isReady, setIsReady, canCheck, betAmount, setBetAmount, minBet, maxBet, bb }: ActionBarProps) {
  return (
    <div className="p-4 sm:p-6 bg-gray-800 border-t border-gray-700 shadow-[0_-10px_30px_rgba(0,0,0,0.3)] lg:shadow-none flex flex-col gap-4">
      {gameState === 'playing' ? (
        <>
          <div className="bg-gray-900/60 rounded-xl p-4 border border-gray-700 flex flex-col gap-3">
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

            <div className="text-gray-400 text-xs">
              范围: {minBet} - {maxBet}
            </div>

            <input 
              type="range" 
              min={minBet} 
              max={maxBet} 
              value={betAmount}
              onChange={(e) => setBetAmount(Number(e.target.value))}
              className="w-full accent-blue-500 h-2 bg-gray-700 rounded-lg appearance-none cursor-pointer"
            />

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

          <div className="grid grid-cols-2 gap-2 sm:gap-3">
            {canCheck ? (
              <button className="bg-[#4b5563] hover:bg-[#374151] text-white font-bold py-2.5 sm:py-3 rounded-lg shadow-md transition-colors text-sm sm:text-base">
                过牌 (Check)
              </button>
            ) : (
              <button className="bg-[#4b5563] hover:bg-[#374151] text-white font-bold py-2.5 sm:py-3 rounded-lg shadow-md transition-colors text-sm sm:text-base">
                弃牌 (Fold)
              </button>
            )}
            
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
        </>
      ) : (
        <div className="flex flex-col items-center justify-center h-full min-h-[160px] gap-4">
          <p className="text-gray-400 text-sm">
            {gameState === 'waiting' ? '等待其他玩家准备...' : '对局结算中...'}
          </p>
          {isReady ? (
            <button 
              onClick={() => setIsReady(false)}
              className="w-full max-w-[200px] bg-gray-600 hover:bg-gray-500 text-white font-bold py-3 rounded-lg shadow-md transition-colors text-lg border border-gray-500"
            >
              取消准备
            </button>
          ) : (
            <button 
              onClick={() => setIsReady(true)}
              className="w-full max-w-[200px] bg-green-600 hover:bg-green-500 text-white font-bold py-3 rounded-lg shadow-[0_0_15px_rgba(22,163,74,0.5)] transition-all hover:shadow-[0_0_20px_rgba(22,163,74,0.8)] text-lg border border-green-400"
            >
              准备 (Ready)
            </button>
          )}
        </div>
      )}
    </div>
  )
}
