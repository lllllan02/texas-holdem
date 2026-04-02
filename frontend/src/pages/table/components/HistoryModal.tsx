import { useState } from 'react'
import { History, ChevronDown, ChevronUp } from 'lucide-react'
import { mockHistories } from '../mockData'

interface HistoryModalProps {
  show: boolean;
  onClose: () => void;
  userName: string;
}

export function HistoryModal({ show, onClose, userName }: HistoryModalProps) {
  const [expandedHistoryId, setExpandedHistoryId] = useState<number | null>(3)

  if (!show) return null;

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
          {mockHistories.map(history => {
            const isExpanded = expandedHistoryId === history.id;
            return (
              <div key={history.id} className="bg-gray-900/50 rounded-lg border border-gray-700 overflow-hidden">
                <div 
                  className="flex justify-between items-center p-3 cursor-pointer hover:bg-gray-800/80 transition-colors"
                  onClick={() => setExpandedHistoryId(isExpanded ? null : history.id)}
                >
                  <div className="flex items-center gap-3">
                    <span className="text-gray-400 text-sm font-mono w-16">Hand #{history.id}</span>
                    <div className="flex flex-col gap-1">
                      {history.details.filter(d => d.amountStr.startsWith('+')).slice(0, 1).map((winner, idx) => (
                        <div key={idx} className="flex items-center gap-2 text-xs">
                          <span className={`font-medium ${winner.isMe ? 'text-blue-400' : 'text-gray-300'}`}>
                            {winner.isMe ? userName : winner.name} {winner.isMe && '(You)'}
                          </span>
                          
                          {winner.cards.length > 0 && (
                            <>
                              <span className="text-gray-600">|</span>
                              <div className="flex gap-0.5">
                                {winner.cards.map((c, i) => (
                                  <div key={i} className={`w-4 h-6 bg-white rounded-sm shadow-sm border border-gray-300 flex items-center justify-center font-bold text-[9px] ${c.includes('♥') || c.includes('♦') ? 'text-red-500' : 'text-black'}`}>
                                    {c}
                                  </div>
                                ))}
                              </div>
                            </>
                          )}
                          
                          <span className="text-gray-600">|</span>
                          <span className="text-gray-400">{winner.handType}</span>
                          
                          <span className="text-gray-600">|</span>
                          <span className="text-green-400 font-bold">{winner.amountStr}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                  <div className="flex items-center gap-4">
                    <span className="text-gray-400 font-bold text-sm">Pot: ${history.pot}</span>
                    {isExpanded ? <ChevronUp className="w-4 h-4 text-gray-500" /> : <ChevronDown className="w-4 h-4 text-gray-500" />}
                  </div>
                </div>

                {isExpanded && (
                  <div className="p-4 border-t border-gray-800 bg-gray-900/30">
                    <div className="flex gap-2 mb-4 justify-center">
                      {history.board.map((card, idx) => (
                        card ? (
                          <div key={idx} className={`w-8 h-12 bg-white rounded shadow-sm border border-gray-300 flex items-center justify-center font-bold text-sm ${card.includes('♥') || card.includes('♦') ? 'text-red-500' : 'text-black'}`}>
                            {card}
                          </div>
                        ) : (
                          <div key={idx} className="w-8 h-12 bg-gray-800 rounded border border-gray-700 opacity-50"></div>
                        )
                      ))}
                    </div>

                    <div className="space-y-2">
                      {history.details.map((detail, idx) => (
                        <div key={idx} className={`flex justify-between items-center p-2 rounded border ${detail.borderClass} ${detail.bgClass}`}>
                          <div className="flex items-center gap-2">
                            <span className={`${detail.roleClass} text-xs font-bold w-12`}>{detail.role}</span>
                            <span className={`text-sm ${detail.isMe ? 'text-blue-400 font-bold' : 'text-white'}`}>
                              {detail.isMe ? userName : detail.name} {detail.isMe && '(You)'}
                            </span>
                          </div>
                          <div className="flex items-center gap-3">
                            {detail.cards.length > 0 && (
                              <div className="flex gap-1 mr-2">
                                {detail.cards.map((c, i) => (
                                  <div key={i} className={`w-5 h-7 bg-white rounded-sm shadow-sm border border-gray-300 flex items-center justify-center font-bold text-[10px] ${c.includes('♥') || c.includes('♦') ? 'text-red-500' : 'text-black'}`}>
                                    {c}
                                  </div>
                                ))}
                              </div>
                            )}
                            <span className={`text-gray-400 text-xs ${detail.italic ? 'italic' : ''}`}>{detail.handType}</span>
                            <span className={`${detail.amountClass} font-bold text-sm`}>{detail.amountStr}</span>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )
          })}
          
          <div className="text-center text-gray-500 text-sm py-4">
            没有更多记录了
          </div>
        </div>
      </div>
    </div>
  )
}
