interface BoardProps {
  gameState: string;
}

export function Board({ gameState }: BoardProps) {
  return (
    <>
      <div className="flex gap-1 sm:gap-2 z-10">
        {[1, 2, 3, 4, 5].map((i) => (
          <div key={i} className="w-10 h-14 sm:w-16 sm:h-24 bg-white rounded-md shadow-md border border-gray-300 flex items-center justify-center text-red-500 font-bold text-lg sm:text-xl">
            A♥
          </div>
        ))}
      </div>

      {gameState !== 'settling' && (
        <div className="absolute top-1/4 left-1/2 -translate-x-1/2 bg-black/60 px-4 py-1 rounded-full text-green-400 font-bold text-xs sm:text-sm border border-green-900/50 transition-opacity">
          底池: $1,500
        </div>
      )}
    </>
  )
}
