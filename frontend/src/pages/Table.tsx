import { useState, useEffect, useRef } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'

import { Header } from '../components/Header'
import { SettingsModal } from '../components/SettingsModal'
import { useUser } from '../hooks/useUser'
import { useWebSocket } from '../hooks/useWebSocket'
import { deleteRoom, getRoom } from '../api/room'
import type { StateUpdateSnapshot } from '../types/game'
import { getSeatPosition } from '../components/tableUtils'

interface GameLog {
  id: string;
  time: string;
  text: string;
  type: 'system' | 'action' | 'chat';
  senderId?: string;
  senderName?: string;
  senderAvatar?: string;
}

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
  const [logs, setLogs] = useState<GameLog[]>([
    { id: 'init-1', time: new Date().toLocaleTimeString('zh-CN', { hour12: false }), text: '欢迎来到德州扑克房间', type: 'system' },
    { id: 'init-2', time: new Date().toLocaleTimeString('zh-CN', { hour12: false }), text: '等待玩家加入和准备...', type: 'system' }
  ])
  const logsEndRef = useRef<HTMLDivElement>(null)
  const prevGameStateRef = useRef<StateUpdateSnapshot | null>(null)

  useEffect(() => {
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [logs])
  
  const [chatInput, setChatInput] = useState('')
  
  const roomNumber = searchParams.get('room')
  
  // 初始化 WebSocket
  const { messageQueue, isKicked, error, sendMessage } = useWebSocket(validatedRoomNumber, user?.id)
  const processedMessageIndexRef = useRef(0)

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
    while (processedMessageIndexRef.current < messageQueue.length) {
      const lastMessage = messageQueue[processedMessageIndexRef.current];
      processedMessageIndexRef.current++;

      if (lastMessage.type === 'room.welcome') {
        // 判断当前用户是否是房主
        setIsOwner(lastMessage.payload.owner_id === user?.id);
        setLogs(prev => [...prev, {
          id: Date.now() + '-' + Math.random(),
          time: new Date().toLocaleTimeString('zh-CN', { hour12: false }),
          text: `欢迎来到房间: ${lastMessage.payload.room_number}`,
          type: 'system' as const
        }].slice(-100));
      } else if (lastMessage.type === 'room.destroyed') {
        if (!isOwner) {
          try {
            sessionStorage.setItem('home_notice', '房间已被房主解散');
          } catch {
            // ignore
          }
        }
        navigate('/');
      } else if (lastMessage.type === 'room.chat') {
        const chatPayload = lastMessage.payload;
        setLogs(prev => [...prev, {
          id: Date.now() + '-' + Math.random(),
          time: new Date().toLocaleTimeString('zh-CN', { hour12: false }),
          text: chatPayload.message,
          type: 'chat' as const,
          senderId: chatPayload.user_id,
          senderName: chatPayload.user_name,
          senderAvatar: chatPayload.avatar
        }].slice(-100));
      } else if (lastMessage.type === 'room.player_join') {
        const p = lastMessage.payload;
        const isMe = p.user_id === user?.id;
        setLogs(prev => [...prev, {
          id: Date.now() + '-' + Math.random(),
          time: new Date().toLocaleTimeString('zh-CN', { hour12: false }),
          text: isMe ? '你进入了房间' : `玩家 [${p.user_name}] 进入了房间`,
          type: 'system' as const
        }].slice(-100));
        // 如果是自己加入房间，此时可能还没有 gameState，尝试从 payload 中获取或等待 state_update
      } else if (lastMessage.type === 'room.player_leave') {
        const p = lastMessage.payload;
        setLogs(prev => [...prev, {
          id: Date.now() + '-' + Math.random(),
          time: new Date().toLocaleTimeString('zh-CN', { hour12: false }),
          text: `玩家 [${p.user_name}] 离开了房间`,
          type: 'system' as const
        }].slice(-100));
      } else if (lastMessage.type === 'texas.state_update') {
        const snap = lastMessage.payload as StateUpdateSnapshot;
        const prevSnap = prevGameStateRef.current;
        
        // 只有在第一次收到 state_update，或者状态确实有变化时才更新
        setGameState(snap);

        const reason = lastMessage.reason;
        if (reason) {
          let logText = '';
          
          const getNewPlayerName = () => {
            const newP = snap.players?.find(p => !prevSnap?.players?.find(old => old.id === p.id));
            return newP ? newP.name : '某玩家';
          };
          
          const getRemovedPlayerName = () => {
            const oldP = prevSnap?.players?.find(old => !snap.players?.find(p => p.id === old.id));
            return oldP ? oldP.name : '某玩家';
          };

          const getReadyPlayerName = (isReady: boolean) => {
            const targetState = isReady ? 'ready' : 'waiting';
            const p = snap.players?.find(p => {
              const oldP = prevSnap?.players?.find(old => old.id === p.id);
              return p.state === targetState && oldP?.state !== targetState;
            });
            return p ? p.name : '某玩家';
          };

          const getOfflineChangedPlayerName = (toOffline: boolean) => {
            const p = snap.players?.find(p => {
              const oldP = prevSnap?.players?.find(old => old.id === p.id);
              return p.is_offline === toOffline && oldP?.is_offline !== toOffline;
            });
            return p ? p.name : '某玩家';
          };

          switch (reason) {
            case 'player_joined': 
              // 忽略 player_joined 的日志，因为 room.player_join 已经处理了
              break;
            case 'player_left': 
              // 忽略 player_left 的日志，因为 room.player_leave 已经处理了
              break;
            case 'player_reconnected': logText = `${getOfflineChangedPlayerName(false)} 重新连接`; break;
            case 'sit_down': logText = `${getNewPlayerName()} 坐下了`; break;
            case 'stand_up': logText = `${getRemovedPlayerName()} 站起了`; break;
            case 'ready': logText = `${getReadyPlayerName(true)} 已准备`; break;
            case 'cancel_ready': logText = `${getReadyPlayerName(false)} 取消了准备`; break;
            case 'deal_hole_cards': logText = `第 ${snap.hand_count + 1} 局游戏开始，正在发牌...`; break;
            case 'player_action': 
              if (snap.last_action) {
                const p = snap.players?.find(p => p.id === snap.last_action?.player_id);
                const pName = p ? p.name : '某玩家';
                const actionMap: Record<string, string> = {
                  'fold': '弃牌', 'check': '过牌', 'call': '跟注', 'bet': '下注', 'raise': '加注', 'allin': 'All-in'
                };
                const actName = actionMap[snap.last_action.action || ''] || snap.last_action.action;
                logText = `${pName} ${actName} ${snap.last_action.amount ? snap.last_action.amount : ''}`;
              }
              break;
            case 'next_stage': 
              const stageMap: Record<string, string> = {
                'PREFLOP': '翻牌前', 'FLOP': '翻牌圈', 'TURN': '转牌圈', 'RIVER': '河牌圈', 'SHOWDOWN': '摊牌'
              };
              logText = `进入 ${stageMap[snap.stage] || snap.stage} 阶段`; 
              break;
            case 'showdown': logText = '进入摊牌结算'; break;
            case 'hand_finished': logText = `第 ${snap.hand_count} 局游戏结束`; break;
            case 'early_finish': logText = `其他玩家均已弃牌，本局提前结束`; break;
          }
          
          if (logText) {
            setLogs(prev => [...prev, {
              id: Date.now() + '-' + Math.random(),
              time: new Date().toLocaleTimeString('zh-CN', { hour12: false }),
              text: logText,
              type: 'action' as const
            }].slice(-100));
          }
        }
        
        prevGameStateRef.current = snap;
      }
    }
  }, [messageQueue, user?.id, navigate, isOwner]);

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

  const handleSendChat = () => {
    if (!chatInput.trim()) return;
    sendMessage('room.chat', { message: chatInput.trim() });
    setChatInput('');
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
                {Array.from({ length: gameState.max_players || 9 }).map((_, index) => {
                  const player = gameState.players?.find(p => p.seat_number === index);
                  const isMySeat = player?.id === user?.id;
                  const position = getSeatPosition(gameState.max_players || 9, index);

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
      <aside className="w-full lg:w-80 xl:w-96 bg-gray-800 border-t lg:border-t-0 lg:border-l border-gray-700 flex flex-col z-30 h-[40vh] lg:h-screen">
        {gameState ? (
          <>
            {/* 上方：游戏日志和聊天内容 (约占 2/3) */}
            <div className="flex-1 flex flex-col bg-gray-900 mx-4 mt-4 mb-2 rounded-lg shadow-inner overflow-hidden">
              <div className="bg-gray-800 px-4 py-2 border-b border-gray-700 font-bold text-sm text-gray-300 flex justify-between items-center">
                <span>游戏记录</span>
              </div>
              
              {/* 日志列表 */}
              <div className="flex-1 overflow-y-auto p-4 space-y-3">
                {logs.map(log => {
                  if (log.type === 'system' || log.type === 'action') {
                    return (
                      <div key={log.id} className="flex justify-center">
                        <span className="text-[11px] bg-gray-800/80 text-gray-400 px-3 py-1 rounded-full shadow-sm">
                          [{log.time}] {log.text}
                        </span>
                      </div>
                    );
                  }

                  // 聊天气泡
                  const isMe = log.senderId === user?.id;
                  return (
                    <div key={log.id} className={`flex w-full gap-2 ${isMe ? 'flex-row-reverse' : 'flex-row'} items-start`}>
                      {/* 头像 */}
                      <div className="w-8 h-8 rounded-full bg-gray-700 flex-shrink-0 flex items-center justify-center overflow-hidden border border-gray-600 shadow-sm mt-1">
                        {log.senderAvatar ? (
                          <img src={log.senderAvatar} alt={log.senderName} className="w-full h-full object-cover" />
                        ) : (
                          <span className="text-xs text-gray-400 font-bold">{log.senderName?.[0]?.toUpperCase()}</span>
                        )}
                      </div>

                      {/* 消息内容区 */}
                      <div className={`flex flex-col ${isMe ? 'items-end' : 'items-start'} max-w-[75%]`}>
                        <span className="text-[10px] text-gray-500 mb-1 mx-1">
                          {log.senderName} {log.time}
                        </span>
                        <div className={`px-3 py-2 rounded-2xl text-sm shadow-md ${
                          isMe 
                            ? 'bg-blue-600 text-white rounded-tr-sm' 
                            : 'bg-gray-700 text-gray-200 rounded-tl-sm'
                        }`}>
                          {log.text}
                        </div>
                      </div>
                    </div>
                  );
                })}
                <div ref={logsEndRef} />
              </div>

              {/* 聊天输入框 */}
              <div className="p-2 bg-gray-800 border-t border-gray-700 flex">
                <input 
                  type="text" 
                  placeholder="输入聊天内容..." 
                  value={chatInput}
                  onChange={e => setChatInput(e.target.value)}
                  onKeyDown={e => e.key === 'Enter' && handleSendChat()}
                  className="flex-1 bg-gray-900 text-white text-sm px-3 py-1.5 rounded-l-md outline-none border border-gray-700 focus:border-gray-500 transition-colors" 
                />
                <button 
                  onClick={handleSendChat}
                  disabled={!chatInput.trim()}
                  className="bg-blue-600 hover:bg-blue-500 px-4 py-1.5 rounded-r-md text-sm text-white font-bold disabled:opacity-50 transition-colors" 
                >
                  发送
                </button>
              </div>
            </div>

            {/* 下方：操作区域 (约占 1/3) */}
            <div className="shrink-0 bg-gray-900 mx-4 mt-2 mb-4 p-4 rounded-lg shadow-inner min-h-[160px] flex flex-col justify-center">
              {(() => {
                const myPlayer = gameState.players?.find(p => p.id === user?.id);
                
                if (!myPlayer) {
                  return <div className="text-sm text-gray-400 text-center">请先在牌桌上找个空位坐下</div>;
                }

                if (gameState.stage === 'WAITING') {
                  if (myPlayer.state === 'ready') {
                    return (
                      <button 
                        onClick={handleCancelReady}
                        className="w-full py-3 bg-red-600 hover:bg-red-500 text-white font-bold rounded-lg transition-colors shadow-lg"
                      >
                        取消准备
                      </button>
                    );
                  } else {
                    return (
                      <button 
                        onClick={handleReady}
                        className="w-full py-3 bg-green-600 hover:bg-green-500 text-white font-bold rounded-lg transition-colors shadow-lg"
                      >
                        准备
                      </button>
                    );
                  }
                }

                // 游戏进行中的操作按钮占位
                return (
                  <div className="grid grid-cols-2 gap-3">
                    <button className="py-3 bg-gray-700 text-white font-bold rounded-lg opacity-50 cursor-not-allowed shadow-md">弃牌 (Fold)</button>
                    <button className="py-3 bg-gray-700 text-white font-bold rounded-lg opacity-50 cursor-not-allowed shadow-md">过牌 (Check)</button>
                    <button className="py-3 bg-gray-700 text-white font-bold rounded-lg opacity-50 cursor-not-allowed shadow-md">跟注 (Call)</button>
                    <button className="py-3 bg-blue-600 text-white font-bold rounded-lg opacity-50 cursor-not-allowed shadow-md">加注 (Raise)</button>
                  </div>
                );
              })()}
            </div>
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center text-gray-500">
            等待游戏数据...
          </div>
        )}
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
