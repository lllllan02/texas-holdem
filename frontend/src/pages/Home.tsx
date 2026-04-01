import { Link } from 'react-router-dom'

export default function Home() {
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
          
          <div className="mt-6 pt-6 border-t border-gray-700 text-center">
            <p className="text-sm text-gray-500 mb-4">UI Design Previews</p>
            <Link 
              to="/table" 
              className="text-green-400 hover:text-green-300 underline text-sm"
            >
              Preview Poker Table UI
            </Link>
          </div>
        </div>
      </div>
    </div>
  )
}
