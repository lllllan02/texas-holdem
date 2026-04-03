import { useState, useEffect } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'

import { Header } from '../components/Header'
import { SettingsModal } from '../components/SettingsModal'
import { useUser } from '../hooks/useUser'
import { useWebSocket } from '../hooks/useWebSocket'
import { deleteRoom, getRoom } from '../api/room'
import type { StateUpdateSnapshot } from '../types/game'
import { getSeatPosition } from '../components/tableUtils'

export default function Table() {
  const { user, loading, updateUserInfo } = useUser()
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const [showSettings, setShowSettings] = useState(false)
  const [isOwner, setIsOwner] = useState(false)
  const [validatedRoomNumber, setValidatedRoomNumber] = useState<string | null>(null)
  const [isValidatingRoom, setIsValidatingRoom] = useState(false)
  const [gameState, setGameState] = useState<StateUpdateSnapshot | null>(null)
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  
  const roomNumber = searchParams.get('room')
  
  // 初始化 WebSocket
  const { lastMessage, isKicked, error, sendMessage } = useWebSocket(validatedRoomNumber, user?.id)

  useEffect(() => {
    if (!roomNumber) {
      navigate('/')
      return
    }

    setIsValidatingRoom(true)
    setValidatedRoomNumber(null)
    getRoom(roomNumber)
      .then(() => {
        setValidatedRoomNumber(roomNumber)
      })
      .catch((err: any) => {
        try {
          sessionStorage.setItem('home_notice', err?.message || '房间不存在或已被回收')
        } catch {
          // ignore
        }
        navigate('/')
      })
      .finally(() => {
        setIsValidatingRoom(false)
      })
  }, [roomNumber, navigate])

  // 监听 WebSocket 消息
  useEffect(() => {
    if (!lastMessage) return;

    if (lastMessage.type === 'room.welcome') {
      // 判断当前用户是否是房主
      setIsOwner(lastMessage.payload.owner_id === user?.id);
    } else if (lastMessage.type === 'room.destroyed') {
      if (!isOwner) {
        try {
          sessionStorage.setItem('home_notice', '房间已被房主解散');
        } catch {
          // ignore
        }
      }
      navigate('/');
    } else if (lastMessage.type === 'texas.state_update') {
      setGameState(lastMessage.payload as StateUpdateSnapshot);
    }
  }, [lastMessage, user?.id, navigate, isOwner]);

  useEffect(() => {
    if (!isKicked) return;
    try {
      sessionStorage.setItem('kick_notice', error || '该账号已在其他设备登录，你已下线');
    } catch {
      // ignore
    }
    navigate('/');
  }, [isKicked, error, navigate]);

  const handleDeleteRoom = () => {
    setShowDeleteConfirm(true);
  };

  const confirmDeleteRoom = async () => {
    if (!roomNumber || !user?.id) return;
    setIsDeleting(true);
    try {
      await deleteRoom(roomNumber, user.id);
      setShowDeleteConfirm(false);
      // 解散成功后，后端会广播 room.destroyed 消息，前端收到后会自动跳转回大厅
    } catch (err: any) {
      alert(err.message || '解散房间失败');
    } finally {
      setIsDeleting(false);
    }
  };

  const handleSitDown = (seatNumber: number) => {
    sendMessage('texas.sit_down', { seat_number: seatNumber });
  };

  const handleStandUp = () => {
    sendMessage('texas.stand_up', {});
  };

  const handleReady = () => {
    sendMessage('texas.ready', {});
  };

  const handleCancelReady = () => {
    sendMessage('texas.cancel', {});
  };

  if (loading) {
    return <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center">加载用户信息中...</div>
  }
  if (!validatedRoomNumber || isValidatingRoom) {
    return <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center">正在进入房间...</div>
  }

  return (
    <div className="min-h-screen bg-gray-900 text-white flex flex-col lg:flex-row overflow-hidden">
      
      {/* 左侧/上方：游戏主区域 */}
      <div className="flex-1 relative flex flex-col min-h-[60vh] lg:min-h-screen">
        
        <Header 
          userName={user?.nickname || '加载中...'} 
          userAvatar={user?.avatar}
          roomNumber={roomNumber || '未知'}
          isOwner={isOwner}
          onOpenHistory={() => {}} 
          onOpenSettings={() => setShowSettings(true)} 
          onDeleteRoom={handleDeleteRoom}
        />

        {/* 牌桌区域 */}
        <main className="flex-1 flex items-center justify-center p-4 sm:p-8 mt-16 lg:mt-0 relative">
          {/* 牌桌 (绿色椭圆) */}
          <div className="relative w-full max-w-2xl lg:max-w-4xl aspect-[2/1] bg-green-800 rounded-[1000px] border-[12px] sm:border-[16px] border-gray-800 shadow-2xl flex items-center justify-center">
            
            {/* 牌桌内圈线 */}
            <div className="absolute inset-3 sm:inset-4 rounded-[1000px] border-2 border-green-700 opacity-50 pointer-events-none"></div>

            {!gameState ? (
              <div className="text-gray-400 font-mono text-lg animate-pulse">
                等待游戏数据...
              </div>
            ) : (
              <>
                {/* 座位渲染 */}
                {Array.from({ length: gameState.max_players }).map((_, index) => {
                  const player = gameState.players?.find(p => p.seat_number === index);
                  const isMySeat = player?.id === user?.id;
                  const position = getSeatPosition(gameState.max_players, index);

                  return (
                    <div 
                      key={`seat-${index}`}
                      className="absolute"
                      style={{ ...position, zIndex: 50 }}
                    >
                      {player ? (
                        <div className="flex flex-col items-center group relative">
                          {/* 玩家头像和信息 */}
                          <div className={`w-14 h-14 sm:w-16 sm:h-16 rounded-full border-4 ${isMySeat ? 'border-yellow-400' : 'border-gray-600'} bg-gray-800 flex items-center justify-center overflow-hidden shadow-lg relative`}>
                            {player.avatar ? (
                              <img src={player.avatar} alt={player.name} className="w-full h-full object-cover" />
                            ) : (
                              <span className="text-xl font-bold text-gray-400">{player.name[0]?.toUpperCase()}</span>
                            )}
                            
                            {/* 离线遮罩 */}
                            {player.is_offline && (
                              <div className="absolute inset-0 bg-black/60 flex items-center justify-center">
                                <span className="text-xs text-red-400 font-bold">离线</span>
                              </div>
                            )}
                          </div>
                          
                          {/* 玩家名字和筹码 */}
                          <div className="mt-1 bg-black/70 px-2 py-1 rounded-md text-center min-w-[80px]">
                            <div className="text-xs text-gray-300 truncate max-w-[80px]">{player.name}</div>
                            <div className="text-sm font-bold text-yellow-400">${player.chips}</div>
                          </div>

                          {/* 状态标签 */}
                          {player.state !== 'waiting' && player.state !== 'active' && (
                            <div className="absolute -top-3 left-1/2 -translate-x-1/2 bg-blue-600 text-white text-[10px] px-2 py-0.5 rounded-full whitespace-nowrap shadow-md">
                              {player.state === 'ready' ? '已准备' : player.state === 'folded' ? '已弃牌' : player.state === 'allin' ? 'All-in' : player.state}
                            </div>
                          )}

                          {/* 起身按钮 (仅对自己显示) */}
                          {isMySeat && gameState.stage === 'WAITING' && (
                            <button 
                              onClick={handleStandUp}
                              className="absolute -right-8 top-0 bg-red-600 hover:bg-red-500 text-white text-xs px-2 py-1 rounded-md shadow-md transition-colors opacity-0 group-hover:opacity-100"
                            >
                              起身
                            </button>
                          )}
                        </div>
                      ) : (
                        <div className="flex flex-col items-center">
                          {/* 空座位 */}
                          <button 
                            onClick={() => handleSitDown(index)}
                            disabled={gameState.stage !== 'WAITING'} // 只要游戏未开始（WAITING阶段），允许任何人（包括已落座玩家换座）坐下
                            className="w-14 h-14 sm:w-16 sm:h-16 rounded-full border-4 border-dashed border-gray-600 bg-black/30 hover:bg-black/50 hover:border-gray-400 flex items-center justify-center transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                          >
                            <span className="text-gray-500 text-sm">坐下</span>
                          </button>
                        </div>
                      )}
                    </div>
                  );
                })}
              </>
            )}
          </div>
        </main>
      </div>

      {/* 右侧/下方：侧边栏 (聊天与操作) */}
      <aside className="w-full lg:w-80 xl:w-96 bg-gray-800 border-t lg:border-t-0 lg:border-l border-gray-700 flex flex-col z-30 h-auto lg:h-screen p-4 items-center justify-center text-gray-500">
        {gameState && (
          <div className="w-full space-y-4">
            {/* 操作区域 */}
            <div className="bg-gray-900 p-4 rounded-lg shadow-inner">
              <h3 className="text-lg font-bold text-white mb-4">操作栏</h3>
              
              {(() => {
                const myPlayer = gameState.players?.find(p => p.id === user?.id);
                
                if (!myPlayer) {
                  return <div className="text-sm text-gray-400">请先在牌桌上找个空位坐下</div>;
                }

                if (gameState.stage === 'WAITING') {
                  if (myPlayer.state === 'ready') {
                    return (
                      <button 
                        onClick={handleCancelReady}
                        className="w-full py-3 bg-red-600 hover:bg-red-500 text-white font-bold rounded-lg transition-colors"
                      >
                        取消准备
                      </button>
                    );
                  } else {
                    return (
                      <button 
                        onClick={handleReady}
                        className="w-full py-3 bg-green-600 hover:bg-green-500 text-white font-bold rounded-lg transition-colors"
                      >
                        准备
                      </button>
                    );
                  }
                }

                return <div className="text-sm text-gray-400">游戏进行中...</div>;
              })()}
            </div>
            
            {/* 聊天占位符 */}
            <div className="bg-gray-900 p-4 rounded-lg shadow-inner flex-1 min-h-[200px] flex items-center justify-center">
              聊天区域占位符
            </div>
          </div>
        )}
        {!gameState && <div>操作栏与聊天占位符</div>}
      </aside>

      {/* 解散房间确认弹窗 */}
      {showDeleteConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 px-4">
          <div className="bg-gray-800 rounded-xl shadow-2xl p-6 w-full max-w-sm border border-gray-700">
            <h3 className="text-xl font-bold text-white mb-4">解散房间</h3>
            <p className="text-gray-300 mb-6">确定要解散房间吗？此操作不可恢复，所有玩家将被移出房间。</p>
            <div className="flex justify-end space-x-3">
              <button 
                onClick={() => setShowDeleteConfirm(false)}
                disabled={isDeleting}
                className="px-4 py-2 rounded-lg bg-gray-700 text-white hover:bg-gray-600 transition-colors disabled:opacity-50"
              >
                取消
              </button>
              <button 
                onClick={confirmDeleteRoom}
                disabled={isDeleting}
                className="px-4 py-2 rounded-lg bg-red-600 text-white hover:bg-red-500 transition-colors disabled:opacity-50 flex items-center"
              >
                {isDeleting ? '解散中...' : '确定解散'}
              </button>
            </div>
          </div>
        </div>
      )}

      <SettingsModal 
        show={showSettings} 
        onClose={() => setShowSettings(false)} 
        userName={user?.nickname || ''} 
        userAvatar={user?.avatar}
        setUserInfo={async (name, avatar) => {
          try {
            await updateUserInfo(name, avatar);
          } catch (e) {
            console.error('Failed to update user info', e);
            alert('修改失败');
          }
        }} 
      />

    </div>
  )
}
