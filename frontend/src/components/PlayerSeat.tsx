import { User } from 'lucide-react'
import { getSeatPosition, getCardsPosition, getBetPosition, getMockPosition } from './tableUtils'

interface PlayerSeatProps {
  index: number;
  playerCount: number;
  gameState: string;
  currentPlayerIndex: number;
  countdown: number;
  betAmount: number;
}

export function PlayerSeat({ index, playerCount, gameState, currentPlayerIndex, countdown, betAmount }: PlayerSeatProps) {
  const pos = getSeatPosition(playerCount, index);
  const isMe = index === 0;

  let playerState = 'active';
  let lastAction = '';

  if (index === 0) {
    lastAction = 'raise';
  } else if (index === 1) {
    lastAction = 'call';
  } else if (index === 2 && playerCount > 2) {
    playerState = 'folded';
    lastAction = 'fold';
  } else if (index === 3 && playerCount > 3) {
    playerState = 'allin';
    lastAction = 'allin';
  } else if (index === 4 && playerCount > 4) {
    lastAction = 'bet';
  } else {
    lastAction = index % 2 === 0 ? 'check' : 'call';
  }

  const isFolded = playerState === 'folded';
  const isAllIn = playerState === 'allin';
  const isCurrentTurn = index === currentPlayerIndex && gameState === 'playing';
  const showCards = (isMe || index === 1) && !isFolded; 

  let mockBet = 0;
  if (!isFolded) {
    if (isMe) mockBet = betAmount;
    else if (isAllIn) mockBet = 1000;
    else if (lastAction === 'bet') mockBet = 200;
    else if (lastAction === 'raise') mockBet = 400;
    else if (lastAction === 'call') mockBet = 50;
  }

  let chipColorClass = 'bg-blue-500 border-blue-300';
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
    chipColorClass = 'bg-yellow-500 border-yellow-300';
    textColorClass = 'text-yellow-400';
    bgColorClass = 'bg-black/60 border-yellow-500/50';
  }

  return (
    <div
      className={`absolute flex flex-col items-center justify-center z-20 transition-all duration-500 ease-in-out ${isFolded ? 'opacity-50 grayscale' : ''}`}
      style={pos}
    >
      {isMe ? (
        <>
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
            <div className={`bg-blue-600/90 px-3 sm:px-4 py-1 rounded-full text-xs sm:text-sm font-bold border ${isCurrentTurn ? 'border-yellow-400 shadow-[0_0_15px_rgba(250,204,21,0.6)] animate-pulse' : 'border-blue-400 shadow-lg'} whitespace-nowrap transition-all duration-300`}>
              You ($2500)
            </div>
            {isCurrentTurn && (
              <div className="absolute -top-6 left-1/2 -translate-x-1/2 bg-yellow-500 text-black text-xs font-bold px-2 py-0.5 rounded-full shadow-lg animate-bounce">
                {countdown}s
              </div>
            )}
            {getMockPosition(playerCount, index) && (
              <div className="absolute -right-2 -top-2 min-w-[20px] px-1 h-5 bg-white rounded-full border border-gray-300 flex items-center justify-center text-black text-[9px] font-bold shadow-sm z-30">
                {getMockPosition(playerCount, index)}
              </div>
            )}
          </div>
        </>
      ) : (
        <>
          <div className="flex flex-col items-center z-20 relative">
            <div className={`w-10 h-10 sm:w-12 sm:h-12 bg-gray-700 rounded-full border-2 ${isAllIn ? 'border-red-500 shadow-[0_0_15px_rgba(239,68,68,0.6)]' : isCurrentTurn ? 'border-yellow-400 shadow-[0_0_20px_rgba(250,204,21,0.8)] animate-pulse' : 'border-gray-500'} shadow-lg flex items-center justify-center mb-1 relative transition-all duration-300`}>
              <User className="w-5 h-5 sm:w-6 sm:h-6 text-gray-400" />
              
              {isCurrentTurn && (
                <div className="absolute -top-6 bg-yellow-500 text-black text-xs font-bold px-2 py-0.5 rounded-full shadow-lg animate-bounce z-40">
                  {countdown}s
                </div>
              )}

              {getMockPosition(playerCount, index) && (
                <div className="absolute -right-3 -bottom-2 min-w-[20px] px-1 h-5 bg-white rounded-full border border-gray-300 flex items-center justify-center text-black text-[9px] font-bold shadow-sm z-30">
                  {getMockPosition(playerCount, index)}
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
              Player {index} ({isAllIn ? '$0' : '$1000'})
            </div>
          </div>

          {!isFolded && (
            <div className="flex gap-0.5" style={getCardsPosition(playerCount, index)}>
              {showCards ? (
                <>
                  <div className="w-6 h-8 sm:w-8 sm:h-11 bg-white rounded shadow border border-gray-300 flex items-center justify-center text-red-500 font-bold text-xs sm:text-sm transform -rotate-3">
                    A♥
                  </div>
                  <div className="w-6 h-8 sm:w-8 sm:h-11 bg-white rounded shadow border border-gray-300 flex items-center justify-center text-red-500 font-bold text-xs sm:text-sm transform rotate-3">
                    K♥
                  </div>
                </>
              ) : (
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

      {mockBet > 0 && !isFolded && (
        <div style={getBetPosition(playerCount, index)} className={`flex items-center rounded-full px-2 py-0.5 border shadow-sm ${bgColorClass}`}>
          <div className={`w-3 h-3 rounded-full border-2 border-dashed mr-1 shadow-inner ${chipColorClass}`}></div>
          <span className={`text-[10px] sm:text-xs font-bold ${textColorClass}`}>
            ${mockBet}
          </span>
        </div>
      )}
    </div>
  )
}
