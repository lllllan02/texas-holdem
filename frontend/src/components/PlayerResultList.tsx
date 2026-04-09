import type { PlayerHandResult } from '../types/game'
import { PlayingCard } from './PlayingCard'

interface PlayerResultListProps {
  results: PlayerHandResult[];
  currentUserId?: string;
}

export function PlayerResultList({ results, currentUserId }: PlayerResultListProps) {
  const getHandRankName = (rank: number) => {
    if (rank < 0) return '';
    return ['高牌', '一对', '两对', '三条', '顺子', '同花', '葫芦', '四条', '同花顺', '皇家同花顺'][rank] || '未知牌型';
  };

  return (
    <div className="space-y-2">
      {/* 排序：赢家在前，然后按盈亏降序 */}
      {[...results].sort((a, b) => {
        if (a.is_winner && !b.is_winner) return -1;
        if (!a.is_winner && b.is_winner) return 1;
        return b.net_profit - a.net_profit;
      }).map((result, idx) => {
        const isMe = result.player_id === currentUserId;
        return (
          <div key={idx} className={`flex justify-between items-center p-2 rounded border ${result.is_winner ? 'border-green-700/50 bg-green-900/20' : 'border-gray-800 bg-gray-800/30'}`}>
            <div className="flex items-center gap-2">
              <span className={`text-sm ${isMe ? 'text-blue-400 font-bold' : 'text-white'}`}>
                {result.player_name} {isMe && '(You)'}
              </span>
            </div>
            <div className="flex items-center gap-3">
              {result.cards && result.cards.length > 0 ? (
                <div className="flex gap-1 mr-2">
                  {result.cards.map((c, i) => (
                    <PlayingCard key={`hole-${i}`} card={c} className="scale-[0.5] sm:scale-[0.6] origin-center" />
                  ))}
                </div>
              ) : (
                <span className="text-gray-500 text-xs italic mr-2">未亮牌</span>
              )}
              {result.hand_rank >= 0 && result.cards && result.cards.length > 0 && (
                <span className="text-gray-400 text-xs">
                  {getHandRankName(result.hand_rank)}
                </span>
              )}
              <span className={`${result.net_profit > 0 ? 'text-green-400' : result.net_profit < 0 ? 'text-red-400' : 'text-gray-500'} font-bold text-sm w-16 text-right`}>
                {result.net_profit > 0 ? '+' : ''}{result.net_profit}
              </span>
            </div>
          </div>
        );
      })}
    </div>
  );
}
