import { useState } from 'react'

function App() {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gray-900 text-white">
      <h1 className="text-5xl font-bold text-green-500 mb-8">
        Texas Hold'em
      </h1>
      <div className="bg-gray-800 p-8 rounded-xl shadow-2xl border border-gray-700 max-w-md w-full">
        <p className="text-gray-400 mb-6 text-center">
          Ready to join the table?
        </p>
        <div className="flex flex-col gap-4">
          <button className="bg-green-600 hover:bg-green-500 text-white font-semibold py-3 px-6 rounded-lg transition-colors duration-200">
            Create Room
          </button>
          <button className="bg-blue-600 hover:bg-blue-500 text-white font-semibold py-3 px-6 rounded-lg transition-colors duration-200">
            Join Room
          </button>
        </div>
      </div>
    </div>
  )
}

export default App
