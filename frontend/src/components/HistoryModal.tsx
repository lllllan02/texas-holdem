import { useState } from 'react'
import { History, ChevronDown, ChevronUp } from 'lucide-react'
import type { ShowdownSummary } from '../types/game'
import { PlayingCard } from './PlayingCard'
import { PlayerResultList } from './PlayerResultList'

interface HistoryModalProps {
  show: boolean;
  onClose: () => void;
  userId?: string;
  histories: ShowdownSummary[];
}

export function HistoryModal({ show, onClose, userId, histories }: HistoryModalProps) {
  // 默认展开最新的一局
  const [expandedHistoryId, setExpandedHistoryId] = useState<number | null>(
    histories.length > 0 ? histories[histories.length - 1].hand_id : null
  )

  if (!show) return null;

  const getHandRankName = (rank: number) => {
    if (rank < 0) return '';
    return ['高牌', '一对', '两对', '三条', '顺子', '同花', '葫芦', '四条', '同花顺', '皇家同花顺'][rank] || '未知牌型';
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
      <div className="bg-gray-800 border border-gray-700 rounded-xl shadow-2xl w-full max-w-2xl max-h-[80vh] flex flex-col">
        <div className="p-4 border-b border-gray-700 flex justify-between items-center">
          <h2 className="text-lg font-bold text-white flex items-center gap-2">
            <History className="w-5 h-5 text-blue-400" />
            历史对局记录
          </h2>
          <button 
            onClick={onClose}
            className="text-gray-400 hover:text-white transition"
          >
            ✕
          </button>
        </div>
        
        <div className="flex-1 overflow-y-auto p-4 flex flex-col gap-3">
          {histories.length === 0 ? (
            <div className="text-center text-gray-500 text-sm py-8">
              暂无对局记录
            </div>
          ) : (
            [...histories].reverse().map(history => {
              const isExpanded = expandedHistoryId === history.hand_id;
              
              // 找出赢家展示在折叠面板的摘要上
              const winners = history.player_results.filter(r => r.is_winner);
              
              return (
                <div key={history.hand_id} className="bg-gray-900/50 rounded-lg border border-gray-700 overflow-hidden">
                  <div 
                    className="flex justify-between items-center p-3 cursor-pointer hover:bg-gray-800/80 transition-colors"
                    onClick={() => setExpandedHistoryId(isExpanded ? null : history.hand_id)}
                  >
                    <div className="flex items-center gap-3">
                      <span className="text-gray-400 text-sm font-mono w-20">第 {history.hand_id} 局</span>
                      <div className="flex flex-col gap-1">
                        {winners.slice(0, 1).map((winner, idx) => {
                          const isMe = winner.player_id === userId;
                          return (
                            <div key={idx} className="flex items-center gap-2 text-xs">
                              <span className={`font-medium ${isMe ? 'text-blue-400' : 'text-gray-300'}`}>
                                {winner.player_name} {isMe && '(You)'}
                              </span>
                              
                              {winner.cards && winner.cards.length > 0 && (
                                <>
                                  <span className="text-gray-600">|</span>
                                  <div className="flex gap-0.5">
                                    {winner.cards.map((c, i) => (
                                      <PlayingCard key={i} card={c} className="scale-[0.4] origin-left -mr-4" />
                                    ))}
                                  </div>
                                </>
                              )}
                              
                              {winner.hand_rank > 0 && (
                                <>
                                  <span className="text-gray-600">|</span>
                                  <span className="text-gray-400">{getHandRankName(winner.hand_rank)}</span>
                                </>
                              )}
                              
                              <span className="text-gray-600">|</span>
                              <span className="text-green-400 font-bold">+{winner.net_profit}</span>
                            </div>
                          );
                        })}
                        {winners.length > 1 && (
                          <div className="text-xs text-gray-500">等 {winners.length} 人获胜...</div>
                        )}
                      </div>
                    </div>
                    <div className="flex items-center gap-4">
                      <span className="text-gray-400 font-bold text-sm">底池: ${history.total_pot}</span>
                      {isExpanded ? <ChevronUp className="w-4 h-4 text-gray-500" /> : <ChevronDown className="w-4 h-4 text-gray-500" />}
                    </div>
                  </div>

                  {isExpanded && (
                    <div className="p-4 border-t border-gray-800 bg-gray-900/30">
                      {history.board_cards && history.board_cards.length > 0 && (
                        <div className="flex gap-2 mb-4 justify-center">
                          {history.board_cards.map((card, idx) => (
                            <PlayingCard key={idx} card={card} className="scale-[0.8] sm:scale-100 origin-top" />
                          ))}
                        </div>
                      )}

                      <PlayerResultList results={history.player_results} currentUserId={userId} />
                    </div>
                  )}
                </div>
              );
            })
          )}
          
          {histories.length > 0 && (
            <div className="text-center text-gray-500 text-sm py-4">
              没有更多记录了
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
