import { mockSettlementData } from './mockData'

interface SettlementOverlayProps {
  gameState: string;
  hideSettlement: boolean;
  setHideSettlement: (b: boolean) => void;
  userName: string;
}

export function SettlementOverlay({ gameState, hideSettlement, setHideSettlement, userName }: SettlementOverlayProps) {
  if (gameState !== 'settling') return null;

  return (
    <>
      {!hideSettlement && (
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 z-50 flex flex-col items-center animate-in fade-in zoom-in duration-500">
          <div className="bg-black/80 backdrop-blur-md px-6 py-4 rounded-2xl border-2 border-yellow-500/50 shadow-[0_0_30px_rgba(234,179,8,0.3)] flex flex-col gap-3 relative min-w-[280px]">
            <button 
              onClick={() => setHideSettlement(true)}
              className="absolute top-2 right-3 text-gray-400 hover:text-white transition"
              title="隐藏结算面板"
            >
              ✕
            </button>
            <div className="text-yellow-400 font-black text-xl text-center border-b border-gray-700/50 pb-2">
              本局结算
            </div>
            
            <div className="flex flex-col gap-2">
              {mockSettlementData.map((item, idx) => (
                <div key={idx} className="flex flex-col bg-gray-900/60 p-2.5 rounded-lg border border-gray-700/50">
                  <div className="text-xs text-gray-400 mb-1.5">{item.title}</div>
                  <div className="flex justify-between items-center">
                    <div className="flex items-center gap-2">
                      <span className={`text-sm font-bold ${item.isMe ? 'text-blue-400' : 'text-white'}`}>
                        {item.isMe ? userName : item.winner} {item.isMe && '(You)'}
                      </span>
                      <span className="text-xs text-gray-300 bg-gray-800 px-1.5 py-0.5 rounded border border-gray-700">{item.hand}</span>
                    </div>
                    <div className="text-green-400 font-bold text-base">
                      {item.amount}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {hideSettlement && (
        <button 
          onClick={() => setHideSettlement(false)}
          className="absolute top-24 left-1/2 -translate-x-1/2 z-40 bg-yellow-600/90 hover:bg-yellow-500 text-white text-xs font-bold px-4 py-1.5 rounded-full shadow-lg backdrop-blur-sm transition-all border border-yellow-400 animate-in fade-in slide-in-from-top-4"
        >
          查看结算
        </button>
      )}
    </>
  )
}
