interface MockControlPanelProps {
  playerCount: number;
  setPlayerCount: (n: number) => void;
  gameState: 'waiting' | 'playing' | 'settling';
  setGameState: (s: 'waiting' | 'playing' | 'settling') => void;
  setHideSettlement: (b: boolean) => void;
}

export function MockControlPanel({ playerCount, setPlayerCount, gameState, setGameState, setHideSettlement }: MockControlPanelProps) {
  return (
    <div className="absolute top-16 right-4 sm:right-8 z-20 bg-gray-800 p-4 rounded-xl border border-gray-700 shadow-xl w-48 flex flex-col gap-3">
      <div>
        <label className="block text-sm text-green-400 font-semibold mb-2">
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
      </div>
      
      <div className="border-t border-gray-700 pt-2">
        <label className="block text-xs text-gray-400 mb-2">Game State (Mock)</label>
        <select 
          value={gameState}
          onChange={(e) => {
            setGameState(e.target.value as any);
            if (e.target.value === 'settling') {
              setHideSettlement(false);
            }
          }}
          className="w-full bg-gray-900 text-xs text-white border border-gray-600 rounded p-1"
        >
          <option value="waiting">Waiting</option>
          <option value="playing">Playing</option>
          <option value="settling">Settling</option>
        </select>
      </div>
    </div>
  )
}
