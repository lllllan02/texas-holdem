export function GameInfoPanel() {
  return (
    <div className="absolute top-16 left-4 sm:left-8 z-20 bg-gray-800/80 backdrop-blur-sm p-3 rounded-xl border border-gray-700 shadow-xl text-xs text-gray-300 flex flex-col gap-1.5 min-w-[120px]">
      <div className="flex justify-between gap-4">
        <span className="text-gray-500">盲注 (SB/BB)</span>
        <span className="font-mono text-white">10 / 20</span>
      </div>
      <div className="flex justify-between gap-4">
        <span className="text-gray-500">前注 (Ante)</span>
        <span className="font-mono text-white">0</span>
      </div>
      <div className="flex justify-between gap-4">
        <span className="text-gray-500">买入范围</span>
        <span className="font-mono text-white">400 - 2000</span>
      </div>
    </div>
  )
}
